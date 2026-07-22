package infra

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/excelx"
	"github.com/xuri/excelize/v2"
)

// ConfigExport godoc
// @Summary Export system parameters
// @Tags Infra Config
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /infra/config/export-excel [get]
func (h *Handler) ConfigExport(c *gin.Context) {
	var rows []Config
	h.db.Where("tenant_id = ?", tenantID(c)).Order("id").Find(&rows)
	book := excelize.NewFile()
	sheet := "参数配置"
	book.SetSheetName("Sheet1", sheet)
	writeRow(book, sheet, 1, []any{"编号", "分类", "参数名称", "参数键", "参数值", "是否公开", "备注", "创建时间"})
	for index, row := range rows {
		writeRow(book, sheet, index+2, []any{row.ID, row.Category, row.Name, row.Key, row.Value, row.Visible, row.Remark, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "参数配置.xlsx")
}

// AccessLogExport godoc
// @Summary Export API access logs
// @Tags Infra Logging
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /infra/api-access-log/export-excel [get]
func (h *Handler) AccessLogExport(c *gin.Context) {
	var rows []APIAccessLog
	h.accessLogQuery(c).Order("id DESC").Limit(10000).Find(&rows)
	book := excelize.NewFile()
	sheet := "API访问日志"
	book.SetSheetName("Sheet1", sheet)
	writeRow(book, sheet, 1, []any{"编号", "Trace ID", "用户编号", "请求方法", "请求地址", "响应状态", "耗时(ms)", "IP", "创建时间"})
	for index, row := range rows {
		writeRow(book, sheet, index+2, []any{row.ID, row.TraceID, row.UserID, row.Method, row.Path, row.Status, row.Duration, row.IP, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "API访问日志.xlsx")
}

func writeRow(book *excelize.File, sheet string, row int, values []any) {
	for column, value := range values {
		cell, _ := excelize.CoordinatesToCellName(column+1, row)
		_ = book.SetCellValue(sheet, cell, value)
	}
}
