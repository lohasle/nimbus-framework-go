package system

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	TenantName string `json:"tenantName"`
	Username   string `json:"username" binding:"required"`
	Nickname   string `json:"nickname"`
	Password   string `json:"password" binding:"required"`
	Mobile     string `json:"mobile"`
}

func requestTenantID(c *gin.Context, service *Service, tenantName string) uint64 {
	id, _ := strconv.ParseUint(c.GetHeader("tenant-id"), 10, 64)
	if id == 0 && strings.TrimSpace(tenantName) != "" {
		id, _ = service.TenantID(strings.TrimSpace(tenantName))
	}
	return id
}

// Register godoc
// @Summary Register an operations-console user
// @Description Creates a tenant admin account without roles. An administrator assigns roles after registration.
// @Tags System Auth
// @Accept json
// @Produce json
// @Param tenant-id header int false "Tenant ID"
// @Param request body registerRequest true "Registration"
// @Success 200 {object} httpx.Response
// @Router /system/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if c.ShouldBindJSON(&req) != nil || !regexp.MustCompile(`^[a-zA-Z0-9]{4,30}$`).MatchString(req.Username) || len(req.Password) < 4 || len(req.Password) > 72 {
		httpx.Fail(c, http.StatusBadRequest, 400, "账号须为 4-30 位字母或数字，密码至少 4 位")
		return
	}
	tenant := requestTenantID(c, h.service, req.TenantName)
	var tenantRow Tenant
	if tenant == 0 || h.service.db.Where("id = ? AND status = 0", tenant).First(&tenantRow).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "租户不存在")
		return
	}
	var count int64
	h.service.db.Model(&AdminUser{}).Where("tenant_id = ?", tenant).Count(&count)
	if tenantRow.AccountCount > 0 && count >= int64(tenantRow.AccountCount) {
		httpx.Fail(c, http.StatusConflict, 409, "租户账号数已达上限")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = req.Username
	}
	row := AdminUser{TenantID: tenant, Username: req.Username, PasswordHash: string(hash), Nickname: nickname, Mobile: req.Mobile, DeptID: 1, Status: 0, LoginDate: time.Now()}
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, http.StatusConflict, 409, "账号已经存在")
		return
	}
	token, err := h.service.issueTokenPair(row)
	if err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "签发令牌失败")
		return
	}
	httpx.OK(c, token)
}

type authSMSSendRequest struct {
	Mobile string `json:"mobile" binding:"required"`
	Scene  int    `json:"scene" binding:"required"`
}

func newSMSCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

// SendAuthSMSCode godoc
// @Summary Send an authentication SMS code
// @Description Generates a five-minute one-time code and dispatches it through the configured SMS adapter when present.
// @Tags System Auth
// @Accept json
// @Produce json
// @Param tenant-id header int true "Tenant ID"
// @Param request body authSMSSendRequest true "SMS code request"
// @Success 200 {object} httpx.Response
// @Router /system/auth/send-sms-code [post]
func (h *Handler) SendAuthSMSCode(c *gin.Context) {
	var req authSMSSendRequest
	tenant := requestTenantID(c, h.service, "")
	if tenant == 0 || c.ShouldBindJSON(&req) != nil || len(strings.TrimSpace(req.Mobile)) < 6 {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	code, err := newSMSCode()
	if err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "生成验证码失败")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	h.service.db.Model(&AuthSMSCode{}).Where("tenant_id = ? AND mobile = ? AND scene = ? AND used = 0", tenant, req.Mobile, req.Scene).Update("used", true)
	h.service.db.Create(&AuthSMSCode{TenantID: tenant, Mobile: req.Mobile, Scene: req.Scene, CodeHash: string(hash), ExpireAt: time.Now().Add(5 * time.Minute)})

	// The auth contract remains usable without baking a vendor SDK into the base.
	// If a generic callback channel exists, dispatch through the same adapter used by SMS templates.
	var channel SMSChannel
	if h.service.db.Where("tenant_id = ? AND status = 0 AND callback_url <> ''", tenant).Order("id").First(&channel).Error == nil {
		now := time.Now()
		log := SMSLog{TenantID: tenant, ChannelID: channel.ID, ChannelCode: channel.Code, Mobile: req.Mobile, TemplateCode: "auth-code", TemplateContent: "验证码：" + code, SendStatus: 0, SendTime: &now}
		h.service.db.Create(&log)
		requestID, sendErr := sendSMSCallback(channel, gin.H{"mobile": req.Mobile, "signature": channel.Signature, "scene": req.Scene, "code": code})
		if sendErr != nil {
			h.service.db.Model(&log).Updates(map[string]any{"send_status": 2, "api_send_msg": sendErr.Error()})
			httpx.Fail(c, http.StatusBadGateway, 502, "短信发送失败")
			return
		}
		h.service.db.Model(&log).Updates(map[string]any{"send_status": 1, "api_request_id": requestID, "api_send_code": "OK"})
	}
	httpx.OK(c, true)
}

