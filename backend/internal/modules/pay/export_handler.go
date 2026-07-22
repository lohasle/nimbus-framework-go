package pay

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/excelx"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"github.com/xuri/excelize/v2"
)

// ChannelExport godoc
// @Summary Export payment channels
// @Tags Pay Channel
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /pay/channel/export-excel [get]
func (h *Handler) ChannelExport(c *gin.Context) {
	var rows []Channel
	h.db.Where("tenant_id = ?", tenantID(c)).Order("id").Find(&rows)
	book := excelize.NewFile()
	sheet := "支付渠道"
	book.SetSheetName("Sheet1", sheet)
	payWriteRow(book, sheet, 1, []any{"编号", "应用编号", "渠道编码", "状态", "费率", "备注", "创建时间"})
	for index, row := range rows {
		payWriteRow(book, sheet, index+2, []any{row.ID, row.AppID, row.Code, row.Status, row.FeeRate, row.Remark, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "支付渠道.xlsx")
}

// OrderExport godoc
// @Summary Export payment orders
// @Tags Pay Order
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /pay/order/export-excel [get]
func (h *Handler) OrderExport(c *gin.Context) {
	var rows []Order
	h.db.Where("tenant_id = ?", tenantID(c)).Order("id DESC").Limit(10000).Find(&rows)
	book := excelize.NewFile()
	sheet := "支付订单"
	book.SetSheetName("Sheet1", sheet)
	payWriteRow(book, sheet, 1, []any{"编号", "应用编号", "渠道", "商户订单号", "渠道订单号", "标题", "金额(分)", "状态", "退款金额(分)", "创建时间"})
	for index, row := range rows {
		payWriteRow(book, sheet, index+2, []any{row.ID, row.AppID, row.ChannelCode, row.MerchantOrderNo, row.ChannelOrderNo, row.Subject, row.Price, row.Status, row.RefundPrice, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "支付订单.xlsx")
}

// RefundExport godoc
// @Summary Export payment refunds
// @Tags Pay Refund
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /pay/refund/export-excel [get]
func (h *Handler) RefundExport(c *gin.Context) {
	var rows []Refund
	h.db.Where("tenant_id = ?", tenantID(c)).Order("id DESC").Limit(10000).Find(&rows)
	book := excelize.NewFile()
	sheet := "退款订单"
	book.SetSheetName("Sheet1", sheet)
	payWriteRow(book, sheet, 1, []any{"编号", "支付订单编号", "商户退款号", "渠道退款号", "退款金额(分)", "原因", "状态", "创建时间"})
	for index, row := range rows {
		payWriteRow(book, sheet, index+2, []any{row.ID, row.OrderID, row.MerchantRefundNo, row.ChannelRefundNo, row.Price, row.Reason, row.Status, row.CreatedAt.Format(time.DateTime)})
	}
	excelx.Write(c, book, "退款订单.xlsx")
}

type RefundUpdateRequest struct {
	ID     uint64 `json:"id" binding:"required"`
	Reason string `json:"reason"`
	Status int    `json:"status"`
}

// RefundUpdate godoc
// @Summary Update a pending refund record
// @Tags Pay Refund
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RefundUpdateRequest true "Refund update"
// @Success 200 {object} httpx.Response
// @Router /pay/refund/update [put]
func (h *Handler) RefundUpdate(c *gin.Context) {
	var req RefundUpdateRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	result := h.db.Model(&Refund{}).Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).
		Updates(map[string]any{"reason": req.Reason, "status": req.Status})
	if result.Error != nil || result.RowsAffected == 0 {
		httpx.Fail(c, http.StatusNotFound, 404, "退款单不存在")
		return
	}
	httpx.OK(c, true)
}

func payWriteRow(book *excelize.File, sheet string, row int, values []any) {
	for column, value := range values {
		cell, _ := excelize.CoordinatesToCellName(column+1, row)
		_ = book.SetCellValue(sheet, cell, value)
	}
}
