package system

import (
	"net/http"
	"strconv"
	"strings"

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
	httpx.OK(c, gin.H{
		"user":  gin.H{"id": user.ID, "username": user.Username, "nickname": user.Nickname, "avatar": "", "deptId": user.DeptID, "email": user.Email},
		"roles": []string{"super_admin"},
		"permissions": []string{
			"system:user:query", "system:user:create", "system:user:update",
			"system:user:delete", "system:user:update-password", "system:user:import", "system:user:export",
			"system:permission:assign-user-role",
			"infra:config:query", "infra:config:create", "infra:config:update", "infra:config:delete", "infra:config:export",
			"infra:file-config:query", "infra:file-config:create", "infra:file-config:update", "infra:file-config:delete",
			"infra:api-access-log:query", "infra:api-access-log:export",
			"member:user:query", "member:user:update", "member:user:update-level", "member:user:update-point",
			"member:level:query", "member:level:create", "member:level:update", "member:level:delete",
			"member:group:query", "member:group:create", "member:group:update", "member:group:delete",
			"member:tag:query", "member:tag:create", "member:tag:update", "member:tag:delete",
			"pay:app:query", "pay:app:create", "pay:app:update", "pay:app:delete",
			"pay:channel:query", "pay:channel:create", "pay:channel:update", "pay:channel:delete",
			"pay:wallet:update-balance", "pay:order:query", "pay:order:export",
			"pay:refund:query", "pay:refund:create", "pay:refund:delete", "system:tenant:export",
		},
		"menus": DefaultMenus(),
	})
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
