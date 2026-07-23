package infra

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lohasle/nimbus-framework-go/internal/platform/excelx"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func newInfraBook(sheet string, headers []any) *excelize.File {
	book := excelize.NewFile()
	book.SetSheetName("Sheet1", sheet)
	infraWriteRow(book, sheet, 1, headers)
	return book
}
func infraWriteRow(book *excelize.File, sheet string, row int, values []any) {
	for column, value := range values {
		cell, _ := excelize.CoordinatesToCellName(column+1, row)
		_ = book.SetCellValue(sheet, cell, value)
	}
}

func (h *Handler) errorLogQuery(c *gin.Context) *gorm.DB {
	query := h.db.Model(&APIErrorLog{}).Where("tenant_id = ?", tenantID(c))
	if user := c.Query("userId"); user != "" {
		query = query.Where("user_id = ?", user)
	}
	if urlValue := strings.TrimSpace(c.Query("requestUrl")); urlValue != "" {
		query = query.Where("request_url LIKE ?", "%"+urlValue+"%")
	}
	if name := strings.TrimSpace(c.Query("exceptionName")); name != "" {
		query = query.Where("exception_name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("processStatus"); status != "" {
		query = query.Where("process_status = ?", status)
	}
	return query
}

// APIErrorLogPage godoc
// @Summary Page API error logs
// @Tags Infra Logging
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/api-error-log/page [get]
func (h *Handler) APIErrorLogPage(c *gin.Context) {
	query := h.errorLogQuery(c)
	var total int64
	query.Count(&total)
	pn, ps := page(c)
	var rows []APIErrorLog
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// APIErrorLogStatus godoc
// @Summary Update API error processing status
// @Tags Infra Logging
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/api-error-log/update-status [put]
func (h *Handler) APIErrorLogStatus(c *gin.Context) {
	id, status := queryID(c), c.Query("processStatus")
	if id == 0 || status == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	now := time.Now()
	result := h.db.Model(&APIErrorLog{}).Where("tenant_id = ? AND id = ?", tenantID(c), id).Updates(map[string]any{"process_status": status, "process_user_id": c.GetUint64("user_id"), "process_time": now})
	if result.RowsAffected == 0 {
		httpx.Fail(c, 404, 404, "错误日志不存在")
		return
	}
	httpx.OK(c, true)
}

// APIErrorLogExport godoc
// @Summary Export API error logs
// @Tags Infra Logging
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /infra/api-error-log/export-excel [get]
func (h *Handler) APIErrorLogExport(c *gin.Context) {
	var rows []APIErrorLog
	h.errorLogQuery(c).Order("id DESC").Limit(10000).Find(&rows)
	book := newInfraBook("API错误日志", []any{"编号", "Trace ID", "应用", "请求方法", "请求地址", "结果码", "异常名称", "异常消息", "处理状态", "异常时间"})
	for i, row := range rows {
		infraWriteRow(book, "API错误日志", i+2, []any{row.ID, row.TraceID, row.ApplicationName, row.RequestMethod, row.RequestURL, row.ResultCode, row.ExceptionName, row.ExceptionMessage, row.ProcessStatus, row.ExceptionTime.Format(time.DateTime)})
	}
	excelx.Write(c, book, "API错误日志.xlsx")
}

func mysqlDSNFromConfig(row DataSourceConfig) (string, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(row.URL), "jdbc:")
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "mysql" {
		return "", fmt.Errorf("仅支持 mysql:// 或 jdbc:mysql:// 数据源")
	}
	database := strings.TrimPrefix(parsed.Path, "/")
	if parsed.Host == "" || database == "" {
		return "", fmt.Errorf("数据源 URL 缺少主机或数据库名")
	}
	query := parsed.Query()
	if query.Get("parseTime") == "" {
		query.Set("parseTime", "true")
	}
	if query.Get("charset") == "" {
		query.Set("charset", "utf8mb4")
	}
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", row.Username, row.Password, parsed.Host, database, query.Encode()), nil
}
func verifyDataSource(row DataSourceConfig) error {
	dsn, err := mysqlDSNFromConfig(row)
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

// DataSourceList godoc
// @Summary List data source configurations
// @Tags Infra Data Source
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/data-source-config/list [get]
func (h *Handler) DataSourceList(c *gin.Context) {
	var rows []DataSourceConfig
	h.db.Where("tenant_id = ?", tenantID(c)).Order("id").Find(&rows)
	httpx.OK(c, rows)
}

// DataSourceGet godoc
// @Summary Get a data source configuration
// @Tags Infra Data Source
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/data-source-config/get [get]
func (h *Handler) DataSourceGet(c *gin.Context) {
	var row DataSourceConfig
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "数据源不存在")
		return
	}
	httpx.OK(c, row)
}

// DataSourceCreate godoc
// @Summary Create and verify a MySQL data source
// @Tags Infra Data Source
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DataSourceConfig true "Data source"
// @Success 200 {object} httpx.Response
// @Router /infra/data-source-config/create [post]
func (h *Handler) DataSourceCreate(c *gin.Context) {
	var row DataSourceConfig
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Name) == "" || strings.TrimSpace(row.URL) == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantID(c)
	if err := verifyDataSource(row); err != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "数据源连接失败: "+err.Error())
		return
	}
	if h.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "创建数据源失败")
		return
	}
	httpx.OK(c, row.ID)
}

// DataSourceUpdate godoc
// @Summary Update and verify a MySQL data source
// @Tags Infra Data Source
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DataSourceConfig true "Data source"
// @Success 200 {object} httpx.Response
// @Router /infra/data-source-config/update [put]
func (h *Handler) DataSourceUpdate(c *gin.Context) {
	var req DataSourceConfig
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row DataSourceConfig
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "数据源不存在")
		return
	}
	if err := verifyDataSource(req); err != nil {
		httpx.Fail(c, 400, 400, "数据源连接失败: "+err.Error())
		return
	}
	row.Name, row.URL, row.Username, row.Password = req.Name, req.URL, req.Username, req.Password
	h.db.Save(&row)
	httpx.OK(c, true)
}

// DataSourceDelete godoc
// @Summary Delete a data source
// @Tags Infra Data Source
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/data-source-config/delete [delete]
func (h *Handler) DataSourceDelete(c *gin.Context) { h.deleteDataSources(c, []uint64{queryID(c)}) }

// DataSourceDeleteList godoc
// @Summary Delete data sources in batch
// @Tags Infra Data Source
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/data-source-config/delete-list [delete]
func (h *Handler) DataSourceDeleteList(c *gin.Context) {
	h.deleteDataSources(c, parseIDs(c.Query("ids")))
}
func (h *Handler) deleteDataSources(c *gin.Context, ids []uint64) {
	h.db.Where("tenant_id = ? AND id IN ?", tenantID(c), ids).Delete(&DataSourceConfig{})
	httpx.OK(c, true)
}
