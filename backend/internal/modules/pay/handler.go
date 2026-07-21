package pay

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// AppPage godoc
// @Summary Page payment applications
// @Tags Pay App
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/app/page [get]
func (h *Handler) AppPage(c *gin.Context) {
	query := h.db.Model(&App{}).Where("tenant_id = ?", tenantID(c))
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []App
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// AppSimple godoc
// @Summary List enabled payment applications
// @Tags Pay App
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/app/simple-list [get]
func (h *Handler) AppSimple(c *gin.Context) {
	var rows []App
	h.db.Where("tenant_id = ? AND status = 0", tenantID(c)).Order("id").Find(&rows)
	httpx.OK(c, rows)
}

// AppList godoc
// @Summary List payment applications
// @Tags Pay App
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/app/list [get]
func (h *Handler) AppList(c *gin.Context) {
	var rows []App
	h.db.Where("tenant_id = ?", tenantID(c)).Order("id").Find(&rows)
	httpx.OK(c, rows)
}

// AppGet godoc
// @Summary Get a payment application
// @Tags Pay App
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/app/get [get]
func (h *Handler) AppGet(c *gin.Context) {
	var row App
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "支付应用不存在")
		return
	}
	httpx.OK(c, row)
}

// AppCreate godoc
// @Summary Create a payment application
// @Tags Pay App
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AppSaveRequest true "Payment app"
// @Success 200 {object} httpx.Response
// @Router /pay/app/create [post]
func (h *Handler) AppCreate(c *gin.Context) { h.saveApp(c, false) }

// AppUpdate godoc
// @Summary Update a payment application
// @Tags Pay App
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AppSaveRequest true "Payment app"
// @Success 200 {object} httpx.Response
// @Router /pay/app/update [put]
func (h *Handler) AppUpdate(c *gin.Context) { h.saveApp(c, true) }

// AppDelete godoc
// @Summary Delete a payment application
// @Tags Pay App
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/app/delete [delete]
func (h *Handler) AppDelete(c *gin.Context) {
	h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).Delete(&App{})
	httpx.OK(c, true)
}

