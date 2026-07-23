package system

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/excelx"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"gorm.io/gorm"
)

// MailAccountPage godoc
// @Summary Page mail accounts
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-account/page [get]
func (h *Handler) MailAccountPage(c *gin.Context) {
	query := h.service.db.Model(&MailAccount{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if mail := strings.TrimSpace(c.Query("mail")); mail != "" {
		query = query.Where("mail LIKE ?", "%"+mail+"%")
	}
	if username := strings.TrimSpace(c.Query("username")); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []MailAccount
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// MailAccountSimpleList godoc
// @Summary List mail accounts
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-account/simple-list [get]
func (h *Handler) MailAccountSimpleList(c *gin.Context) {
	var rows []MailAccount
	h.service.db.Where("tenant_id = ?", tenantIDFromContext(c)).Order("mail,id").Find(&rows)
	httpx.OK(c, rows)
}

// MailAccountGet godoc
// @Summary Get a mail account
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-account/get [get]
func (h *Handler) MailAccountGet(c *gin.Context) {
	var row MailAccount
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "邮箱账号不存在")
		return
	}
	httpx.OK(c, row)
}

func validateMailAccount(row MailAccount) bool {
	return strings.TrimSpace(row.Mail) != "" && strings.TrimSpace(row.Host) != "" && row.Port > 0
}

// MailAccountCreate godoc
// @Summary Create a mail account
// @Tags System Mail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MailAccount true "Mail account"
// @Success 200 {object} httpx.Response
// @Router /system/mail-account/create [post]
func (h *Handler) MailAccountCreate(c *gin.Context) {
	var row MailAccount
	if c.ShouldBindJSON(&row) != nil || !validateMailAccount(row) {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantIDFromContext(c)
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "创建邮箱账号失败")
		return
	}
	httpx.OK(c, row.ID)
}

// MailAccountUpdate godoc
// @Summary Update a mail account
// @Tags System Mail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MailAccount true "Mail account"
// @Success 200 {object} httpx.Response
// @Router /system/mail-account/update [put]
func (h *Handler) MailAccountUpdate(c *gin.Context) {
	var req MailAccount
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 || !validateMailAccount(req) {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row MailAccount
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "邮箱账号不存在")
		return
	}
	created := row.CreatedAt
	row = req
	row.TenantID = tenantIDFromContext(c)
	row.CreatedAt = created
	h.service.db.Save(&row)
	httpx.OK(c, true)
}

// MailAccountDelete godoc
// @Summary Delete a mail account
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-account/delete [delete]
func (h *Handler) MailAccountDelete(c *gin.Context) { h.deleteMailAccounts(c, []uint64{queryID(c)}) }

// MailAccountDeleteList godoc
// @Summary Delete mail accounts in batch
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-account/delete-list [delete]
func (h *Handler) MailAccountDeleteList(c *gin.Context) {
	h.deleteMailAccounts(c, splitIDs(c.Query("ids")))
}
func (h *Handler) deleteMailAccounts(c *gin.Context, ids []uint64) {
	var count int64
	h.service.db.Model(&MailTemplate{}).Where("tenant_id = ? AND account_id IN ?", tenantIDFromContext(c), ids).Count(&count)
	if count > 0 {
		httpx.Fail(c, 400, 400, "邮箱账号已被模板使用，不能删除")
		return
	}
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), ids).Delete(&MailAccount{})
	httpx.OK(c, true)
}

