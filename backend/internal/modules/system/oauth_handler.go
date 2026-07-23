package system

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"golang.org/x/crypto/bcrypt"
)

type OAuth2ClientSaveRequest struct {
	ID                          uint64   `json:"id"`
	ClientID                    string   `json:"clientId" binding:"required"`
	Secret                      string   `json:"secret" binding:"required"`
	Name                        string   `json:"name" binding:"required"`
	Logo                        string   `json:"logo"`
	Description                 string   `json:"description"`
	Status                      int      `json:"status"`
	AccessTokenValiditySeconds  int      `json:"accessTokenValiditySeconds"`
	RefreshTokenValiditySeconds int      `json:"refreshTokenValiditySeconds"`
	RedirectURIs                []string `json:"redirectUris"`
	AutoApprove                 bool     `json:"autoApprove"`
	AuthorizedGrantTypes        []string `json:"authorizedGrantTypes"`
	Scopes                      []string `json:"scopes"`
	Authorities                 []string `json:"authorities"`
	ResourceIDs                 []string `json:"resourceIds"`
	AdditionalInformation       string   `json:"additionalInformation"`
}

type OAuth2ClientView struct {
	OAuth2Client
	RedirectURIs                []string `json:"redirectUris"`
	AuthorizedGrantTypes        []string `json:"authorizedGrantTypes"`
	Scopes                      []string `json:"scopes"`
	Authorities                 []string `json:"authorities"`
	ResourceIDs                 []string `json:"resourceIds"`
	IsAdditionalInformationJSON bool     `json:"isAdditionalInformationJson"`
}

func csvList(value string) []string {
	result := []string{}
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func oauthClientView(row OAuth2Client) OAuth2ClientView {
	var object any
	return OAuth2ClientView{OAuth2Client: row, RedirectURIs: csvList(row.RedirectURIs), AuthorizedGrantTypes: csvList(row.AuthorizedGrantTypes), Scopes: csvList(row.Scopes), Authorities: csvList(row.Authorities), ResourceIDs: csvList(row.ResourceIDs), IsAdditionalInformationJSON: json.Unmarshal([]byte(row.AdditionalInformation), &object) == nil}
}

func applyOAuthClient(row *OAuth2Client, req OAuth2ClientSaveRequest) {
	row.ClientID, row.Secret, row.Name, row.Logo, row.Description, row.Status = strings.TrimSpace(req.ClientID), req.Secret, req.Name, req.Logo, req.Description, req.Status
	row.AccessTokenValiditySeconds, row.RefreshTokenValiditySeconds = req.AccessTokenValiditySeconds, req.RefreshTokenValiditySeconds
	if row.AccessTokenValiditySeconds <= 0 {
		row.AccessTokenValiditySeconds = 7200
	}
	if row.RefreshTokenValiditySeconds <= 0 {
		row.RefreshTokenValiditySeconds = 2592000
	}
	row.RedirectURIs, row.AuthorizedGrantTypes, row.Scopes = strings.Join(req.RedirectURIs, ","), strings.Join(req.AuthorizedGrantTypes, ","), strings.Join(req.Scopes, ",")
	row.AutoApprove, row.Authorities, row.ResourceIDs = req.AutoApprove, strings.Join(req.Authorities, ","), strings.Join(req.ResourceIDs, ",")
	row.AdditionalInformation = req.AdditionalInformation
}

// OAuth2ClientPage godoc
// @Summary Page OAuth2 clients
// @Tags System OAuth2
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/oauth2-client/page [get]
func (h *Handler) OAuth2ClientPage(c *gin.Context) {
	query := h.service.db.Model(&OAuth2Client{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if client := strings.TrimSpace(c.Query("clientId")); client != "" {
		query = query.Where("client_id LIKE ?", "%"+client+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	pageNo, pageSize := pageParams(c)
	var rows []OAuth2Client
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	views := make([]OAuth2ClientView, 0, len(rows))
	for _, row := range rows {
		views = append(views, oauthClientView(row))
	}
	httpx.OK(c, gin.H{"list": views, "total": total})
}

// OAuth2ClientGet godoc
// @Summary Get an OAuth2 client
// @Tags System OAuth2
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/oauth2-client/get [get]
func (h *Handler) OAuth2ClientGet(c *gin.Context) {
	var row OAuth2Client
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "OAuth2 客户端不存在")
		return
	}
	httpx.OK(c, oauthClientView(row))
}

// OAuth2ClientCreate godoc
// @Summary Create an OAuth2 client
// @Tags System OAuth2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body OAuth2ClientSaveRequest true "OAuth2 client"
// @Success 200 {object} httpx.Response
// @Router /system/oauth2-client/create [post]
func (h *Handler) OAuth2ClientCreate(c *gin.Context) {
	var req OAuth2ClientSaveRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row := OAuth2Client{TenantID: tenantIDFromContext(c)}
	applyOAuthClient(&row, req)
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 409, 409, "客户端编号已存在")
		return
	}
	httpx.OK(c, row.ID)
}

// OAuth2ClientUpdate godoc
// @Summary Update an OAuth2 client
// @Tags System OAuth2
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body OAuth2ClientSaveRequest true "OAuth2 client"
// @Success 200 {object} httpx.Response
// @Router /system/oauth2-client/update [put]
func (h *Handler) OAuth2ClientUpdate(c *gin.Context) {
	var req OAuth2ClientSaveRequest
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row OAuth2Client
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "OAuth2 客户端不存在")
		return
	}
	applyOAuthClient(&row, req)
	if h.service.db.Save(&row).Error != nil {
		httpx.Fail(c, 409, 409, "客户端编号已存在")
		return
	}
	httpx.OK(c, true)
}

