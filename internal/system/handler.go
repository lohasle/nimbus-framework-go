package system

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/httpx"
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
		"roles": []string{"super_admin"}, "permissions": []string{"*:*:*"}, "menus": DefaultMenus(),
	})
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
