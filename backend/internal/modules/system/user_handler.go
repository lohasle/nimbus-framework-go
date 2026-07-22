package system

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"golang.org/x/crypto/bcrypt"
)

func tenantIDFromContext(c *gin.Context) uint64 {
	value, ok := c.Get("tenant_id")
	if !ok {
		return 0
	}
	id, _ := value.(uint64)
	return id
}

func queryID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	return id
}

func pageParams(c *gin.Context) (int, int) {
	pageNo, _ := strconv.Atoi(c.DefaultQuery("pageNo", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	return pageNo, pageSize
}

// UserPage godoc
// @Summary Page operations-console users
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/page [get]
func (h *Handler) UserPage(c *gin.Context) {
	tenantID := tenantIDFromContext(c)
	query := h.service.db.Model(&AdminUser{}).Where("tenant_id = ?", tenantID)
	if username := strings.TrimSpace(c.Query("username")); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if mobile := strings.TrimSpace(c.Query("mobile")); mobile != "" {
		query = query.Where("mobile LIKE ?", "%"+mobile+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if deptID := c.Query("deptId"); deptID != "" {
		query = query.Where("dept_id = ?", deptID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "查询用户失败")
		return
	}
	pageNo, pageSize := pageParams(c)
	var users []AdminUser
	if err := query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "查询用户失败")
		return
	}
	httpx.OK(c, gin.H{"list": h.userViews(tenantID, users), "total": total})
}

// UserList godoc
// @Summary List operations-console users by IDs
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Param ids query string true "Comma-separated user IDs"
// @Success 200 {object} httpx.Response
// @Router /system/user/list [get]
func (h *Handler) UserList(c *gin.Context) {
	ids := splitIDs(c.Query("ids"))
	if len(ids) == 0 {
		httpx.OK(c, []UserView{})
		return
	}
	var users []AdminUser
	tenantID := tenantIDFromContext(c)
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantID, ids).Order("id").Find(&users)
	httpx.OK(c, h.userViews(tenantID, users))
}

// SimpleUsers godoc
// @Summary List enabled operations-console users
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/simple-list [get]
func (h *Handler) SimpleUsers(c *gin.Context) {
	var users []AdminUser
	tenantID := tenantIDFromContext(c)
	h.service.db.Where("tenant_id = ? AND status = 0", tenantID).Order("id").Find(&users)
	httpx.OK(c, h.userViews(tenantID, users))
}

// SimpleUser godoc
// @Summary Get a compact user card
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/get-simple [get]
func (h *Handler) SimpleUser(c *gin.Context) { h.UserGet(c) }

// UsersByNickname godoc
// @Summary Search users by nickname
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/list-by-nickname [get]
func (h *Handler) UsersByNickname(c *gin.Context) {
	var users []AdminUser
	tenantID := tenantIDFromContext(c)
	h.service.db.Where("tenant_id = ? AND nickname LIKE ?", tenantID, "%"+strings.TrimSpace(c.Query("nickname"))+"%").Limit(20).Find(&users)
	httpx.OK(c, h.userViews(tenantID, users))
}

// UserGet godoc
// @Summary Get an operations-console user
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/get [get]
func (h *Handler) UserGet(c *gin.Context) {
	var user AdminUser
	tenantID := tenantIDFromContext(c)
	if err := h.service.db.Where("tenant_id = ? AND id = ?", tenantID, queryID(c)).First(&user).Error; err != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "用户不存在")
		return
	}
	httpx.OK(c, h.userViews(tenantID, []AdminUser{user})[0])
}

// UserCreate godoc
// @Summary Create an operations-console user
// @Tags System User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserSaveRequest true "User"
// @Success 200 {object} httpx.Response
// @Router /system/user/create [post]
func (h *Handler) UserCreate(c *gin.Context) {
	var req UserSaveRequest
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httpx.Fail(c, http.StatusInternalServerError, 500, "密码处理失败")
		return
	}
	user := AdminUser{
		TenantID: tenantIDFromContext(c), Username: req.Username,
		PasswordHash: string(hash), LoginDate: time.Now(),
	}
	applyUserSave(&user, req)
	if err = h.service.db.Create(&user).Error; err != nil {
		httpx.Fail(c, http.StatusConflict, 409, "创建用户失败，请检查用户名是否重复")
		return
	}
	h.replaceUserAssignments(user.ID, req.PostIDs, req.RoleIDs)
	httpx.OK(c, user.ID)
}