// OAuth2ClientDelete godoc
// @Summary Delete an OAuth2 client
// @Tags System OAuth2
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/oauth2-client/delete [delete]
func (h *Handler) OAuth2ClientDelete(c *gin.Context) { h.deleteOAuthClients(c, []uint64{queryID(c)}) }

// OAuth2ClientDeleteList godoc
// @Summary Delete OAuth2 clients in batch
// @Tags System OAuth2
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/oauth2-client/delete-list [delete]
func (h *Handler) OAuth2ClientDeleteList(c *gin.Context) {
	h.deleteOAuthClients(c, splitIDs(c.Query("ids")))
}

func (h *Handler) deleteOAuthClients(c *gin.Context, ids []uint64) {
	var clients []OAuth2Client
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), ids).Find(&clients)
	clientIDs := make([]string, 0, len(clients))
	for _, row := range clients {
		clientIDs = append(clientIDs, row.ClientID)
	}
	tx := h.service.db.Begin()
	tx.Where("client_id IN ?", clientIDs).Delete(&OAuth2AccessToken{})
	tx.Where("client_id IN ?", clientIDs).Delete(&OAuth2AuthorizationCode{})
	tx.Where("client_id IN ?", clientIDs).Delete(&OAuth2Approve{})
	tx.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), ids).Delete(&OAuth2Client{})
	tx.Commit()
	httpx.OK(c, true)
}

// OAuth2TokenPage godoc
// @Summary Page valid OAuth2 access tokens
// @Tags System OAuth2
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/oauth2-token/page [get]
func (h *Handler) OAuth2TokenPage(c *gin.Context) {
	query := h.service.db.Model(&OAuth2AccessToken{}).Where("tenant_id = ? AND expires_time > ?", tenantIDFromContext(c), time.Now())
	if user := c.Query("userId"); user != "" {
		query = query.Where("user_id = ?", user)
	}
	if client := c.Query("clientId"); client != "" {
		query = query.Where("client_id = ?", client)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []OAuth2AccessToken
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	views := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		views = append(views, gin.H{"id": row.ID, "accessToken": row.AccessToken, "refreshToken": row.RefreshToken, "userId": row.UserID, "userType": row.UserType, "clientId": row.ClientID, "scopes": csvList(row.Scopes), "createTime": row.CreatedAt, "expiresTime": row.ExpiresTime})
	}
	httpx.OK(c, gin.H{"list": views, "total": total})
}

// OAuth2TokenDelete godoc
// @Summary Revoke an OAuth2 access token
// @Tags System OAuth2
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/oauth2-token/delete [delete]
func (h *Handler) OAuth2TokenDelete(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND access_token = ?", tenantIDFromContext(c), c.Query("accessToken")).Delete(&OAuth2AccessToken{})
	httpx.OK(c, true)
}

