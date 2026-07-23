package system

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/excelx"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func (h *Handler) loginLogQuery(c *gin.Context) *gorm.DB {
	query := h.service.db.Model(&LoginLog{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if username := strings.TrimSpace(c.Query("username")); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if ip := strings.TrimSpace(c.Query("userIp")); ip != "" {
		query = query.Where("user_ip LIKE ?", "%"+ip+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	return query
}

// LoginLogPage godoc
// @Summary Page login logs
// @Tags System Audit
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/login-log/page [get]
func (h *Handler) LoginLogPage(c *gin.Context) {
	query := h.loginLogQuery(c)
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []LoginLog
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

func (h *Handler) operateLogQuery(c *gin.Context) *gorm.DB {
	query := h.service.db.Model(&OperateLog{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if user := strings.TrimSpace(c.Query("userName")); user != "" {
		query = query.Where("user_name LIKE ?", "%"+user+"%")
	}
	if typ := strings.TrimSpace(c.Query("type")); typ != "" {
		query = query.Where("type = ?", typ)
	}
	if sub := strings.TrimSpace(c.Query("subType")); sub != "" {
		query = query.Where("sub_type = ?", sub)
	}
	return query
}

// OperateLogPage godoc
// @Summary Page operation logs
// @Tags System Audit
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/operate-log/page [get]
func (h *Handler) OperateLogPage(c *gin.Context) {
	query := h.operateLogQuery(c)
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []OperateLog
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

func newSystemBook(sheet string, headers []any) *excelize.File {
	book := excelize.NewFile()
	book.SetSheetName("Sheet1", sheet)
	systemWriteRow(book, sheet, 1, headers)
	return book
}
func systemWriteRow(book *excelize.File, sheet string, row int, values []any) {
	for column, value := range values {
		cell, _ := excelize.CoordinatesToCellName(column+1, row)
		_ = book.SetCellValue(sheet, cell, value)
	}
}

// RoleExport godoc
// @Summary Export roles
// @Tags System Role
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/role/export-excel [get]
func (h *Handler) RoleExport(c *gin.Context) {
	var rows []Role
	h.service.db.Where("tenant_id = ?", tenantIDFromContext(c)).Order("sort,id").Find(&rows)
	book := newSystemBook("角色数据", []any{"编号", "角色名称", "角色标识", "排序", "状态", "数据范围", "备注", "创建时间"})
	for i, row := range rows {
		systemWriteRow(book, "角色数据", i+2, []any{row.ID, row.Name, row.Code, row.Sort, row.Status, row.DataScope, row.Remark, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "角色数据.xlsx")
}

// PostExport godoc
// @Summary Export posts
// @Tags System Organization
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/post/export-excel [get]
func (h *Handler) PostExport(c *gin.Context) {
	var rows []Post
	h.service.db.Where("tenant_id = ?", tenantIDFromContext(c)).Order("sort,id").Find(&rows)
	book := newSystemBook("岗位数据", []any{"编号", "岗位名称", "岗位编码", "排序", "状态", "备注", "创建时间"})
	for i, row := range rows {
		systemWriteRow(book, "岗位数据", i+2, []any{row.ID, row.Name, row.Code, row.Sort, row.Status, row.Remark, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "岗位数据.xlsx")
}

// DictTypeExport godoc
// @Summary Export dictionary types
// @Tags System Dictionary
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/dict-type/export-excel [get]
func (h *Handler) DictTypeExport(c *gin.Context) {
	var rows []DictType
	h.service.db.Order("id").Find(&rows)
	book := newSystemBook("字典类型", []any{"编号", "名称", "类型", "状态", "备注", "创建时间"})
	for i, row := range rows {
		systemWriteRow(book, "字典类型", i+2, []any{row.ID, row.Name, row.Type, row.Status, row.Remark, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "字典类型.xlsx")
}

// DictDataExport godoc
// @Summary Export dictionary data
// @Tags System Dictionary
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/dict-data/export-excel [get]
func (h *Handler) DictDataExport(c *gin.Context) {
	var rows []DictData
	query := h.service.db.Order("dict_type,sort,id")
	if typ := c.Query("dictType"); typ != "" {
		query = query.Where("dict_type = ?", typ)
	}
	query.Find(&rows)
	book := newSystemBook("字典数据", []any{"编号", "类型", "标签", "值", "排序", "状态", "颜色", "样式", "备注"})
	for i, row := range rows {
		systemWriteRow(book, "字典数据", i+2, []any{row.ID, row.DictType, row.Label, row.Value, row.Sort, row.Status, row.ColorType, row.CSSClass, row.Remark})
	}
	excelx.Write(c, book, "字典数据.xlsx")
}

// TenantExport godoc
// @Summary Export tenants
// @Tags System Tenant
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/tenant/export-excel [get]
func (h *Handler) TenantExport(c *gin.Context) {
	var rows []Tenant
	h.service.db.Order("id").Find(&rows)
	book := newSystemBook("租户数据", []any{"编号", "租户名称", "联系人", "联系电话", "域名", "套餐编号", "状态", "过期时间", "账号额度", "创建时间"})
	for i, row := range rows {
		systemWriteRow(book, "租户数据", i+2, []any{row.ID, row.Name, row.ContactName, row.ContactMobile, row.Domain, row.PackageID, row.Status, row.ExpireTime.Format(time.DateTime), row.AccountCount, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "租户数据.xlsx")
}

// LoginLogExport godoc
// @Summary Export login logs
// @Tags System Audit
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/login-log/export-excel [get]
func (h *Handler) LoginLogExport(c *gin.Context) {
	var rows []LoginLog
	h.loginLogQuery(c).Order("id DESC").Limit(10000).Find(&rows)
	book := newSystemBook("登录日志", []any{"编号", "Trace ID", "用户编号", "用户名", "结果", "状态", "IP", "用户代理", "时间"})
	for i, row := range rows {
		systemWriteRow(book, "登录日志", i+2, []any{row.ID, row.TraceID, row.UserID, row.Username, row.Result, row.Status, row.UserIP, row.UserAgent, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "登录日志.xlsx")
}

// OperateLogExport godoc
// @Summary Export operation logs
// @Tags System Audit
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/operate-log/export-excel [get]
func (h *Handler) OperateLogExport(c *gin.Context) {
	var rows []OperateLog
	h.operateLogQuery(c).Order("id DESC").Limit(10000).Find(&rows)
	book := newSystemBook("操作日志", []any{"编号", "Trace ID", "用户", "类型", "子类型", "请求方法", "请求地址", "IP", "时间"})
	for i, row := range rows {
		systemWriteRow(book, "操作日志", i+2, []any{row.ID, row.TraceID, row.UserName, row.Type, row.SubType, row.RequestMethod, row.RequestURL, row.UserIP, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "操作日志.xlsx")
}
