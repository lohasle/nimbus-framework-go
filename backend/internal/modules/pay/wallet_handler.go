package pay

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"gorm.io/gorm"
)

// WalletGet godoc
// @Summary Get or initialize a member wallet
// @Tags Pay Wallet
// @Produce json
// @Security BearerAuth
// @Param userId query int true "Member user ID"
// @Success 200 {object} httpx.Response
// @Router /pay/wallet/get [get]
func (h *Handler) WalletGet(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	if userID == 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "会员编号不能为空")
		return
	}
	wallet, err := h.getOrCreateWallet(tenantID(c), userID)
	if err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "初始化钱包失败")
		return
	}
	httpx.OK(c, wallet)
}

// WalletPage godoc
// @Summary Page member wallets
// @Tags Pay Wallet
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/wallet/page [get]
func (h *Handler) WalletPage(c *gin.Context) {
	query := h.db.Model(&Wallet{}).Where("tenant_id = ?", tenantID(c))
	if userID := c.Query("userId"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []Wallet
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

type WalletBalanceUpdateRequest struct {
	UserID  uint64 `json:"userId" binding:"required"`
	Balance int64  `json:"balance" binding:"required"`
}

// WalletBalanceUpdate godoc
// @Summary Adjust a member wallet balance
// @Tags Pay Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body WalletBalanceUpdateRequest true "Balance adjustment in cents"
// @Success 200 {object} httpx.Response
// @Router /pay/wallet/update-balance [put]
func (h *Handler) WalletBalanceUpdate(c *gin.Context) {
	var req WalletBalanceUpdateRequest
	if c.ShouldBindJSON(&req) != nil || req.Balance == 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "变动余额不能为空")
		return
	}
	wallet, err := h.getOrCreateWallet(tenantID(c), req.UserID)
	if err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "查询钱包失败")
		return
	}
	newBalance := wallet.Balance + req.Balance
	if newBalance < 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "余额不足")
		return
	}
	tx := h.db.Begin()
	if tx.Model(&Wallet{}).Where("tenant_id = ? AND id = ?", tenantID(c), wallet.ID).Update("balance", newBalance).Error != nil {
		tx.Rollback()
		httpx.Fail(c, http.StatusInternalServerError, 500, "更新钱包失败")
		return
	}
	if tx.Create(&WalletTransaction{TenantID: tenantID(c), WalletID: wallet.ID, Title: "管理员调整余额", Price: req.Balance, Balance: newBalance}).Error != nil {
		tx.Rollback()
		httpx.Fail(c, http.StatusInternalServerError, 500, "记录钱包流水失败")
		return
	}
	// member_user.balance is retained for list-page compatibility.
	tx.Table("member_user").Where("tenant_id = ? AND id = ?", tenantID(c), req.UserID).Update("balance", newBalance)
	if tx.Commit().Error != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "更新钱包失败")
		return
	}
	httpx.OK(c, true)
}

// WalletTransactionPage godoc
// @Summary Page wallet transactions
// @Tags Pay Wallet
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /pay/wallet-transaction/page [get]
func (h *Handler) WalletTransactionPage(c *gin.Context) {
	query := h.db.Model(&WalletTransaction{}).Where("tenant_id = ?", tenantID(c))
	if walletID := c.Query("walletId"); walletID != "" && walletID != "null" && walletID != "undefined" {
		query = query.Where("wallet_id = ?", walletID)
	}
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []WalletTransaction
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

func (h *Handler) getOrCreateWallet(tenant, userID uint64) (Wallet, error) {
	wallet := Wallet{TenantID: tenant, UserID: userID, UserType: 1}
	err := h.db.Where("tenant_id = ? AND user_id = ? AND user_type = 1", tenant, userID).
		Attrs(Wallet{Balance: memberBalance(h.db, tenant, userID)}).FirstOrCreate(&wallet).Error
	return wallet, err
}

func memberBalance(db *gorm.DB, tenant, userID uint64) int64 {
	var balance int64
	db.Table("member_user").Select("balance").Where("tenant_id = ? AND id = ?", tenant, userID).Scan(&balance)
	return balance
}
