package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/excelx"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"gorm.io/gorm"
)

// SMSChannelPage godoc
// @Summary Page SMS channels
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-channel/page [get]
func (h *Handler) SMSChannelPage(c *gin.Context) {
	query := h.service.db.Model(&SMSChannel{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if code := strings.TrimSpace(c.Query("code")); code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []SMSChannel
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// SMSChannelSimpleList godoc
// @Summary List enabled SMS channels
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-channel/simple-list [get]
func (h *Handler) SMSChannelSimpleList(c *gin.Context) {
	var rows []SMSChannel
	h.service.db.Where("tenant_id = ? AND status = 0", tenantIDFromContext(c)).Order("code,id").Find(&rows)
	httpx.OK(c, rows)
}

// SMSChannelGet godoc
// @Summary Get an SMS channel
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-channel/get [get]
func (h *Handler) SMSChannelGet(c *gin.Context) {
	var row SMSChannel
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "短信渠道不存在")
		return
	}
	httpx.OK(c, row)
}

// SMSChannelCreate godoc
// @Summary Create an SMS channel
// @Tags System SMS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SMSChannel true "SMS channel"
// @Success 200 {object} httpx.Response
// @Router /system/sms-channel/create [post]
func (h *Handler) SMSChannelCreate(c *gin.Context) {
	var row SMSChannel
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Code) == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantIDFromContext(c)
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "创建短信渠道失败")
		return
	}
	httpx.OK(c, row.ID)
}

// SMSChannelUpdate godoc
// @Summary Update an SMS channel
// @Tags System SMS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SMSChannel true "SMS channel"
// @Success 200 {object} httpx.Response
// @Router /system/sms-channel/update [put]
func (h *Handler) SMSChannelUpdate(c *gin.Context) {
	var req SMSChannel
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row SMSChannel
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "短信渠道不存在")
		return
	}
	created := row.CreatedAt
	row = req
	row.TenantID = tenantIDFromContext(c)
	row.CreatedAt = created
	h.service.db.Save(&row)
	httpx.OK(c, true)
}

// SMSChannelDelete godoc
// @Summary Delete an SMS channel
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-channel/delete [delete]
func (h *Handler) SMSChannelDelete(c *gin.Context) { h.deleteSMSChannels(c, []uint64{queryID(c)}) }

// SMSChannelDeleteList godoc
// @Summary Delete SMS channels in batch
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-channel/delete-list [delete]
func (h *Handler) SMSChannelDeleteList(c *gin.Context) {
	h.deleteSMSChannels(c, splitIDs(c.Query("ids")))
}
func (h *Handler) deleteSMSChannels(c *gin.Context, ids []uint64) {
	var count int64
	h.service.db.Model(&SMSTemplate{}).Where("tenant_id = ? AND channel_id IN ?", tenantIDFromContext(c), ids).Count(&count)
	if count > 0 {
		httpx.Fail(c, 400, 400, "短信渠道已被模板使用，不能删除")
		return
	}
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), ids).Delete(&SMSChannel{})
	httpx.OK(c, true)
}

type SMSTemplateView struct {
	SMSTemplate
	ChannelCode string   `json:"channelCode"`
	Params      []string `json:"params"`
}

func (h *Handler) smsTemplateViews(rows []SMSTemplate) []SMSTemplateView {
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ChannelID)
	}
	var channels []SMSChannel
	h.service.db.Where("id IN ?", ids).Find(&channels)
	codes := map[uint64]string{}
	for _, channel := range channels {
		codes[channel.ID] = channel.Code
	}
	views := make([]SMSTemplateView, 0, len(rows))
	for _, row := range rows {
		views = append(views, SMSTemplateView{SMSTemplate: row, ChannelCode: codes[row.ChannelID], Params: csvList(row.Params)})
	}
	return views
}

