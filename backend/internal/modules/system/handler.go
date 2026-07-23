package system

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// TenantID godoc
// @Summary Resolve tenant ID
// @Description Returns an enabled tenant ID by tenant name.
// @Tags System Auth
// @Produce json
// @Param name query string true "Tenant name"
// @Success 200 {object} httpx.Response
// @Router /system/tenant/get-id-by-name [get]
func (h *Handler) TenantID(c *gin.Context) {
	id, err := h.service.TenantID(c.Query("name"))
	if err != nil {
		httpx.Fail(c, http.StatusNotFound, 1001, "租户不存在")
		return
	}
	httpx.OK(c, id)
}

// TenantByWebsite godoc
// @Summary Resolve tenant by website host
// @Description Returns an enabled tenant matching the request host, or null when no custom host is configured.
// @Tags System Tenant
// @Produce json
// @Param website query string true "Website host"
// @Success 200 {object} httpx.Response
// @Router /system/tenant/get-by-website [get]
func (h *Handler) TenantByWebsite(c *gin.Context) {
	website := strings.TrimSpace(c.Query("website"))
	host := strings.Split(website, ":")[0]
	var tenant Tenant
	if err := h.service.db.Where("status = 0 AND domain IN ?", []string{website, host}).First(&tenant).Error; err != nil {
		httpx.OK(c, nil)
		return
	}
	httpx.OK(c, tenant)
}