// MailTemplatePage godoc
// @Summary Page mail templates
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-template/page [get]
func (h *Handler) MailTemplatePage(c *gin.Context) {
	query := h.service.db.Model(&MailTemplate{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if code := strings.TrimSpace(c.Query("code")); code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []MailTemplate
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// MailTemplateSimpleList godoc
// @Summary List enabled mail templates
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-template/simple-list [get]
func (h *Handler) MailTemplateSimpleList(c *gin.Context) {
	var rows []MailTemplate
	h.service.db.Where("tenant_id = ? AND status = 0", tenantIDFromContext(c)).Order("name,id").Find(&rows)
	httpx.OK(c, rows)
}

// MailTemplateGet godoc
// @Summary Get a mail template
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-template/get [get]
func (h *Handler) MailTemplateGet(c *gin.Context) {
	var row MailTemplate
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "邮件模板不存在")
		return
	}
	httpx.OK(c, row)
}

func validateMailTemplate(row MailTemplate) bool {
	return strings.TrimSpace(row.Name) != "" && strings.TrimSpace(row.Code) != "" && row.AccountID > 0
}

// MailTemplateCreate godoc
// @Summary Create a mail template
// @Tags System Mail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MailTemplate true "Mail template"
// @Success 200 {object} httpx.Response
// @Router /system/mail-template/create [post]
func (h *Handler) MailTemplateCreate(c *gin.Context) {
	var row MailTemplate
	if c.ShouldBindJSON(&row) != nil || !validateMailTemplate(row) {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantIDFromContext(c)
	if h.service.db.Where("tenant_id = ? AND id = ?", row.TenantID, row.AccountID).First(&MailAccount{}).Error != nil {
		httpx.Fail(c, 400, 400, "邮箱账号不存在")
		return
	}
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 409, 409, "邮件模板编码已存在")
		return
	}
	httpx.OK(c, row.ID)
}

// MailTemplateUpdate godoc
// @Summary Update a mail template
// @Tags System Mail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MailTemplate true "Mail template"
// @Success 200 {object} httpx.Response
// @Router /system/mail-template/update [put]
func (h *Handler) MailTemplateUpdate(c *gin.Context) {
	var req MailTemplate
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 || !validateMailTemplate(req) {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row MailTemplate
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "邮件模板不存在")
		return
	}
	created := row.CreatedAt
	row = req
	row.TenantID = tenantIDFromContext(c)
	row.CreatedAt = created
	if h.service.db.Save(&row).Error != nil {
		httpx.Fail(c, 409, 409, "邮件模板编码已存在")
		return
	}
	httpx.OK(c, true)
}

// MailTemplateDelete godoc
// @Summary Delete a mail template
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-template/delete [delete]
func (h *Handler) MailTemplateDelete(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).Delete(&MailTemplate{})
	httpx.OK(c, true)
}

// MailTemplateDeleteList godoc
// @Summary Delete mail templates in batch
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-template/delete-list [delete]
func (h *Handler) MailTemplateDeleteList(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), splitIDs(c.Query("ids"))).Delete(&MailTemplate{})
	httpx.OK(c, true)
}

type MailSendRequest struct {
	ToMails        []string       `json:"toMails" binding:"required"`
	CCMails        []string       `json:"ccMails"`
	BCCMails       []string       `json:"bccMails"`
	TemplateCode   string         `json:"templateCode" binding:"required"`
	TemplateParams map[string]any `json:"templateParams"`
}
type MailLogView struct {
	MailLog
	ToMails  []string `json:"toMails"`
	CCMails  []string `json:"ccMails"`
	BCCMails []string `json:"bccMails"`
}

func mailLogView(row MailLog) MailLogView {
	return MailLogView{MailLog: row, ToMails: csvList(row.ToMails), CCMails: csvList(row.CCMails), BCCMails: csvList(row.BCCMails)}
}