func basicCredentials(c *gin.Context) (string, string, bool) {
	raw := c.GetHeader("Authorization")
	if !strings.HasPrefix(raw, "Basic ") {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(raw, "Basic "))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func splitScope(value string) []string { return strings.Fields(strings.ReplaceAll(value, ",", " ")) }
func randomOAuthToken() string {
	first, _ := newTokenID()
	second, _ := newTokenID()
	return first + second
}

func (h *Handler) authenticateOAuthClient(c *gin.Context, grant, redirect string, scopes []string) (OAuth2Client, bool) {
	clientID, secret, ok := basicCredentials(c)
	if !ok {
		clientID, secret = c.PostForm("client_id"), c.PostForm("client_secret")
	}
	var client OAuth2Client
	if clientID == "" || h.service.db.Where("client_id = ? AND status = 0", clientID).First(&client).Error != nil || subtle.ConstantTimeCompare([]byte(client.Secret), []byte(secret)) != 1 {
		httpx.Fail(c, http.StatusUnauthorized, 401, "OAuth2 客户端认证失败")
		return client, false
	}
	if grant != "" && !contains(csvList(client.AuthorizedGrantTypes), grant) {
		httpx.Fail(c, 400, 400, "客户端不支持该授权类型")
		return client, false
	}
	if redirect != "" && !contains(csvList(client.RedirectURIs), redirect) {
		httpx.Fail(c, 400, 400, "回调地址不合法")
		return client, false
	}
	allowed := csvList(client.Scopes)
	for _, scope := range scopes {
		if !contains(allowed, scope) {
			httpx.Fail(c, 400, 400, "授权范围不合法")
			return client, false
		}
	}
	return client, true
}
func (h *Handler) issueOAuthToken(client OAuth2Client, userID uint64, userType int, scopes []string) (OAuth2AccessToken, error) {
	now := time.Now()
	row := OAuth2AccessToken{TenantID: client.TenantID, AccessToken: randomOAuthToken(), RefreshToken: randomOAuthToken(), UserID: userID, UserType: userType, ClientID: client.ClientID, Scopes: strings.Join(scopes, ","), CreatedAt: now, ExpiresTime: now.Add(time.Duration(client.AccessTokenValiditySeconds) * time.Second)}
	return row, h.service.db.Create(&row).Error
}
func oauthTokenResponse(row OAuth2AccessToken) gin.H {
	return gin.H{"access_token": row.AccessToken, "refresh_token": row.RefreshToken, "token_type": "Bearer", "expires_in": int64(time.Until(row.ExpiresTime).Seconds()), "scope": row.Scopes, "user_id": row.UserID}
}

// OAuth2OpenToken godoc
// @Summary Issue an OAuth2 access token
// @Description Supports authorization_code, password, client_credentials and refresh_token grants using HTTP Basic client authentication.
// @Tags System OAuth2 Open
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param grant_type formData string true "Grant type"
// @Success 200 {object} httpx.Response
// @Router /system/oauth2/token [post]
func (h *Handler) OAuth2OpenToken(c *gin.Context) {
	grant := c.PostForm("grant_type")
	scopes := splitScope(c.PostForm("scope"))
	client, ok := h.authenticateOAuthClient(c, grant, c.PostForm("redirect_uri"), scopes)
	if !ok {
		return
	}
	var userID uint64
	userType := 2
	switch grant {
	case "client_credentials":
		userType = 0
	case "password":
		var user AdminUser
		if h.service.db.Where("tenant_id = ? AND username = ? AND status = 0", client.TenantID, c.PostForm("username")).First(&user).Error != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(c.PostForm("password"))) != nil {
			httpx.Fail(c, 401, 401, "用户名或密码错误")
			return
		}
		userID = user.ID
	case "refresh_token":
		var old OAuth2AccessToken
		if h.service.db.Where("client_id = ? AND refresh_token = ?", client.ClientID, c.PostForm("refresh_token")).First(&old).Error != nil {
			httpx.Fail(c, 400, 400, "刷新令牌无效")
			return
		}
		userID, userType, scopes = old.UserID, old.UserType, csvList(old.Scopes)
		h.service.db.Delete(&old)
	case "authorization_code":
		var code OAuth2AuthorizationCode
		if h.service.db.Where("client_id = ? AND code = ? AND redirect_uri = ? AND expires_time > ?", client.ClientID, c.PostForm("code"), c.PostForm("redirect_uri"), time.Now()).First(&code).Error != nil {
			httpx.Fail(c, 400, 400, "授权码无效")
			return
		}
		userID, userType, scopes = code.UserID, code.UserType, csvList(code.Scopes)
		h.service.db.Delete(&code)
	default:
		httpx.Fail(c, 400, 400, "未知授权类型")
		return
	}
	row, err := h.issueOAuthToken(client, userID, userType, scopes)
	if err != nil {
		httpx.Fail(c, 500, 500, "签发令牌失败")
		return
	}
	httpx.OK(c, oauthTokenResponse(row))
}

// OAuth2OpenRevoke godoc
// @Summary Revoke an OAuth2 token
// @Tags System OAuth2 Open
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /system/oauth2/token [delete]
func (h *Handler) OAuth2OpenRevoke(c *gin.Context) {
	client, ok := h.authenticateOAuthClient(c, "", "", nil)
	if !ok {
		return
	}
	token := c.Query("token")
	result := h.service.db.Where("client_id = ? AND (access_token = ? OR refresh_token = ?)", client.ClientID, token, token).Delete(&OAuth2AccessToken{})
	httpx.OK(c, result.RowsAffected > 0)
}

