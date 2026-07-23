package system

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"golang.org/x/crypto/bcrypt"
)

// NoticePage godoc
// @Summary Page notices
// @Tags System Notice
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notice/page [get]
func (h *Handler) NoticePage(c *gin.Context) {
	query := h.service.db.Model(&Notice{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if title := strings.TrimSpace(c.Query("title")); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if typ := c.Query("type"); typ != "" {
		query = query.Where("type = ?", typ)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []Notice
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// NoticeGet godoc
// @Summary Get a notice
// @Tags System Notice
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notice/get [get]
func (h *Handler) NoticeGet(c *gin.Context) {
	var row Notice
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "公告不存在")
		return
	}
	httpx.OK(c, row)
}

// NoticeCreate godoc
// @Summary Create a notice
// @Tags System Notice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Notice true "Notice"
// @Success 200 {object} httpx.Response
// @Router /system/notice/create [post]
func (h *Handler) NoticeCreate(c *gin.Context) {
	var row Notice
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Title) == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID, row.Creator = 0, tenantIDFromContext(c), currentUsername(h, c)
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "创建公告失败")
		return
	}
	httpx.OK(c, row.ID)
}

// NoticeUpdate godoc
// @Summary Update a notice
// @Tags System Notice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Notice true "Notice"
// @Success 200 {object} httpx.Response
// @Router /system/notice/update [put]
func (h *Handler) NoticeUpdate(c *gin.Context) {
	var req Notice
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row Notice
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "公告不存在")
		return
	}
	row.Title, row.Type, row.Content, row.Status, row.Remark = req.Title, req.Type, req.Content, req.Status, req.Remark
	h.service.db.Save(&row)
	httpx.OK(c, true)
}

// NoticeDelete godoc
// @Summary Delete a notice
// @Tags System Notice
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notice/delete [delete]
func (h *Handler) NoticeDelete(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).Delete(&Notice{})
	httpx.OK(c, true)
}

// NoticeDeleteList godoc
// @Summary Delete notices in batch
// @Tags System Notice
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notice/delete-list [delete]
func (h *Handler) NoticeDeleteList(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), splitIDs(c.Query("ids"))).Delete(&Notice{})
	httpx.OK(c, true)
}

// NoticePush godoc
// @Summary Push a notice to enabled tenant users
// @Tags System Notice
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notice/push [post]
func (h *Handler) NoticePush(c *gin.Context) {
	tenant := tenantIDFromContext(c)
	var notice Notice
	if h.service.db.Where("tenant_id = ? AND id = ?", tenant, queryID(c)).First(&notice).Error != nil {
		httpx.Fail(c, 404, 404, "公告不存在")
		return
	}
	var users []AdminUser
	h.service.db.Where("tenant_id = ? AND status = 0", tenant).Find(&users)
	tx := h.service.db.Begin()
	for _, user := range users {
		tx.Create(&NotifyMessage{TenantID: tenant, UserID: user.ID, UserType: 2, TemplateCode: "notice", TemplateNickname: notice.Title, TemplateContent: notice.Content, TemplateType: notice.Type, TemplateParams: "{}"})
	}
	now := time.Now()
	tx.Model(&notice).Updates(map[string]any{"pushed_at": now, "status": 0})
	if tx.Commit().Error != nil {
		httpx.Fail(c, 500, 500, "推送公告失败")
		return
	}
	httpx.OK(c, true)
}

func currentUsername(h *Handler, c *gin.Context) string {
	var user AdminUser
	h.service.db.Select("username").First(&user, c.GetUint64("user_id"))
	return user.Username
}

// NotifyTemplateSimpleList godoc
// @Summary List enabled notification templates
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-template/simple-list [get]
func (h *Handler) NotifyTemplateSimpleList(c *gin.Context) {
	var rows []NotifyTemplate
	h.service.db.Where("tenant_id = ? AND status = 0", tenantIDFromContext(c)).Order("name,id").Find(&rows)
	httpx.OK(c, rows)
}

// NotifyTemplatePage godoc
// @Summary Page notification templates
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-template/page [get]
func (h *Handler) NotifyTemplatePage(c *gin.Context) {
	query := h.service.db.Model(&NotifyTemplate{}).Where("tenant_id = ?", tenantIDFromContext(c))
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
	var rows []NotifyTemplate
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// NotifyTemplateGet godoc
// @Summary Get a notification template
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-template/get [get]
func (h *Handler) NotifyTemplateGet(c *gin.Context) {
	var row NotifyTemplate
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "站内信模板不存在")
		return
	}
	httpx.OK(c, row)
}