// SMSTemplatePage godoc
// @Summary Page SMS templates
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-template/page [get]
func (h *Handler) SMSTemplatePage(c *gin.Context) {
	query := h.service.db.Model(&SMSTemplate{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if code := strings.TrimSpace(c.Query("code")); code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if channel := c.Query("channelId"); channel != "" {
		query = query.Where("channel_id = ?", channel)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []SMSTemplate
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": h.smsTemplateViews(rows), "total": total})
}

// SMSTemplateSimpleList godoc
// @Summary List enabled SMS templates
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-template/simple-list [get]
func (h *Handler) SMSTemplateSimpleList(c *gin.Context) {
	var rows []SMSTemplate
	h.service.db.Where("tenant_id = ? AND status = 0", tenantIDFromContext(c)).Order("name,id").Find(&rows)
	httpx.OK(c, h.smsTemplateViews(rows))
}

// SMSTemplateGet godoc
// @Summary Get an SMS template
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-template/get [get]
func (h *Handler) SMSTemplateGet(c *gin.Context) {
	var row SMSTemplate
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "短信模板不存在")
		return
	}
	httpx.OK(c, h.smsTemplateViews([]SMSTemplate{row})[0])
}

type SMSTemplateRequest struct {
	ID            uint64   `json:"id"`
	Type          int      `json:"type"`
	Status        int      `json:"status"`
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	Content       string   `json:"content"`
	Remark        string   `json:"remark"`
	APITemplateID string   `json:"apiTemplateId"`
	ChannelID     uint64   `json:"channelId"`
	Params        []string `json:"params"`
}

func applySMSTemplate(row *SMSTemplate, req SMSTemplateRequest) {
	row.Type, row.Status, row.Code, row.Name, row.Content, row.Remark, row.APITemplateID, row.ChannelID, row.Params = req.Type, req.Status, strings.TrimSpace(req.Code), req.Name, req.Content, req.Remark, req.APITemplateID, req.ChannelID, strings.Join(req.Params, ",")
}

// SMSTemplateCreate godoc
// @Summary Create an SMS template
// @Tags System SMS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SMSTemplateRequest true "SMS template"
// @Success 200 {object} httpx.Response
// @Router /system/sms-template/create [post]
func (h *Handler) SMSTemplateCreate(c *gin.Context) {
	var req SMSTemplateRequest
	if c.ShouldBindJSON(&req) != nil || req.Code == "" || req.Name == "" || req.ChannelID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row := SMSTemplate{TenantID: tenantIDFromContext(c)}
	applySMSTemplate(&row, req)
	if h.service.db.Where("tenant_id = ? AND id = ?", row.TenantID, row.ChannelID).First(&SMSChannel{}).Error != nil {
		httpx.Fail(c, 400, 400, "短信渠道不存在")
		return
	}
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 409, 409, "短信模板编码已存在")
		return
	}
	httpx.OK(c, row.ID)
}

// SMSTemplateUpdate godoc
// @Summary Update an SMS template
// @Tags System SMS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SMSTemplateRequest true "SMS template"
// @Success 200 {object} httpx.Response
// @Router /system/sms-template/update [put]
func (h *Handler) SMSTemplateUpdate(c *gin.Context) {
	var req SMSTemplateRequest
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row SMSTemplate
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "短信模板不存在")
		return
	}
	applySMSTemplate(&row, req)
	if h.service.db.Save(&row).Error != nil {
		httpx.Fail(c, 409, 409, "短信模板编码已存在")
		return
	}
	httpx.OK(c, true)
}

// SMSTemplateDelete godoc
// @Summary Delete an SMS template
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-template/delete [delete]
func (h *Handler) SMSTemplateDelete(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).Delete(&SMSTemplate{})
	httpx.OK(c, true)
}

// SMSTemplateDeleteList godoc
// @Summary Delete SMS templates in batch
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-template/delete-list [delete]
func (h *Handler) SMSTemplateDeleteList(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), splitIDs(c.Query("ids"))).Delete(&SMSTemplate{})
	httpx.OK(c, true)
}

