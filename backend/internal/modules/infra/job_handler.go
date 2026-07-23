package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/excelx"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"gorm.io/gorm"
)

type jobScheduler struct {
	db   *gorm.DB
	mu   sync.Mutex
	last map[uint64]string
}

func newJobScheduler(db *gorm.DB) *jobScheduler {
	return &jobScheduler{db: db, last: map[uint64]string{}}
}
func (s *jobScheduler) start() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for now := range ticker.C {
			var jobs []Job
			s.db.Where("status = 0").Find(&jobs)
			for _, job := range jobs {
				if !cronMatches(job.CronExpression, now) {
					continue
				}
				key := now.Format("20060102150405")
				s.mu.Lock()
				if s.last[job.ID] == key {
					s.mu.Unlock()
					continue
				}
				s.last[job.ID] = key
				s.mu.Unlock()
				go executeJob(s.db, job)
			}
		}
	}()
}

func cronFields(expression string) ([]string, bool) {
	fields := strings.Fields(expression)
	if len(fields) == 5 {
		return append([]string{"0"}, fields...), true
	}
	if len(fields) == 6 {
		return fields, true
	}
	if len(fields) == 7 {
		return fields[:6], true
	}
	return nil, false
}
func cronValueMatches(field string, value, min, max int) bool {
	if field == "*" || field == "?" {
		return true
	}
	for _, part := range strings.Split(field, ",") {
		step := 1
		if values := strings.SplitN(part, "/", 2); len(values) == 2 {
			part = values[0]
			parsed, err := strconv.Atoi(values[1])
			if err != nil || parsed < 1 {
				return false
			}
			step = parsed
		}
		start, end := min, max
		if part != "*" && part != "?" {
			if rangeValues := strings.SplitN(part, "-", 2); len(rangeValues) == 2 {
				var err error
				start, err = strconv.Atoi(rangeValues[0])
				if err != nil {
					return false
				}
				end, err = strconv.Atoi(rangeValues[1])
				if err != nil {
					return false
				}
			} else {
				parsed, err := strconv.Atoi(part)
				if err != nil {
					return false
				}
				start, end = parsed, parsed
			}
		}
		if value >= start && value <= end && (value-start)%step == 0 {
			return true
		}
	}
	return false
}
func cronMatches(expression string, at time.Time) bool {
	fields, ok := cronFields(expression)
	if !ok {
		return false
	}
	values := []int{at.Second(), at.Minute(), at.Hour(), at.Day(), int(at.Month()), int(at.Weekday())}
	limits := [][2]int{{0, 59}, {0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
	for i, value := range values {
		if !cronValueMatches(fields[i], value, limits[i][0], limits[i][1]) {
			if i == 5 && value == 0 && cronValueMatches(fields[i], 7, 0, 7) {
				continue
			}
			return false
		}
	}
	return true
}
func nextCronTimes(expression string, from time.Time, count int) ([]time.Time, error) {
	fields, ok := cronFields(expression)
	if !ok {
		return nil, fmt.Errorf("CRON 表达式必须是 5、6 或 7 段")
	}
	second := 0
	if fields[0] != "*" && fields[0] != "?" && !strings.ContainsAny(fields[0], ",-/") {
		parsed, err := strconv.Atoi(fields[0])
		if err != nil || parsed < 0 || parsed > 59 {
			return nil, fmt.Errorf("秒字段不合法")
		}
		second = parsed
	}
	candidate := from.Truncate(time.Minute).Add(time.Minute).Add(time.Duration(second) * time.Second)
	result := []time.Time{}
	deadline := from.AddDate(5, 0, 0)
	for candidate.Before(deadline) && len(result) < count {
		if cronMatches(expression, candidate) {
			result = append(result, candidate)
		}
		candidate = candidate.Add(time.Minute)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("未来五年内没有匹配的执行时间")
	}
	return result, nil
}

type httpJobParam struct {
	URL    string `json:"url"`
	Method string `json:"method"`
	Body   string `json:"body"`
}

func runJobHandler(job Job) (string, error) {
	switch strings.ToLower(strings.TrimSpace(job.HandlerName)) {
	case "noop", "health":
		return "OK", nil
	case "http":
		param := httpJobParam{URL: strings.TrimSpace(job.HandlerParam), Method: http.MethodGet}
		if strings.HasPrefix(strings.TrimSpace(job.HandlerParam), "{") {
			_ = json.Unmarshal([]byte(job.HandlerParam), &param)
		}
		if param.URL == "" {
			return "", fmt.Errorf("HTTP 任务缺少 url")
		}
		if param.Method == "" {
			param.Method = http.MethodGet
		}
		request, err := http.NewRequest(strings.ToUpper(param.Method), param.URL, bytes.NewBufferString(param.Body))
		if err != nil {
			return "", err
		}
		request.Header.Set("Content-Type", "application/json")
		client := http.Client{Timeout: 30 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			return "", err
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return string(body), fmt.Errorf("HTTP %d", response.StatusCode)
		}
		return string(body), nil
	default:
		return "", fmt.Errorf("未注册任务处理器 %q，可用处理器: noop、health、http", job.HandlerName)
	}
}
func executeJob(db *gorm.DB, job Job) {
	begin := time.Now()
	result, err := runJobHandler(job)
	attempt := 0
	for err != nil && attempt < job.RetryCount {
		attempt++
		if job.RetryInterval > 0 {
			time.Sleep(time.Duration(job.RetryInterval) * time.Millisecond)
		}
		result, err = runJobHandler(job)
	}
	end := time.Now()
	status := 0
	if err != nil {
		status = 1
		if result != "" {
			result += "\n"
		}
		result += err.Error()
	}
	db.Create(&JobLog{TenantID: job.TenantID, JobID: job.ID, HandlerName: job.HandlerName, HandlerParam: job.HandlerParam, CronExpression: job.CronExpression, ExecuteIndex: attempt, BeginTime: begin, EndTime: end, Duration: end.Sub(begin).Milliseconds(), Status: status, Result: result})
}

// JobPage godoc
// @Summary Page scheduled jobs
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job/page [get]
func (h *Handler) JobPage(c *gin.Context) {
	query := h.db.Model(&Job{}).Where("tenant_id = ?", tenantID(c))
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if handler := strings.TrimSpace(c.Query("handlerName")); handler != "" {
		query = query.Where("handler_name LIKE ?", "%"+handler+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	pn, ps := page(c)
	var rows []Job
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// JobGet godoc
// @Summary Get a scheduled job
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job/get [get]
func (h *Handler) JobGet(c *gin.Context) {
	var row Job
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "定时任务不存在")
		return
	}
	httpx.OK(c, row)
}
func validateJob(row Job) error {
	if strings.TrimSpace(row.Name) == "" || strings.TrimSpace(row.HandlerName) == "" {
		return fmt.Errorf("任务名称和处理器不能为空")
	}
	_, err := nextCronTimes(row.CronExpression, time.Now(), 1)
	return err
}

// JobCreate godoc
// @Summary Create a scheduled job
// @Tags Infra Job
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Job true "Job"
// @Success 200 {object} httpx.Response
// @Router /infra/job/create [post]
func (h *Handler) JobCreate(c *gin.Context) {
	var row Job
	if c.ShouldBindJSON(&row) != nil {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantID(c)
	if err := validateJob(row); err != nil {
		httpx.Fail(c, 400, 400, err.Error())
		return
	}
	if h.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "创建任务失败")
		return
	}
	httpx.OK(c, row.ID)
}

// JobUpdate godoc
// @Summary Update a scheduled job
// @Tags Infra Job
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Job true "Job"
// @Success 200 {object} httpx.Response
// @Router /infra/job/update [put]
func (h *Handler) JobUpdate(c *gin.Context) {
	var req Job
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	if err := validateJob(req); err != nil {
		httpx.Fail(c, 400, 400, err.Error())
		return
	}
	var row Job
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "定时任务不存在")
		return
	}
	row.Name, row.HandlerParam, row.CronExpression, row.RetryCount, row.RetryInterval, row.MonitorTimeout = req.Name, req.HandlerParam, req.CronExpression, req.RetryCount, req.RetryInterval, req.MonitorTimeout
	h.db.Save(&row)
	httpx.OK(c, true)
}