// NotifyTemplateCreate godoc
// @Summary Create a notification template
// @Tags System Notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body NotifyTemplate true "Notification template"
// @Success 200 {object} httpx.Response
// @Router /system/notify-template/create [post]
func (h *Handler) NotifyTemplateCreate(c *gin.Context) {
	var row NotifyTemplate
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Name) == "" || strings.TrimSpace(row.Code) == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantIDFromContext(c)
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 409, 409, "模板编码已存在")
		return
	}
	httpx.OK(c, row.ID)
}

// NotifyTemplateUpdate godoc
// @Summary Update a notification template
// @Tags System Notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body NotifyTemplate true "Notification template"
// @Success 200 {object} httpx.Response
// @Router /system/notify-template/update [put]
func (h *Handler) NotifyTemplateUpdate(c *gin.Context) {
	var req NotifyTemplate
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row NotifyTemplate
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "站内信模板不存在")
		return
	}
	row.Name, row.Nickname, row.Code, row.Content, row.Type, row.Params, row.Status, row.Remark = req.Name, req.Nickname, req.Code, req.Content, req.Type, req.Params, req.Status, req.Remark
	if h.service.db.Save(&row).Error != nil {
		httpx.Fail(c, 409, 409, "模板编码已存在")
		return
	}
	httpx.OK(c, true)
}

// NotifyTemplateDelete godoc
// @Summary Delete a notification template
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-template/delete [delete]
func (h *Handler) NotifyTemplateDelete(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).Delete(&NotifyTemplate{})
	httpx.OK(c, true)
}

// NotifyTemplateDeleteList godoc
// @Summary Delete notification templates in batch
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-template/delete-list [delete]
func (h *Handler) NotifyTemplateDeleteList(c *gin.Context) {
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), splitIDs(c.Query("ids"))).Delete(&NotifyTemplate{})
	httpx.OK(c, true)
}

type NotifySendRequest struct {
	UserID         uint64         `json:"userId" binding:"required"`
	TemplateCode   string         `json:"templateCode" binding:"required"`
	TemplateParams map[string]any `json:"templateParams"`
}

func renderTemplate(content string, params map[string]any) string {
	for key, value := range params {
		text, _ := json.Marshal(value)
		replacement := strings.Trim(string(text), "\"")
		content = strings.ReplaceAll(content, "{{"+key+"}}", replacement)
		content = strings.ReplaceAll(content, "{"+key+"}", replacement)
	}
	return content
}

// NotifyTemplateSend godoc
// @Summary Send a notification from a template
// @Tags System Notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body NotifySendRequest true "Notification"
// @Success 200 {object} httpx.Response
// @Router /system/notify-template/send-notify [post]
func (h *Handler) NotifyTemplateSend(c *gin.Context) {
	var req NotifySendRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	tenant := tenantIDFromContext(c)
	var tmpl NotifyTemplate
	if h.service.db.Where("tenant_id = ? AND code = ? AND status = 0", tenant, req.TemplateCode).First(&tmpl).Error != nil {
		httpx.Fail(c, 404, 404, "站内信模板不存在或已停用")
		return
	}
	if h.service.db.Where("tenant_id = ? AND id = ?", tenant, req.UserID).First(&AdminUser{}).Error != nil {
		httpx.Fail(c, 404, 404, "接收用户不存在")
		return
	}
	params, _ := json.Marshal(req.TemplateParams)
	row := NotifyMessage{TenantID: tenant, UserID: req.UserID, UserType: 2, TemplateID: tmpl.ID, TemplateCode: tmpl.Code, TemplateNickname: tmpl.Nickname, TemplateContent: renderTemplate(tmpl.Content, req.TemplateParams), TemplateType: tmpl.Type, TemplateParams: string(params)}
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "发送站内信失败")
		return
	}
	httpx.OK(c, row.ID)
}