type SMSSendRequest struct {
	Mobile         string         `json:"mobile" binding:"required"`
	TemplateCode   string         `json:"templateCode" binding:"required"`
	TemplateParams map[string]any `json:"templateParams"`
}

func sendSMSCallback(channel SMSChannel, payload any) (string, error) {
	if strings.TrimSpace(channel.CallbackURL) == "" {
		return "", fmt.Errorf("短信渠道未配置发送回调地址")
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, channel.CallbackURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", channel.APIKey)
	req.Header.Set("X-API-Secret", channel.APISecret)
	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("短信网关返回 HTTP %d", resp.StatusCode)
	}
	return resp.Header.Get("X-Request-ID"), nil
}

// SMSTemplateSend godoc
// @Summary Send an SMS from a template
// @Tags System SMS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SMSSendRequest true "SMS"
// @Success 200 {object} httpx.Response
// @Router /system/sms-template/send-sms [post]
func (h *Handler) SMSTemplateSend(c *gin.Context) {
	var req SMSSendRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	tenant := tenantIDFromContext(c)
	var tmpl SMSTemplate
	if h.service.db.Where("tenant_id = ? AND code = ? AND status = 0", tenant, req.TemplateCode).First(&tmpl).Error != nil {
		httpx.Fail(c, 404, 404, "短信模板不存在或已停用")
		return
	}
	var channel SMSChannel
	if h.service.db.Where("tenant_id = ? AND id = ? AND status = 0", tenant, tmpl.ChannelID).First(&channel).Error != nil {
		httpx.Fail(c, 404, 404, "短信渠道不存在或已停用")
		return
	}
	params, _ := json.Marshal(req.TemplateParams)
	now := time.Now()
	log := SMSLog{TenantID: tenant, ChannelID: channel.ID, ChannelCode: channel.Code, TemplateID: tmpl.ID, TemplateCode: tmpl.Code, TemplateType: tmpl.Type, TemplateContent: renderTemplate(tmpl.Content, req.TemplateParams), TemplateParams: string(params), APITemplateID: tmpl.APITemplateID, Mobile: req.Mobile, UserID: c.GetUint64("user_id"), UserType: 2, SendStatus: 0, SendTime: &now}
	h.service.db.Create(&log)
	requestID, err := sendSMSCallback(channel, gin.H{"mobile": req.Mobile, "signature": channel.Signature, "templateId": tmpl.APITemplateID, "content": log.TemplateContent, "params": req.TemplateParams})
	if err != nil {
		h.service.db.Model(&log).Updates(map[string]any{"send_status": 2, "api_send_msg": err.Error()})
		httpx.Fail(c, 502, 502, "短信发送失败: "+err.Error())
		return
	}
	h.service.db.Model(&log).Updates(map[string]any{"send_status": 1, "api_request_id": requestID, "api_send_code": "OK"})
	httpx.OK(c, log.ID)
}

