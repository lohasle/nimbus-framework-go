package system

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
)

// SimpleRoles godoc
// @Summary List enabled roles
// @Tags System Role
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/role/simple-list [get]
func (h *Handler) SimpleRoles(c *gin.Context) {
	var rows []Role
	h.service.db.Where("tenant_id = ? AND status = 0", tenantIDFromContext(c)).Order("sort,id").Find(&rows)
	httpx.OK(c, rows)
}

// UserRoleList godoc
// @Summary List role IDs assigned to a user
// @Tags System Permission
// @Produce json
// @Security BearerAuth
// @Param userId query int true "User ID"
// @Success 200 {object} httpx.Response
// @Router /system/permission/list-user-roles [get]
func (h *Handler) UserRoleList(c *gin.Context) {
	userID := queryUint64(c, "userId")
	if userID == 0 || h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), userID).First(&AdminUser{}).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "用户不存在")
		return
	}
	var roleIDs []uint64
	h.service.db.Model(&UserRole{}).Where("user_id = ?", userID).Order("role_id").Pluck("role_id", &roleIDs)
	httpx.OK(c, roleIDs)
}

type AssignUserRoleRequest struct {
	UserID  uint64   `json:"userId" binding:"required"`
	RoleIDs []uint64 `json:"roleIds"`
}

// AssignUserRole godoc
// @Summary Assign roles to an operations-console user
// @Tags System Permission
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AssignUserRoleRequest true "User role assignment"
// @Success 200 {object} httpx.Response
// @Router /system/permission/assign-user-role [post]
func (h *Handler) AssignUserRole(c *gin.Context) {
	var req AssignUserRoleRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	tenantID := tenantIDFromContext(c)
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantID, req.UserID).First(&AdminUser{}).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "用户不存在")
		return
	}
	if len(req.RoleIDs) > 0 {
		var count int64
		h.service.db.Model(&Role{}).Where("tenant_id = ? AND id IN ?", tenantID, req.RoleIDs).Count(&count)
		if count != int64(len(uniqueIDs(req.RoleIDs))) {
			httpx.Fail(c, http.StatusBadRequest, 400, "包含无效角色")
			return
		}
	}
	tx := h.service.db.Begin()
	tx.Where("user_id = ?", req.UserID).Delete(&UserRole{})
	for _, roleID := range uniqueIDs(req.RoleIDs) {
		tx.Create(&UserRole{UserID: req.UserID, RoleID: roleID})
	}
	if tx.Commit().Error != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "分配角色失败")
		return
	}
	httpx.OK(c, true)
}

func queryUint64(c *gin.Context, key string) uint64 {
	value, _ := strconv.ParseUint(c.Query(key), 10, 64)
	return value
}

func uniqueIDs(values []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
