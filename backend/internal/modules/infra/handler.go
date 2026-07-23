package infra

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"gorm.io/gorm"
)

type Handler struct{ db *gorm.DB }

func NewHandler(db *gorm.DB) *Handler { return &Handler{db: db} }

func tenantID(c *gin.Context) uint64 { return c.GetUint64("tenant_id") }

func queryID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	return id
}

func page(c *gin.Context) (int, int) {
	pageNo, _ := strconv.Atoi(c.DefaultQuery("pageNo", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	return pageNo, pageSize
}

// ConfigPage godoc
// @Summary Page system parameters
// @Tags Infra Config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/config/page [get]
func (h *Handler) ConfigPage(c *gin.Context) {
	query := h.db.Model(&Config{}).Where("tenant_id = ?", tenantID(c))
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if key := strings.TrimSpace(c.Query("key")); key != "" {
		query = query.Where("`key` LIKE ?", "%"+key+"%")
	}
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []Config
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// ConfigGet godoc
// @Summary Get a system parameter
// @Tags Infra Config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/config/get [get]
func (h *Handler) ConfigGet(c *gin.Context) {
	var row Config
	if err := h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error; err != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "参数配置不存在")
		return
	}
	httpx.OK(c, row)
}

// ConfigValue godoc
// @Summary Get a public parameter value by key
// @Tags Infra Config
// @Produce json
// @Param key query string true "Config key"
// @Success 200 {object} httpx.Response
// @Router /infra/config/get-value-by-key [get]
func (h *Handler) ConfigValue(c *gin.Context) {
	var row Config
	if err := h.db.Where("tenant_id = ? AND `key` = ? AND visible = ?", tenantID(c), c.Query("key"), true).First(&row).Error; err != nil {
		httpx.OK(c, nil)
		return
	}
	httpx.OK(c, row.Value)
}

// ConfigCreate godoc
// @Summary Create a system parameter
// @Tags Infra Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ConfigSaveRequest true "Config"
// @Success 200 {object} httpx.Response
// @Router /infra/config/create [post]
func (h *Handler) ConfigCreate(c *gin.Context) {
	var req ConfigSaveRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	row := Config{TenantID: tenantID(c)}
	applyConfig(&row, req)
	if err := h.db.Create(&row).Error; err != nil {
		httpx.Fail(c, http.StatusConflict, 409, "参数键已存在")
		return
	}
	httpx.OK(c, row.ID)
}

// ConfigUpdate godoc
// @Summary Update a system parameter
// @Tags Infra Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ConfigSaveRequest true "Config"
// @Success 200 {object} httpx.Response
// @Router /infra/config/update [put]
func (h *Handler) ConfigUpdate(c *gin.Context) {
	var req ConfigSaveRequest
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	var row Config
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "参数配置不存在")
		return
	}
	applyConfig(&row, req)
	h.db.Save(&row)
	httpx.OK(c, true)
}

// ConfigDelete godoc
// @Summary Delete a system parameter
// @Tags Infra Config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/config/delete [delete]
func (h *Handler) ConfigDelete(c *gin.Context) {
	h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).Delete(&Config{})
	httpx.OK(c, true)
}

// ConfigDeleteList godoc
// @Summary Delete system parameters in batch
// @Tags Infra Config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/config/delete-list [delete]
func (h *Handler) ConfigDeleteList(c *gin.Context) {
	h.db.Where("tenant_id = ? AND id IN ?", tenantID(c), strings.Split(c.Query("ids"), ",")).Delete(&Config{})
	httpx.OK(c, true)
}

// FileConfigPage godoc
// @Summary Page file storage configurations
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file-config/page [get]
func (h *Handler) FileConfigPage(c *gin.Context) {
	query := h.db.Model(&FileConfig{}).Where("tenant_id = ?", tenantID(c))
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []FileConfig
	query.Order("master DESC,id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// FileConfigGet godoc
// @Summary Get a file storage configuration
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file-config/get [get]
func (h *Handler) FileConfigGet(c *gin.Context) {
	var row FileConfig
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "文件配置不存在")
		return
	}
	httpx.OK(c, row)
}

// FileConfigCreate godoc
// @Summary Create a file storage configuration
// @Tags Infra File
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body FileConfigSaveRequest true "File config"
// @Success 200 {object} httpx.Response
// @Router /infra/file-config/create [post]
func (h *Handler) FileConfigCreate(c *gin.Context) {
	var req FileConfigSaveRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	row := FileConfig{TenantID: tenantID(c)}
	applyFileConfig(&row, req)
	h.db.Create(&row)
	httpx.OK(c, row.ID)
}