func (h *Handler) smsLogQuery(c *gin.Context) *gorm.DB {
	query := h.service.db.Model(&SMSLog{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if mobile := strings.TrimSpace(c.Query("mobile")); mobile != "" {
		query = query.Where("mobile LIKE ?", "%"+mobile+"%")
	}
	if code := c.Query("templateCode"); code != "" {
		query = query.Where("template_code = ?", code)
	}
	if status := c.Query("sendStatus"); status != "" {
		query = query.Where("send_status = ?", status)
	}
	return query
}

// SMSLogPage godoc
// @Summary Page SMS logs
// @Tags System SMS
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/sms-log/page [get]
func (h *Handler) SMSLogPage(c *gin.Context) {
	query := h.smsLogQuery(c)
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []SMSLog
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// SMSLogExport godoc
// @Summary Export SMS logs
// @Tags System SMS
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/sms-log/export-excel [get]
func (h *Handler) SMSLogExport(c *gin.Context) {
	var rows []SMSLog
	h.smsLogQuery(c).Order("id DESC").Limit(10000).Find(&rows)
	book := newSystemBook("短信日志", []any{"编号", "渠道", "模板", "手机号", "内容", "发送状态", "发送时间", "网关消息"})
	for i, row := range rows {
		sendTime := ""
		if row.SendTime != nil {
			sendTime = row.SendTime.Format(time.DateTime)
		}
		systemWriteRow(book, "短信日志", i+2, []any{row.ID, row.ChannelCode, row.TemplateCode, row.Mobile, row.TemplateContent, row.SendStatus, sendTime, row.APISendMsg})
	}
	excelx.Write(c, book, "短信日志.xlsx")
}

// SMSTemplateExport godoc
// @Summary Export SMS templates
// @Tags System SMS
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/sms-template/export-excel [get]
func (h *Handler) SMSTemplateExport(c *gin.Context) {
	var rows []SMSTemplate
	h.service.db.Where("tenant_id = ?", tenantIDFromContext(c)).Order("id").Find(&rows)
	book := newSystemBook("短信模板", []any{"编号", "编码", "名称", "类型", "状态", "渠道编号", "API模板编号", "内容"})
	for i, row := range rows {
		systemWriteRow(book, "短信模板", i+2, []any{row.ID, row.Code, row.Name, row.Type, row.Status, row.ChannelID, row.APITemplateID, row.Content})
	}
	excelx.Write(c, book, "短信模板.xlsx")
}

// SocialClientPage godoc
// @Summary Page social login clients
// @Tags System Social
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/social-client/page [get]
func (h *Handler) SocialClientPage(c *gin.Context) {
	query := h.service.db.Model(&SocialClient{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if socialType := c.Query("socialType"); socialType != "" {
		query = query.Where("social_type = ?", socialType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []SocialClient
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// SocialClientGet godoc
// @Summary Get a social login client
// @Tags System Social
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/social-client/get [get]
func (h *Handler) SocialClientGet(c *gin.Context) {
	var row SocialClient
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "社交客户端不存在")
		return
	}
	httpx.OK(c, row)
}

// SocialClientCreate godoc
// @Summary Create a social login client
// @Tags System Social
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SocialClient true "Social client"
// @Success 200 {object} httpx.Response
// @Router /system/social-client/create [post]
func (h *Handler) SocialClientCreate(c *gin.Context) {
	var row SocialClient
	if c.ShouldBindJSON(&row) != nil || row.SocialType == 0 || row.UserType == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantIDFromContext(c)
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "创建社交客户端失败")
		return
	}
	httpx.OK(c, row.ID)
}

// SocialClientUpdate godoc
// @Summary Update a social login client
// @Tags System Social
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SocialClient true "Social client"
// @Success 200 {object} httpx.Response
// @Router /system/social-client/update [put]
func (h *Handler) SocialClientUpdate(c *gin.Context) {
	var req SocialClient
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row SocialClient
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "社交客户端不存在")
		return
	}
	created := row.CreatedAt
	row = req
	row.TenantID = tenantIDFromContext(c)
	row.CreatedAt = created
	h.service.db.Save(&row)
	httpx.OK(c, true)
}

// SocialClientDelete godoc
// @Summary Delete a social login client
// @Tags System Social
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/social-client/delete [delete]
func (h *Handler) SocialClientDelete(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).Delete(&SocialClient{})
	httpx.OK(c, true)
}

// SocialUserPage godoc
// @Summary Page social users
// @Tags System Social
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/social-user/page [get]
func (h *Handler) SocialUserPage(c *gin.Context) {
	query := h.service.db.Model(&SocialUser{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if typ := c.Query("type"); typ != "" {
		query = query.Where("type = ?", typ)
	}
	if nickname := strings.TrimSpace(c.Query("nickname")); nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []SocialUser
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// SocialUserGet godoc
// @Summary Get a social user
// @Tags System Social
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/social-user/get [get]
func (h *Handler) SocialUserGet(c *gin.Context) {
	var row SocialUser
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "社交用户不存在")
		return
	}
	httpx.OK(c, row)
}

// SocialUserBindList godoc
// @Summary List current user's bound social accounts
// @Tags System Social
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/social-user/get-bind-list [get]
func (h *Handler) SocialUserBindList(c *gin.Context) {
	var ids []uint64
	h.service.db.Model(&SocialUserBind{}).Where("user_id = ? AND user_type = 2", c.GetUint64("user_id")).Pluck("social_user_id", &ids)
	var rows []SocialUser
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), ids).Find(&rows)
	httpx.OK(c, rows)
}