// JobDelete godoc
// @Summary Delete a scheduled job
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job/delete [delete]
func (h *Handler) JobDelete(c *gin.Context) { h.deleteJobs(c, []uint64{queryID(c)}) }

// JobDeleteList godoc
// @Summary Delete scheduled jobs in batch
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job/delete-list [delete]
func (h *Handler) JobDeleteList(c *gin.Context) { h.deleteJobs(c, parseIDs(c.Query("ids"))) }
func (h *Handler) deleteJobs(c *gin.Context, ids []uint64) {
	tx := h.db.Begin()
	tx.Where("tenant_id = ? AND job_id IN ?", tenantID(c), ids).Delete(&JobLog{})
	tx.Where("tenant_id = ? AND id IN ?", tenantID(c), ids).Delete(&Job{})
	tx.Commit()
	httpx.OK(c, true)
}

// JobStatus godoc
// @Summary Update scheduled job status
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job/update-status [put]
func (h *Handler) JobStatus(c *gin.Context) {
	result := h.db.Model(&Job{}).Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).Update("status", c.Query("status"))
	if result.RowsAffected == 0 {
		httpx.Fail(c, 404, 404, "定时任务不存在")
		return
	}
	httpx.OK(c, true)
}

// JobTrigger godoc
// @Summary Trigger a scheduled job immediately
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job/trigger [put]
func (h *Handler) JobTrigger(c *gin.Context) {
	var row Job
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "定时任务不存在")
		return
	}
	go executeJob(h.db, row)
	httpx.OK(c, true)
}