func sendSMTP(account MailAccount, to, cc, bcc []string, subject, html string) error {
	recipients := append(append(append([]string{}, to...), cc...), bcc...)
	if len(recipients) == 0 {
		return fmt.Errorf("收件人不能为空")
	}
	address := net.JoinHostPort(account.Host, fmt.Sprint(account.Port))
	auth := smtp.PlainAuth("", account.Username, account.Password, account.Host)
	headers := []string{"From: " + account.Mail, "To: " + strings.Join(to, ","), "Subject: " + subject, "MIME-Version: 1.0", "Content-Type: text/html; charset=UTF-8"}
	if len(cc) > 0 {
		headers = append(headers, "Cc: "+strings.Join(cc, ","))
	}
	message := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + html)
	if !account.SSLEnable {
		return smtp.SendMail(address, auth, account.Mail, recipients, message)
	}
	connection, err := tls.Dial("tcp", address, &tls.Config{ServerName: account.Host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	defer connection.Close()
	client, err := smtp.NewClient(connection, account.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if account.Username != "" {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}
	if err = client.Mail(account.Mail); err != nil {
		return err
	}
	for _, recipient := range recipients {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = writer.Write(message); err != nil {
		return err
	}
	return writer.Close()
}

// MailTemplateSend godoc
// @Summary Send an email from a template
// @Tags System Mail
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MailSendRequest true "Mail"
// @Success 200 {object} httpx.Response
// @Router /system/mail-template/send-mail [post]
func (h *Handler) MailTemplateSend(c *gin.Context) {
	var req MailSendRequest
	if c.ShouldBindJSON(&req) != nil || len(req.ToMails) == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	tenant := tenantIDFromContext(c)
	var tmpl MailTemplate
	if h.service.db.Where("tenant_id = ? AND code = ? AND status = 0", tenant, req.TemplateCode).First(&tmpl).Error != nil {
		httpx.Fail(c, 404, 404, "邮件模板不存在或已停用")
		return
	}
	var account MailAccount
	if h.service.db.Where("tenant_id = ? AND id = ?", tenant, tmpl.AccountID).First(&account).Error != nil {
		httpx.Fail(c, 404, 404, "邮箱账号不存在")
		return
	}
	params, _ := json.Marshal(req.TemplateParams)
	log := MailLog{TenantID: tenant, UserID: c.GetUint64("user_id"), UserType: 2, ToMails: strings.Join(req.ToMails, ","), CCMails: strings.Join(req.CCMails, ","), BCCMails: strings.Join(req.BCCMails, ","), AccountID: account.ID, FromMail: account.Mail, TemplateID: tmpl.ID, TemplateCode: tmpl.Code, TemplateNickname: tmpl.Nickname, TemplateTitle: renderTemplate(tmpl.Title, req.TemplateParams), TemplateContent: renderTemplate(tmpl.Content, req.TemplateParams), TemplateParams: string(params), SendStatus: 0}
	h.service.db.Create(&log)
	err := sendSMTP(account, req.ToMails, req.CCMails, req.BCCMails, log.TemplateTitle, log.TemplateContent)
	now := time.Now()
	if err != nil {
		h.service.db.Model(&log).Updates(map[string]any{"send_status": 2, "send_time": now, "send_exception": err.Error()})
		httpx.Fail(c, 502, 502, "邮件发送失败: "+err.Error())
		return
	}
	h.service.db.Model(&log).Updates(map[string]any{"send_status": 1, "send_time": now, "send_message_id": fmt.Sprintf("mail-%d", log.ID)})
	httpx.OK(c, log.ID)
}

func (h *Handler) mailLogQuery(c *gin.Context) *gorm.DB {
	query := h.service.db.Model(&MailLog{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if account := c.Query("accountId"); account != "" {
		query = query.Where("account_id = ?", account)
	}
	if code := c.Query("templateCode"); code != "" {
		query = query.Where("template_code = ?", code)
	}
	if status := c.Query("sendStatus"); status != "" {
		query = query.Where("send_status = ?", status)
	}
	return query
}

// MailLogPage godoc
// @Summary Page mail logs
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-log/page [get]
func (h *Handler) MailLogPage(c *gin.Context) {
	query := h.mailLogQuery(c)
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []MailLog
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	views := make([]MailLogView, 0, len(rows))
	for _, row := range rows {
		views = append(views, mailLogView(row))
	}
	httpx.OK(c, gin.H{"list": views, "total": total})
}

// MailLogGet godoc
// @Summary Get a mail log
// @Tags System Mail
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/mail-log/get [get]
func (h *Handler) MailLogGet(c *gin.Context) {
	var row MailLog
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "邮件日志不存在")
		return
	}
	httpx.OK(c, mailLogView(row))
}

// MailLogExport godoc
// @Summary Export mail logs
// @Tags System Mail
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} file
// @Router /system/mail-log/export-excel [get]
func (h *Handler) MailLogExport(c *gin.Context) {
	var rows []MailLog
	h.mailLogQuery(c).Order("id DESC").Limit(10000).Find(&rows)
	book := newSystemBook("邮件日志", []any{"编号", "发件邮箱", "收件邮箱", "模板编码", "标题", "状态", "发送时间", "异常"})
	for i, row := range rows {
		sendTime := ""
		if row.SendTime != nil {
			sendTime = row.SendTime.Format(time.DateTime)
		}
		systemWriteRow(book, "邮件日志", i+2, []any{row.ID, row.FromMail, row.ToMails, row.TemplateCode, row.TemplateTitle, row.SendStatus, sendTime, row.SendException})
	}
	excelx.Write(c, book, "邮件日志.xlsx")
}