// OAuth2OpenCheckToken godoc
// @Summary Introspect an OAuth2 token
// @Tags System OAuth2 Open
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /system/oauth2/check-token [post]
func (h *Handler) OAuth2OpenCheckToken(c *gin.Context) {
	_, ok := h.authenticateOAuthClient(c, "", "", nil)
	if !ok {
		return
	}
	token := c.PostForm("token")
	if token == "" {
		token = c.Query("token")
	}
	var row OAuth2AccessToken
	if h.service.db.Where("access_token = ? AND expires_time > ?", token, time.Now()).First(&row).Error != nil {
		httpx.Fail(c, 400, 400, "访问令牌无效")
		return
	}
	httpx.OK(c, gin.H{"active": true, "userId": row.UserID, "userType": row.UserType, "clientId": row.ClientID, "scopes": csvList(row.Scopes), "expiresTime": row.ExpiresTime})
}

// OAuth2AuthorizeInfo godoc
// @Summary Get OAuth2 authorization information
// @Tags System OAuth2 Open
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/oauth2/authorize [get]
func (h *Handler) OAuth2AuthorizeInfo(c *gin.Context) {
	var client OAuth2Client
	if h.service.db.Where("client_id = ? AND status = 0", c.Query("clientId")).First(&client).Error != nil {
		httpx.Fail(c, 404, 404, "OAuth2 客户端不存在")
		return
	}
	var user AdminUser
	h.service.db.First(&user, c.GetUint64("user_id"))
	var approved []string
	h.service.db.Model(&OAuth2Approve{}).Where("user_id = ? AND client_id = ? AND approved = ? AND expires_at > ?", user.ID, client.ClientID, true, time.Now()).Pluck("scope", &approved)
	httpx.OK(c, gin.H{"client": oauthClientView(client), "scopes": csvList(client.Scopes), "approvedScopes": approved, "user": gin.H{"id": user.ID, "nickname": user.Nickname, "avatar": user.Avatar}})
}

// OAuth2Authorize godoc
// @Summary Approve OAuth2 scopes and issue an authorization result
// @Tags System OAuth2 Open
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/oauth2/authorize [post]
func (h *Handler) OAuth2Authorize(c *gin.Context) {
	clientID, redirect, responseType := c.Query("client_id"), c.Query("redirect_uri"), c.Query("response_type")
	var client OAuth2Client
	if h.service.db.Where("client_id = ? AND status = 0", clientID).First(&client).Error != nil || !contains(csvList(client.RedirectURIs), redirect) {
		httpx.Fail(c, 400, 400, "客户端或回调地址不合法")
		return
	}
	var decisions map[string]bool
	_ = json.Unmarshal([]byte(c.Query("scope")), &decisions)
	approved := []string{}
	tx := h.service.db.Begin()
	for scope, value := range decisions {
		if !contains(csvList(client.Scopes), scope) {
			continue
		}
		row := OAuth2Approve{TenantID: client.TenantID, UserID: c.GetUint64("user_id"), UserType: 2, ClientID: clientID, Scope: scope, Approved: value, ExpiresAt: time.Now().AddDate(1, 0, 0)}
		tx.Where("user_id = ? AND user_type = ? AND client_id = ? AND scope = ?", row.UserID, row.UserType, row.ClientID, row.Scope).Assign(row).FirstOrCreate(&row)
		if value {
			approved = append(approved, scope)
		}
	}
	tx.Commit()
	target, err := url.Parse(redirect)
	if err != nil {
		httpx.Fail(c, 400, 400, "回调地址不合法")
		return
	}
	query := target.Query()
	query.Set("state", c.Query("state"))
	if responseType == "code" {
		code := OAuth2AuthorizationCode{TenantID: client.TenantID, Code: randomOAuthToken(), ClientID: clientID, UserID: c.GetUint64("user_id"), UserType: 2, Scopes: strings.Join(approved, ","), RedirectURI: redirect, State: c.Query("state"), ExpiresTime: time.Now().Add(5 * time.Minute)}
		h.service.db.Create(&code)
		query.Set("code", code.Code)
	} else if responseType == "token" {
		token, issueErr := h.issueOAuthToken(client, c.GetUint64("user_id"), 2, approved)
		if issueErr != nil {
			httpx.Fail(c, 500, 500, "签发令牌失败")
			return
		}
		query.Set("access_token", token.AccessToken)
		query.Set("token_type", "Bearer")
	} else {
		httpx.Fail(c, 400, 400, "不支持的响应类型")
		return
	}
	target.RawQuery = query.Encode()
	httpx.OK(c, target.String())
}