// AppStatus godoc
// @Summary Change a payment application status
// @Tags Pay App
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/app/update-status [put]
func (h *Handler) AppStatus(c *gin.Context) {
	var req struct {
		ID     uint64 `json:"id" binding:"required"`
		Status int    `json:"status"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	h.db.Model(&App{}).Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).Update("status", req.Status)
	httpx.OK(c, true)
}

// ChannelPage godoc
// @Summary Page payment channels
// @Tags Pay Channel
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/channel/page [get]
func (h *Handler) ChannelPage(c *gin.Context) {
	query := h.db.Model(&Channel{}).Where("tenant_id = ?", tenantID(c))
	if appID := c.Query("appId"); appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []Channel
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// ChannelCodes godoc
// @Summary List enabled channel codes for an application
// @Tags Pay Channel
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/channel/get-enable-channel-code-list [get]
func (h *Handler) ChannelCodes(c *gin.Context) {
	var codes []string
	h.db.Model(&Channel{}).Where("tenant_id = ? AND app_id = ? AND status = 0", tenantID(c), c.Query("appId")).Order("id").Pluck("code", &codes)
	httpx.OK(c, codes)
}

// ChannelGet godoc
// @Summary Get a payment channel
// @Tags Pay Channel
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/channel/get [get]
func (h *Handler) ChannelGet(c *gin.Context) {
	var row Channel
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "支付渠道不存在")
		return
	}
	httpx.OK(c, row)
}

// ChannelCreate godoc
// @Summary Create a payment channel
// @Tags Pay Channel
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChannelSaveRequest true "Payment channel"
// @Success 200 {object} httpx.Response
// @Router /pay/channel/create [post]
func (h *Handler) ChannelCreate(c *gin.Context) { h.saveChannel(c, false) }

// ChannelUpdate godoc
// @Summary Update a payment channel
// @Tags Pay Channel
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChannelSaveRequest true "Payment channel"
// @Success 200 {object} httpx.Response
// @Router /pay/channel/update [put]
func (h *Handler) ChannelUpdate(c *gin.Context) { h.saveChannel(c, true) }

// ChannelDelete godoc
// @Summary Delete a payment channel
// @Tags Pay Channel
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/channel/delete [delete]
func (h *Handler) ChannelDelete(c *gin.Context) {
	h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).Delete(&Channel{})
	httpx.OK(c, true)
}

// OrderPage godoc
// @Summary Page payment orders
// @Tags Pay Order
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/order/page [get]
func (h *Handler) OrderPage(c *gin.Context) {
	query := h.db.Model(&Order{}).Where("tenant_id = ?", tenantID(c))
	if appID := c.Query("appId"); appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if merchant := strings.TrimSpace(c.Query("merchantOrderId")); merchant != "" {
		query = query.Where("merchant_order_no LIKE ?", "%"+merchant+"%")
	}
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []Order
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// OrderGet godoc
// @Summary Get a payment order
// @Tags Pay Order
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/order/get [get]
func (h *Handler) OrderGet(c *gin.Context) {
	var row Order
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "支付订单不存在")
		return
	}
	httpx.OK(c, row)
}

// OrderDetail godoc
// @Summary Get a payment order detail
// @Tags Pay Order
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/order/get-detail [get]
func (h *Handler) OrderDetail(c *gin.Context) { h.OrderGet(c) }

// OrderCreate godoc
// @Summary Create a pending payment order
// @Tags Pay Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body OrderCreateRequest true "Payment order"
// @Success 200 {object} httpx.Response
// @Router /pay/order/create [post]
func (h *Handler) OrderCreate(c *gin.Context) {
	var req OrderCreateRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	row := Order{TenantID: tenantID(c), AppID: req.AppID, ChannelCode: req.ChannelCode, MerchantOrderNo: req.MerchantOrderNo, Subject: req.Subject, Body: req.Body, Price: req.Price, ClientIP: c.ClientIP()}
	if h.db.Create(&row).Error != nil {
		httpx.Fail(c, http.StatusConflict, 409, "商户订单号已存在")
		return
	}
	httpx.OK(c, row.ID)
}

// RefundPage godoc
// @Summary Page payment refunds
// @Tags Pay Refund
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/refund/page [get]
func (h *Handler) RefundPage(c *gin.Context) {
	query := h.db.Model(&Refund{}).Where("tenant_id = ?", tenantID(c))
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []Refund
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// RefundGet godoc
// @Summary Get a payment refund
// @Tags Pay Refund
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/refund/get [get]
func (h *Handler) RefundGet(c *gin.Context) {
	var row Refund
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "退款单不存在")
		return
	}
	httpx.OK(c, row)
}

// RefundCreate godoc
// @Summary Create a pending refund request
// @Tags Pay Refund
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RefundCreateRequest true "Refund"
// @Success 200 {object} httpx.Response
// @Router /pay/refund/create [post]
func (h *Handler) RefundCreate(c *gin.Context) {
	var req RefundCreateRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	var order Order
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.OrderID).First(&order).Error != nil || req.Price > order.Price-order.RefundPrice {
		httpx.Fail(c, http.StatusBadRequest, 400, "退款金额无效")
		return
	}
	tx := h.db.Begin()
	row := Refund{TenantID: tenantID(c), OrderID: order.ID, MerchantRefundNo: req.MerchantRefundNo, ChannelRefundNo: uuid.NewString(), Price: req.Price, Reason: req.Reason}
	if tx.Create(&row).Error != nil {
		tx.Rollback()
		httpx.Fail(c, http.StatusConflict, 409, "商户退款号已存在")
		return
	}
	tx.Model(&order).Update("refund_price", gorm.Expr("refund_price + ?", req.Price))
	tx.Commit()
	httpx.OK(c, row.ID)
}

// RefundDelete godoc
// @Summary Delete a pending refund record
// @Tags Pay Refund
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/refund/delete [delete]
func (h *Handler) RefundDelete(c *gin.Context) {
	var refund Refund
	if h.db.Where("tenant_id = ? AND id = ? AND status = 0", tenantID(c), queryID(c)).First(&refund).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "待处理退款单不存在")
		return
	}
	tx := h.db.Begin()
	tx.Model(&Order{}).Where("tenant_id = ? AND id = ?", tenantID(c), refund.OrderID).
		Update("refund_price", gorm.Expr("GREATEST(refund_price - ?, 0)", refund.Price))
	tx.Delete(&refund)
	tx.Commit()
	httpx.OK(c, true)
}

func (h *Handler) saveApp(c *gin.Context, update bool) {
	var req AppSaveRequest
	if c.ShouldBindJSON(&req) != nil || (update && req.ID == 0) {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	if req.AppKey == "" {
		req.AppKey = strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	row := App{ID: req.ID, TenantID: tenantID(c), Name: req.Name, AppKey: req.AppKey, Status: req.Status, Remark: req.Remark}
	if update {
		h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).Updates(&row)
		httpx.OK(c, true)
		return
	}
	if h.db.Create(&row).Error != nil {
		httpx.Fail(c, http.StatusConflict, 409, "应用标识已存在")
		return
	}
	httpx.OK(c, row.ID)
}

func (h *Handler) saveChannel(c *gin.Context, update bool) {
	var req ChannelSaveRequest
	if c.ShouldBindJSON(&req) != nil || (update && req.ID == 0) {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	row := Channel{ID: req.ID, TenantID: tenantID(c), AppID: req.AppID, Code: req.Code, Status: req.Status, FeeRate: req.FeeRate, Config: req.Config, Remark: req.Remark}
	if update {
		h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).Updates(&row)
		httpx.OK(c, true)
		return
	}
	if h.db.Create(&row).Error != nil {
		httpx.Fail(c, http.StatusConflict, 409, "该应用已存在同名渠道")
		return
	}
	httpx.OK(c, row.ID)
}
