package member

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"gorm.io/gorm"
)

// PointRecordPage godoc
// @Summary Page member point records
// @Tags Member Detail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/point/record/page [get]
func (h *Handler) PointRecordPage(c *gin.Context) {
	query := h.db.Model(&PointRecord{}).Where("tenant_id = ?", tenantID(c))
	if userID := c.Query("userId"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if bizType := c.Query("bizType"); bizType != "" {
		query = query.Where("biz_type = ?", bizType)
	}
	if title := strings.TrimSpace(c.Query("title")); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	respondPage(c, query, &[]PointRecord{})
}

// ExperienceRecordPage godoc
// @Summary Page member experience records
// @Tags Member Detail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/experience-record/page [get]
func (h *Handler) ExperienceRecordPage(c *gin.Context) {
	query := h.db.Model(&ExperienceRecord{}).Where("tenant_id = ?", tenantID(c))
	if userID := c.Query("userId"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if bizType := c.Query("bizType"); bizType != "" {
		query = query.Where("biz_type = ?", bizType)
	}
	if title := strings.TrimSpace(c.Query("title")); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	respondPage(c, query, &[]ExperienceRecord{})
}

// ExperienceRecordGet godoc
// @Summary Get a member experience record
// @Tags Member Detail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/experience-record/get [get]
func (h *Handler) ExperienceRecordGet(c *gin.Context) {
	var row ExperienceRecord
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "经验记录不存在")
		return
	}
	httpx.OK(c, row)
}

// SignInRecordPage godoc
// @Summary Page member sign-in records
// @Tags Member Detail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/sign-in/record/page [get]
func (h *Handler) SignInRecordPage(c *gin.Context) {
	query := h.db.Model(&SignInRecord{}).Where("tenant_id = ?", tenantID(c))
	if userID := c.Query("userId"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if day := c.Query("day"); day != "" {
		query = query.Where("day = ?", day)
	}
	respondPage(c, query, &[]SignInRecord{})
}

// AddressList godoc
// @Summary List a member's addresses
// @Tags Member Detail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/address/list [get]
func (h *Handler) AddressList(c *gin.Context) {
	var rows []Address
	h.db.Where("tenant_id = ? AND user_id = ?", tenantID(c), c.Query("userId")).Order("default_status DESC,id DESC").Find(&rows)
	httpx.OK(c, rows)
}

func respondPage[T any](c *gin.Context, query *gorm.DB, rows *[]T) {
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(rows)
	httpx.OK(c, gin.H{"list": *rows, "total": total})
}