func (h *Handler) consumeSMSCode(tenant uint64, mobile, code string, scene int) bool {
	var row AuthSMSCode
	if h.service.db.Where("tenant_id = ? AND mobile = ? AND scene = ? AND used = 0 AND expire_at > ?", tenant, mobile, scene, time.Now()).Order("id DESC").First(&row).Error != nil {
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(row.CodeHash), []byte(code)) != nil {
		return false
	}
	h.service.db.Model(&row).Update("used", true)
	return true
}

// SMSLogin godoc
// @Summary Log in with an SMS verification code
// @Tags System Auth
// @Accept json
// @Produce json
// @Param tenant-id header int true "Tenant ID"
// @Success 200 {object} httpx.Response
// @Router /system/auth/sms-login [post]
func (h *Handler) SMSLogin(c *gin.Context) {
	var req struct {
		Mobile string `json:"mobile" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	tenant := requestTenantID(c, h.service, "")
	if tenant == 0 || c.ShouldBindJSON(&req) != nil || !h.consumeSMSCode(tenant, req.Mobile, req.Code, 21) {
		httpx.Fail(c, http.StatusUnauthorized, 401, "验证码无效或已过期")
		return
	}
	var user AdminUser
	if h.service.db.Where("tenant_id = ? AND mobile = ? AND status = 0", tenant, req.Mobile).First(&user).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "手机号未绑定后台账号")
		return
	}
	token, _ := h.service.issueTokenPair(user)
	httpx.OK(c, token)
}

// ResetPassword godoc
// @Summary Reset an operations-console password with an SMS code
// @Tags System Auth
// @Accept json
// @Produce json
// @Param tenant-id header int true "Tenant ID"
// @Success 200 {object} httpx.Response
// @Router /system/auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req struct {
		Mobile   string `json:"mobile" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	tenant := requestTenantID(c, h.service, "")
	if tenant == 0 || c.ShouldBindJSON(&req) != nil || len(req.Password) < 4 || !h.consumeSMSCode(tenant, req.Mobile, req.Code, 23) {
		httpx.Fail(c, http.StatusUnauthorized, 401, "验证码无效或请求参数错误")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	result := h.service.db.Model(&AdminUser{}).Where("tenant_id = ? AND mobile = ? AND status = 0", tenant, req.Mobile).Update("password_hash", string(hash))
	if result.RowsAffected == 0 {
		httpx.Fail(c, http.StatusNotFound, 404, "手机号未绑定后台账号")
		return
	}
	httpx.OK(c, true)
}

// SocialLogin godoc
// @Summary Log in with a previously bound social account
// @Description Provider code exchange is delegated to the configured social adapter; the resulting social record must already be bound to an admin user.
// @Tags System Auth
// @Accept json
// @Produce json
// @Param tenant-id header int true "Tenant ID"
// @Success 200 {object} httpx.Response
// @Router /system/auth/social-login [post]
func (h *Handler) SocialLogin(c *gin.Context) {
	var req struct {
		Type  any    `json:"type" binding:"required"`
		Code  string `json:"code" binding:"required"`
		State string `json:"state"`
	}
	tenant := requestTenantID(c, h.service, "")
	if tenant == 0 || c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	typeValue, _ := strconv.Atoi(fmt.Sprint(req.Type))
	var social SocialUser
	query := h.service.db.Where("tenant_id = ? AND type = ? AND code = ?", tenant, typeValue, req.Code)
	if req.State != "" {
		query = query.Where("state = ?", req.State)
	}
	if query.First(&social).Error != nil {
		httpx.Fail(c, http.StatusUnauthorized, 401, "社交账号尚未绑定")
		return
	}
	var bind SocialUserBind
	if h.service.db.Where("social_user_id = ? AND user_type = 2", social.ID).First(&bind).Error != nil {
		httpx.Fail(c, http.StatusUnauthorized, 401, "社交账号尚未绑定")
		return
	}
	var user AdminUser
	if h.service.db.Where("id = ? AND tenant_id = ? AND status = 0", bind.UserID, tenant).First(&user).Error != nil {
		httpx.Fail(c, http.StatusUnauthorized, 401, "绑定用户不存在")
		return
	}
	token, _ := h.service.issueTokenPair(user)
	httpx.OK(c, token)
}