// FileConfigUpdate godoc
// @Summary Update a file storage configuration
// @Tags Infra File
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body FileConfigSaveRequest true "File config"
// @Success 200 {object} httpx.Response
// @Router /infra/file-config/update [put]
func (h *Handler) FileConfigUpdate(c *gin.Context) {
	var req FileConfigSaveRequest
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	var row FileConfig
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "文件配置不存在")
		return
	}
	applyFileConfig(&row, req)
	h.db.Save(&row)
	httpx.OK(c, true)
}

// FileConfigDelete godoc
// @Summary Delete a file storage configuration
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file-config/delete [delete]
func (h *Handler) FileConfigDelete(c *gin.Context) {
	h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).Delete(&FileConfig{})
	httpx.OK(c, true)
}

// FileConfigDeleteList godoc
// @Summary Delete file storage configurations in batch
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file-config/delete-list [delete]
func (h *Handler) FileConfigDeleteList(c *gin.Context) {
	h.db.Where("tenant_id = ? AND id IN ?", tenantID(c), strings.Split(c.Query("ids"), ",")).Delete(&FileConfig{})
	httpx.OK(c, true)
}

// FileConfigTest godoc
// @Summary Validate a file storage configuration
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file-config/test [get]
func (h *Handler) FileConfigTest(c *gin.Context) {
	var row FileConfig
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "文件配置不存在")
		return
	}
	httpx.OK(c, row.Config != "")
}

// FileConfigMaster godoc
// @Summary Set the primary file storage configuration
// @Tags Infra File
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/file-config/update-master [put]
func (h *Handler) FileConfigMaster(c *gin.Context) {
	id := queryID(c)
	tx := h.db.Begin()
	tx.Model(&FileConfig{}).Where("tenant_id = ?", tenantID(c)).Update("master", false)
	tx.Model(&FileConfig{}).Where("tenant_id = ? AND id = ?", tenantID(c), id).Update("master", true)
	tx.Commit()
	httpx.OK(c, true)
}

// AccessLogPage godoc
// @Summary Page API access logs
// @Tags Infra Logging
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /infra/api-access-log/page [get]
func (h *Handler) AccessLogPage(c *gin.Context) {
	query := h.accessLogQuery(c)
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []APIAccessLog
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	views := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		resultCode := 0
		resultMessage := ""
		if row.Status >= http.StatusBadRequest {
			resultCode = row.Status
			resultMessage = http.StatusText(row.Status)
		}
		views = append(views, gin.H{
			"id": row.ID, "traceId": row.TraceID, "userId": row.UserID, "userType": 2,
			"applicationName": "nimbus-admin-api", "requestMethod": row.Method, "requestUrl": row.Path,
			"requestParams": "", "responseBody": "", "userIp": row.IP, "userAgent": row.UserAgent,
			"operateModule": "", "operateName": "", "operateType": 0,
			"beginTime": row.CreatedAt, "endTime": row.CreatedAt.Add(time.Duration(row.Duration) * time.Millisecond),
			"duration": row.Duration, "resultCode": resultCode, "resultMsg": resultMessage, "createTime": row.CreatedAt,
		})
	}
	httpx.OK(c, gin.H{"list": views, "total": total})
}

func (h *Handler) accessLogQuery(c *gin.Context) *gorm.DB {
	query := h.db.Model(&APIAccessLog{}).Where("tenant_id = ?", tenantID(c))
	if userID := strings.TrimSpace(c.Query("userId")); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if requestURL := strings.TrimSpace(c.Query("requestUrl")); requestURL != "" {
		query = query.Where("path LIKE ?", "%"+requestURL+"%")
	}
	if application := strings.TrimSpace(c.Query("applicationName")); application != "" && !strings.Contains(strings.ToLower("nimbus-admin-api"), strings.ToLower(application)) {
		query = query.Where("1 = 0")
	}
	if duration := strings.TrimSpace(c.Query("duration")); duration != "" {
		query = query.Where("duration >= ?", duration)
	}
	if resultCode := strings.TrimSpace(c.Query("resultCode")); resultCode != "" {
		if resultCode == "0" {
			query = query.Where("status < 400")
		} else {
			query = query.Where("status >= 400")
		}
	}
	if begin, end := c.Query("beginTime[0]"), c.Query("beginTime[1]"); begin != "" && end != "" {
		query = query.Where("created_at BETWEEN ? AND ?", begin, end)
	}
	return query
}

func applyConfig(row *Config, req ConfigSaveRequest) {
	row.Category, row.Type, row.Name, row.Key = req.Category, req.Type, req.Name, req.Key
	row.Value, row.Visible, row.Remark = req.Value, req.Visible, req.Remark
}

func applyFileConfig(row *FileConfig, req FileConfigSaveRequest) {
	row.Name, row.Storage, row.Master, row.Config, row.Remark = req.Name, req.Storage, req.Master, req.Config, req.Remark
}