// UserUpdate godoc
// @Summary Update an operations-console user
// @Tags System User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserSaveRequest true "User"
// @Success 200 {object} httpx.Response
// @Router /system/user/update [put]
func (h *Handler) UserUpdate(c *gin.Context) {
	var req UserSaveRequest
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	var user AdminUser
	if err := h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&user).Error; err != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "用户不存在")
		return
	}
	applyUserSave(&user, req)
	h.service.db.Save(&user)
	h.replaceUserAssignments(user.ID, req.PostIDs, req.RoleIDs)
	httpx.OK(c, true)
}

// UserDelete godoc
// @Summary Delete an operations-console user
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/delete [delete]
func (h *Handler) UserDelete(c *gin.Context) {
	id := queryID(c)
	tx := h.service.db.Begin()
	tx.Where("user_id = ?", id).Delete(&UserRole{})
	tx.Where("user_id = ?", id).Delete(&UserPost{})
	tx.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), id).Delete(&AdminUser{})
	tx.Commit()
	httpx.OK(c, true)
}

// UserDeleteList godoc
// @Summary Delete operations-console users in batch
// @Tags System User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/delete-list [delete]
func (h *Handler) UserDeleteList(c *gin.Context) {
	ids := splitIDs(c.Query("ids"))
	tx := h.service.db.Begin()
	tx.Where("user_id IN ?", ids).Delete(&UserRole{})
	tx.Where("user_id IN ?", ids).Delete(&UserPost{})
	tx.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), ids).Delete(&AdminUser{})
	tx.Commit()
	httpx.OK(c, true)
}

// UserPassword godoc
// @Summary Reset a user password
// @Tags System User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/update-password [put]
func (h *Handler) UserPassword(c *gin.Context) {
	var req struct {
		ID       uint64 `json:"id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	h.service.db.Model(&AdminUser{}).Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).Update("password_hash", string(hash))
	httpx.OK(c, true)
}

// UserStatus godoc
// @Summary Change a user status
// @Tags System User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/user/update-status [put]
func (h *Handler) UserStatus(c *gin.Context) {
	var req struct {
		ID     uint64 `json:"id" binding:"required"`
		Status int    `json:"status"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	h.service.db.Model(&AdminUser{}).Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).Update("status", req.Status)
	httpx.OK(c, true)
}

// SimpleDepartments godoc
// @Summary List enabled departments
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dept/simple-list [get]
func (h *Handler) SimpleDepartments(c *gin.Context) {
	var rows []Department
	h.service.db.Where("tenant_id = ? AND status = 0", tenantIDFromContext(c)).Order("sort,id").Find(&rows)
	httpx.OK(c, rows)
}

// SimplePosts godoc
// @Summary List enabled posts
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/post/simple-list [get]
func (h *Handler) SimplePosts(c *gin.Context) {
	var rows []Post
	h.service.db.Where("tenant_id = ? AND status = 0", tenantIDFromContext(c)).Order("sort,id").Find(&rows)
	httpx.OK(c, rows)
}

func (h *Handler) userViews(tenantID uint64, users []AdminUser) []UserView {
	var departments []Department
	h.service.db.Where("tenant_id = ?", tenantID).Find(&departments)
	deptNames := make(map[uint64]string, len(departments))
	for _, dept := range departments {
		deptNames[dept.ID] = dept.Name
	}
	views := make([]UserView, 0, len(users))
	for _, user := range users {
		var postIDs, roleIDs []uint64
		h.service.db.Model(&UserPost{}).Where("user_id = ?", user.ID).Order("post_id").Pluck("post_id", &postIDs)
		h.service.db.Model(&UserRole{}).Where("user_id = ?", user.ID).Order("role_id").Pluck("role_id", &roleIDs)
		views = append(views, UserView{AdminUser: user, DeptName: deptNames[user.DeptID], PostIDs: postIDs, RoleIDs: roleIDs})
	}
	return views
}

func (h *Handler) replaceUserAssignments(userID uint64, postIDs, roleIDs []uint64) {
	tx := h.service.db.Begin()
	tx.Where("user_id = ?", userID).Delete(&UserPost{})
	for _, postID := range postIDs {
		tx.Create(&UserPost{UserID: userID, PostID: postID})
	}
	if roleIDs != nil {
		tx.Where("user_id = ?", userID).Delete(&UserRole{})
		for _, roleID := range roleIDs {
			tx.Create(&UserRole{UserID: userID, RoleID: roleID})
		}
	}
	tx.Commit()
}

func splitIDs(raw string) []uint64 {
	parts := strings.Split(raw, ",")
	ids := make([]uint64, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func applyUserSave(user *AdminUser, req UserSaveRequest) {
	user.Nickname = req.Nickname
	user.DeptID = req.DeptID
	user.Mobile = req.Mobile
	user.Email = req.Email
	user.Sex = req.Sex
	user.Remark = req.Remark
	user.Status = req.Status
}