// JobNextTimes godoc
// @Summary Get the next ten execution times
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job/get_next_times [get]
func (h *Handler) JobNextTimes(c *gin.Context) {
	var row Job
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "定时任务不存在")
		return
	}
	times, err := nextCronTimes(row.CronExpression, time.Now(), 10)
	if err != nil {
		httpx.Fail(c, 400, 400, err.Error())
		return
	}
	httpx.OK(c, times)
}

// JobSync godoc
// @Summary Synchronize enabled jobs with the in-process scheduler
// @Description The scheduler reads the database dynamically; this endpoint validates all enabled CRON expressions.
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job/sync [post]
func (h *Handler) JobSync(c *gin.Context) {
	var rows []Job
	h.db.Where("tenant_id = ? AND status = 0", tenantID(c)).Find(&rows)
	for _, row := range rows {
		if err := validateJob(row); err != nil {
			httpx.Fail(c, 400, 400, fmt.Sprintf("任务 %s: %v", row.Name, err))
			return
		}
	}
	httpx.OK(c, true)
}

func (h *Handler) jobLogQuery(c *gin.Context) *gorm.DB {
	query := h.db.Model(&JobLog{}).Where("tenant_id = ?", tenantID(c))
	if job := c.Query("jobId"); job != "" {
		query = query.Where("job_id = ?", job)
	}
	if handler := c.Query("handlerName"); handler != "" {
		query = query.Where("handler_name LIKE ?", "%"+handler+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	return query
}

// JobLogPage godoc
// @Summary Page scheduled job execution logs
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job-log/page [get]
func (h *Handler) JobLogPage(c *gin.Context) {
	query := h.jobLogQuery(c)
	var total int64
	query.Count(&total)
	pn, ps := page(c)
	var rows []JobLog
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// JobLogGet godoc
// @Summary Get a scheduled job execution log
// @Tags Infra Job
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/job-log/get [get]
func (h *Handler) JobLogGet(c *gin.Context) {
	var row JobLog
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "任务日志不存在")
		return
	}
	httpx.OK(c, row)
}

// JobExport godoc
// @Summary Export scheduled jobs
// @Tags Infra Job
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /infra/job/export-excel [get]
func (h *Handler) JobExport(c *gin.Context) {
	var rows []Job
	h.db.Where("tenant_id = ?", tenantID(c)).Order("id").Find(&rows)
	book := newInfraBook("定时任务", []any{"编号", "名称", "状态", "处理器", "参数", "CRON", "重试次数", "重试间隔", "超时"})
	for i, row := range rows {
		infraWriteRow(book, "定时任务", i+2, []any{row.ID, row.Name, row.Status, row.HandlerName, row.HandlerParam, row.CronExpression, row.RetryCount, row.RetryInterval, row.MonitorTimeout})
	}
	excelx.Write(c, book, "定时任务.xlsx")
}

// JobLogExport godoc
// @Summary Export scheduled job logs
// @Tags Infra Job
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /infra/job-log/export-excel [get]
func (h *Handler) JobLogExport(c *gin.Context) {
	var rows []JobLog
	h.jobLogQuery(c).Order("id DESC").Limit(10000).Find(&rows)
	book := newInfraBook("任务日志", []any{"编号", "任务编号", "处理器", "开始时间", "结束时间", "耗时(ms)", "状态", "结果"})
	for i, row := range rows {
		infraWriteRow(book, "任务日志", i+2, []any{row.ID, row.JobID, row.HandlerName, row.BeginTime.Format(time.DateTime), row.EndTime.Format(time.DateTime), row.Duration, row.Status, row.Result})
	}
	excelx.Write(c, book, "任务日志.xlsx")
}