// Login godoc
// @Summary Admin login
// @Description Authenticates an operations-console user in a tenant.
// @Tags System Auth
// @Accept json
// @Produce json
// @Param tenant-id header int true "Tenant ID"
// @Param request body LoginRequest true "Credentials"
// @Success 200 {object} httpx.Response
// @Failure 401 {object} httpx.Response
// @Router /system/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	tenantID, _ := strconv.ParseUint(c.GetHeader("tenant-id"), 10, 64)
	var req LoginRequest
	if tenantID == 0 || c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	token, err := h.service.Login(tenantID, req)
	if err != nil {
		h.service.db.Create(&LoginLog{TenantID: tenantID, LogType: 100, TraceID: c.GetString("trace_id"), Username: req.Username, Result: 1, Status: 1, UserIP: c.ClientIP(), UserAgent: c.Request.UserAgent()})
		httpx.Fail(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	h.service.db.Model(&AdminUser{}).Where("id = ?", token.UserID).Updates(map[string]any{"login_ip": c.ClientIP(), "login_date": time.Now()})
	h.service.db.Create(&LoginLog{TenantID: tenantID, LogType: 100, TraceID: c.GetString("trace_id"), UserID: token.UserID, UserType: 2, Username: req.Username, Result: 0, Status: 0, UserIP: c.ClientIP(), UserAgent: c.Request.UserAgent()})
	httpx.OK(c, token)
}

// RefreshToken godoc
// @Summary Refresh admin tokens
// @Description Validates and rotates a refresh token, returning a new access/refresh token pair.
// @Tags System Auth
// @Produce json
// @Param refreshToken query string true "Refresh token"
// @Success 200 {object} httpx.Response
// @Failure 401 {object} httpx.Response
// @Router /system/auth/refresh-token [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("refreshToken"))
	if raw == "" {
		httpx.Fail(c, http.StatusUnauthorized, 401, "无效的刷新令牌")
		return
	}
	token, err := h.service.RefreshToken(raw)
	if err != nil {
		httpx.Fail(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	httpx.OK(c, token)
}

// PermissionInfo godoc
// @Summary Current user and menus
// @Description Returns current admin profile, roles, permissions and menu tree.
// @Tags System Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Failure 401 {object} httpx.Response
// @Router /system/auth/get-permission-info [get]
func (h *Handler) PermissionInfo(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		httpx.Fail(c, http.StatusUnauthorized, 401, "未登录")
		return
	}
	user, err := h.service.User(uid.(uint64))
	if err != nil {
		httpx.Fail(c, http.StatusUnauthorized, 401, "用户不存在")
		return
	}
	roles, permissions, menus := h.permissionData(user)
	httpx.OK(c, gin.H{
		"user":        gin.H{"id": user.ID, "username": user.Username, "nickname": user.Nickname, "avatar": user.Avatar, "deptId": user.DeptID, "email": user.Email},
		"roles":       roles,
		"permissions": permissions,
		"menus":       menus,
	})
}

func (h *Handler) permissionData(user AdminUser) ([]string, []string, []Menu) {
	var roles []Role
	h.service.db.Table("roles").Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = 0", user.ID).Order("roles.sort,roles.id").Find(&roles)
	roleCodes := make([]string, 0, len(roles))
	roleIDs := make([]uint64, 0, len(roles))
	superAdmin := false
	for _, role := range roles {
		roleCodes = append(roleCodes, role.Code)
		roleIDs = append(roleIDs, role.ID)
		if role.Code == "super_admin" {
			superAdmin = true
		}
	}
	query := h.service.db.Where("status = 0")
	allowedIDs := h.allowedMenuIDs(user.TenantID)
	if len(allowedIDs) > 0 {
		query = query.Where("id IN ?", allowedIDs)
	}
	if !superAdmin {
		var menuIDs []uint64
		h.service.db.Model(&RoleMenu{}).Where("role_id IN ?", roleIDs).Pluck("menu_id", &menuIDs)
		menuIDs = h.withMenuAncestors(menuIDs)
		query = query.Where("id IN ?", menuIDs)
	}
	var rows []SystemMenu
	query.Order("sort,id").Find(&rows)
	permissions := make([]string, 0)
	for _, row := range rows {
		if row.Permission != "" {
			permissions = append(permissions, row.Permission)
		}
	}
	return roleCodes, permissions, buildMenuTree(rows)
}

func (h *Handler) allowedMenuIDs(tenantID uint64) []uint64 {
	var tenant Tenant
	if h.service.db.First(&tenant, tenantID).Error != nil || tenant.PackageID == 0 {
		return nil
	}
	var pack TenantPackage
	if h.service.db.First(&pack, tenant.PackageID).Error != nil {
		return nil
	}
	var ids []uint64
	if json.Unmarshal([]byte(pack.MenuIDs), &ids) != nil {
		return nil
	}
	if len(ids) == 0 {
		return []uint64{0}
	}
	var buttonIDs []uint64
	h.service.db.Model(&SystemMenu{}).Where("type = 3 AND parent_id IN ?", ids).Pluck("id", &buttonIDs)
	return uniqueIDs(append(ids, buttonIDs...))
}

func (h *Handler) withMenuAncestors(ids []uint64) []uint64 {
	result := uniqueIDs(ids)
	seen := make(map[uint64]struct{}, len(result))
	for _, id := range result {
		seen[id] = struct{}{}
	}
	current := append([]uint64{}, result...)
	for len(current) > 0 {
		var parents []uint64
		h.service.db.Model(&SystemMenu{}).Where("id IN ? AND parent_id > 0", current).Pluck("parent_id", &parents)
		current = current[:0]
		for _, parent := range parents {
			if _, ok := seen[parent]; !ok {
				seen[parent] = struct{}{}
				result = append(result, parent)
				current = append(current, parent)
			}
		}
	}
	return result
}

func buildMenuTree(rows []SystemMenu) []Menu {
	byParent := make(map[uint64][]SystemMenu)
	for _, row := range rows {
		if row.Type == 1 || row.Type == 2 {
			byParent[row.ParentID] = append(byParent[row.ParentID], row)
		}
	}
	var build func(uint64) []Menu
	build = func(parentID uint64) []Menu {
		result := make([]Menu, 0, len(byParent[parentID]))
		for _, row := range byParent[parentID] {
			var component, componentName *string
			if row.Component != "" {
				value := row.Component
				component = &value
			}
			if row.ComponentName != "" {
				value := row.ComponentName
				componentName = &value
			}
			result = append(result, Menu{ID: row.ID, ParentID: row.ParentID, Name: row.Name, Path: row.Path, Component: component, ComponentName: componentName, Icon: row.Icon, Visible: row.Visible, KeepAlive: row.KeepAlive, AlwaysShow: row.AlwaysShow, Children: build(row.ID)})
		}
		return result
	}
	return build(0)
}

// SimpleDictData godoc
// @Summary Frontend dictionary bootstrap
// @Description Returns enabled dictionary entries required by the operations console.
// @Tags System Bootstrap
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-data/simple-list [get]
func (h *Handler) SimpleDictData(c *gin.Context) {
	var rows []DictData
	h.service.db.Where("status = 0").Order("dict_type,sort,id").Find(&rows)
	httpx.OK(c, rows)
}

// UnreadNotifyMessageCount godoc
// @Summary Count current user's unread messages
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-message/get-unread-count [get]
func (h *Handler) UnreadNotifyMessageCount(c *gin.Context) {
	var count int64
	h.service.db.Model(&NotifyMessage{}).
		Where("tenant_id = ? AND user_id = ? AND read_status = ?", tenantIDFromContext(c), c.GetUint64("user_id"), false).
		Count(&count)
	httpx.OK(c, count)
}

// UnreadNotifyMessageList godoc
// @Summary List current user's latest unread messages
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-message/get-unread-list [get]
func (h *Handler) UnreadNotifyMessageList(c *gin.Context) {
	var rows []NotifyMessage
	h.service.db.Where("tenant_id = ? AND user_id = ? AND read_status = ?", tenantIDFromContext(c), c.GetUint64("user_id"), false).
		Order("id DESC").Limit(10).Find(&rows)
	httpx.OK(c, rows)
}

// Logout godoc
// @Summary Admin logout
// @Description Ends the client session. JWT tokens expire naturally in this scaffold.
// @Tags System Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) { httpx.OK(c, true) }

func (h *Handler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer"))
		uid, tenantID, err := h.service.ParseToken(raw)
		if err != nil {
			httpx.Fail(c, http.StatusUnauthorized, 401, "登录状态已失效")
			return
		}
		c.Set("user_id", uid)
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}