func (h *Handler) notifyMessagePage(c *gin.Context, mine bool) {
	query := h.service.db.Model(&NotifyMessage{}).Where("tenant_id = ?", tenantIDFromContext(c))
	if mine {
		query = query.Where("user_id = ?", c.GetUint64("user_id"))
	} else if user := c.Query("userId"); user != "" {
		query = query.Where("user_id = ?", user)
	}
	if read := c.Query("readStatus"); read != "" {
		query = query.Where("read_status = ?", read)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []NotifyMessage
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// NotifyMessagePage godoc
// @Summary Page all notification messages
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-message/page [get]
func (h *Handler) NotifyMessagePage(c *gin.Context) { h.notifyMessagePage(c, false) }

// MyNotifyMessagePage godoc
// @Summary Page current user's notification messages
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-message/my-page [get]
func (h *Handler) MyNotifyMessagePage(c *gin.Context) { h.notifyMessagePage(c, true) }

// NotifyMessageRead godoc
// @Summary Mark selected notification messages read
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-message/update-read [put]
func (h *Handler) NotifyMessageRead(c *gin.Context) {
	ids := c.QueryArray("ids")
	if len(ids) == 0 {
		ids = strings.Split(c.Query("ids"), ",")
	}
	now := time.Now()
	h.service.db.Model(&NotifyMessage{}).Where("tenant_id = ? AND user_id = ? AND id IN ?", tenantIDFromContext(c), c.GetUint64("user_id"), ids).Updates(map[string]any{"read_status": true, "read_time": now})
	httpx.OK(c, true)
}

// NotifyMessageReadAll godoc
// @Summary Mark all current user's notification messages read
// @Tags System Notification
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/notify-message/update-all-read [put]
func (h *Handler) NotifyMessageReadAll(c *gin.Context) {
	now := time.Now()
	h.service.db.Model(&NotifyMessage{}).Where("tenant_id = ? AND user_id = ? AND read_status = ?", tenantIDFromContext(c), c.GetUint64("user_id"), false).Updates(map[string]any{"read_status": true, "read_time": now})
	httpx.OK(c, true)
}

// UserProfileGet godoc
// @Summary Get current user's profile
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/profile/get [get]
func (h *Handler) UserProfileGet(c *gin.Context) {
	tenant, userID := tenantIDFromContext(c), c.GetUint64("user_id")
	var user AdminUser
	if h.service.db.Where("tenant_id = ? AND id = ?", tenant, userID).First(&user).Error != nil {
		httpx.Fail(c, 404, 404, "用户不存在")
		return
	}
	var dept Department
	h.service.db.Where("tenant_id = ? AND id = ?", tenant, user.DeptID).First(&dept)
	var roleIDs, postIDs []uint64
	h.service.db.Model(&UserRole{}).Where("user_id = ?", userID).Pluck("role_id", &roleIDs)
	h.service.db.Model(&UserPost{}).Where("user_id = ?", userID).Pluck("post_id", &postIDs)
	var roles []Role
	var posts []Post
	h.service.db.Where("tenant_id = ? AND id IN ?", tenant, roleIDs).Find(&roles)
	h.service.db.Where("tenant_id = ? AND id IN ?", tenant, postIDs).Find(&posts)
	httpx.OK(c, gin.H{"id": user.ID, "username": user.Username, "nickname": user.Nickname, "dept": dept, "roles": roles, "posts": posts, "email": user.Email, "mobile": user.Mobile, "sex": user.Sex, "avatar": user.Avatar, "status": user.Status, "remark": user.Remark, "loginIp": user.LoginIP, "loginDate": user.LoginDate, "createTime": user.CreatedAt})
}

// UserProfileUpdate godoc
// @Summary Update current user's profile
// @Tags System User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/profile/update [put]
func (h *Handler) UserProfileUpdate(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Mobile   string `json:"mobile"`
		Sex      int    `json:"sex"`
		Avatar   string `json:"avatar"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	if strings.TrimSpace(req.Nickname) == "" {
		httpx.Fail(c, 400, 400, "昵称不能为空")
		return
	}
	h.service.db.Model(&AdminUser{}).Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), c.GetUint64("user_id")).Updates(map[string]any{"nickname": req.Nickname, "email": req.Email, "mobile": req.Mobile, "sex": req.Sex, "avatar": req.Avatar})
	httpx.OK(c, true)
}

// UserProfilePassword godoc
// @Summary Change current user's password
// @Tags System User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/profile/update-password [put]
func (h *Handler) UserProfilePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil || len(req.NewPassword) < 6 {
		httpx.Fail(c, http.StatusBadRequest, 400, "新密码至少 6 位")
		return
	}
	var user AdminUser
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), c.GetUint64("user_id")).First(&user).Error != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)) != nil {
		httpx.Fail(c, 400, 400, "原密码错误")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	h.service.db.Model(&user).Update("password_hash", string(hash))
	httpx.OK(c, true)
}