// SocialAuthRedirect godoc
// @Summary Build a social provider authorization redirect
// @Tags System Social
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /system/auth/social-auth-redirect [get]
func (h *Handler) SocialAuthRedirect(c *gin.Context) {
	typ := queryUint64(c, "type")
	tenant, _ := strconv.ParseUint(c.GetHeader("tenant-id"), 10, 64)
	var client SocialClient
	if h.service.db.Where("tenant_id = ? AND social_type = ? AND user_type = 2 AND status = 0", tenant, typ).First(&client).Error != nil {
		httpx.Fail(c, 404, 404, "社交登录客户端未配置")
		return
	}
	redirect := url.QueryEscape(c.Query("redirectUri"))
	var target string
	switch typ {
	case 10:
		target = "https://github.com/login/oauth/authorize?client_id=" + url.QueryEscape(client.ClientID) + "&redirect_uri=" + redirect + "&state=" + url.QueryEscape(randomOAuthToken())
	case 30:
		target = "https://open.weixin.qq.com/connect/qrconnect?appid=" + url.QueryEscape(client.ClientID) + "&redirect_uri=" + redirect + "&response_type=code&scope=snsapi_login&state=" + url.QueryEscape(randomOAuthToken()) + "#wechat_redirect"
	default:
		httpx.Fail(c, 400, 400, "该社交类型尚未配置标准授权地址")
		return
	}
	httpx.OK(c, target)
}

// SocialUserBind godoc
// @Summary Bind a social authorization result to current user
// @Description Stores a provider authorization result. Provider token exchange is expected to be completed by an integration adapter.
// @Tags System Social
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/social-user/bind [post]
func (h *Handler) SocialUserBind(c *gin.Context) {
	var req struct {
		Type     int    `json:"type" binding:"required"`
		Code     string `json:"code" binding:"required"`
		State    string `json:"state"`
		OpenID   string `json:"openid"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	if req.OpenID == "" {
		req.OpenID = req.Code
	}
	tenant := tenantIDFromContext(c)
	row := SocialUser{TenantID: tenant, Type: req.Type, OpenID: req.OpenID, Code: req.Code, State: req.State, Nickname: req.Nickname, Avatar: req.Avatar}
	h.service.db.Where("tenant_id = ? AND type = ? AND open_id = ?", tenant, row.Type, row.OpenID).Assign(row).FirstOrCreate(&row)
	bind := SocialUserBind{UserID: c.GetUint64("user_id"), UserType: 2, SocialUserID: row.ID}
	h.service.db.Where("user_id = ? AND user_type = ? AND social_user_id = ?", bind.UserID, bind.UserType, bind.SocialUserID).FirstOrCreate(&bind)
	httpx.OK(c, true)
}

// SocialUserUnbind godoc
// @Summary Unbind a social account from current user
// @Tags System Social
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/social-user/unbind [delete]
func (h *Handler) SocialUserUnbind(c *gin.Context) {
	var req struct {
		Type   int    `json:"type"`
		OpenID string `json:"openid"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var social SocialUser
	if h.service.db.Where("tenant_id = ? AND type = ? AND open_id = ?", tenantIDFromContext(c), req.Type, req.OpenID).First(&social).Error != nil {
		httpx.OK(c, true)
		return
	}
	h.service.db.Where("user_id = ? AND user_type = 2 AND social_user_id = ?", c.GetUint64("user_id"), social.ID).Delete(&SocialUserBind{})
	httpx.OK(c, true)
}
