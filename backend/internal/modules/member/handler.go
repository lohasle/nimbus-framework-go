package member

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"gorm.io/gorm"
)

type Handler struct{ db *gorm.DB }

func NewHandler(db *gorm.DB) *Handler { return &Handler{db: db} }

func tenantID(c *gin.Context) uint64 { return c.GetUint64("tenant_id") }

func queryID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	return id
}

func page(c *gin.Context) (int, int) {
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
// @Summary Page APP members
// @Tags Member User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/user/page [get]
func (h *Handler) UserPage(c *gin.Context) {
	query := h.db.Model(&User{}).Where("tenant_id = ?", tenantID(c))
	if mobile := strings.TrimSpace(c.Query("mobile")); mobile != "" {
		query = query.Where("mobile LIKE ?", "%"+mobile+"%")
	}
	if nickname := strings.TrimSpace(c.Query("nickname")); nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if levelID := c.Query("levelId"); levelID != "" {
		query = query.Where("level_id = ?", levelID)
	}
	if groupID := c.Query("groupId"); groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}
	var total int64
	query.Count(&total)
	pageNo, pageSize := page(c)
	var rows []User
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// UserGet godoc
// @Summary Get an APP member
// @Tags Member User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/user/get [get]
func (h *Handler) UserGet(c *gin.Context) {
	var row User
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "会员不存在")
		return
	}
	var tagIDs []uint64
	h.db.Model(&UserTag{}).Where("user_id = ?", row.ID).Pluck("tag_id", &tagIDs)
	httpx.OK(c, gin.H{"id": row.ID, "tenantId": row.TenantID, "mobile": row.Mobile, "nickname": row.Nickname, "avatar": row.Avatar, "sex": row.Sex, "status": row.Status, "levelId": row.LevelID, "groupId": row.GroupID, "tagIds": tagIDs, "point": row.Point, "experience": row.Experience, "balance": row.Balance, "remark": row.Remark, "createTime": row.CreatedAt})
}

// UserCreate godoc
// @Summary Create an APP member
// @Tags Member User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserSaveRequest true "Member"
// @Success 200 {object} httpx.Response
// @Router /member/user/create [post]
func (h *Handler) UserCreate(c *gin.Context) {
	var req UserSaveRequest
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	row := User{TenantID: tenantID(c), RegisterIP: c.ClientIP()}
	applyUser(&row, req)
	if err := h.db.Create(&row).Error; err != nil {
		httpx.Fail(c, http.StatusConflict, 409, "手机号已存在")
		return
	}
	h.replaceTags(row.ID, req.TagIDs)
	httpx.OK(c, row.ID)
}

// UserUpdate godoc
// @Summary Update an APP member
// @Tags Member User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserSaveRequest true "Member"
// @Success 200 {object} httpx.Response
// @Router /member/user/update [put]
func (h *Handler) UserUpdate(c *gin.Context) {
	var req UserSaveRequest
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	var row User
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, "会员不存在")
		return
	}
	applyUser(&row, req)
	h.db.Save(&row)
	h.replaceTags(row.ID, req.TagIDs)
	httpx.OK(c, true)
}

// UserStatus godoc
// @Summary Change an APP member status
// @Tags Member User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/user/update-status [put]
func (h *Handler) UserStatus(c *gin.Context) {
	var req struct {
		ID     uint64 `json:"id" binding:"required"`
		Status int    `json:"status"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	h.db.Model(&User{}).Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).Update("status", req.Status)
	httpx.OK(c, true)
}

// UserDelete godoc
// @Summary Delete an APP member
// @Tags Member User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/user/delete [delete]
func (h *Handler) UserDelete(c *gin.Context) {
	id := queryID(c)
	tx := h.db.Begin()
	tx.Where("user_id = ?", id).Delete(&UserTag{})
	tx.Where("tenant_id = ? AND id = ?", tenantID(c), id).Delete(&User{})
	tx.Commit()
	httpx.OK(c, true)
}

// UserPoint godoc
// @Summary Adjust a member's points
// @Tags Member User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/user/update-point [put]
func (h *Handler) UserPoint(c *gin.Context) {
	var req struct {
		ID          uint64 `json:"id" binding:"required"`
		Point       int64  `json:"point"`
		Description string `json:"description"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	tx := h.db.Begin()
	var row User
	if tx.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).First(&row).Error != nil {
		tx.Rollback()
		httpx.Fail(c, http.StatusNotFound, 404, "会员不存在")
		return
	}
	row.Point += req.Point
	tx.Save(&row)
	tx.Create(&PointRecord{TenantID: tenantID(c), UserID: row.ID, BizType: "admin_adjust", Point: req.Point, TotalPoint: row.Point, Description: req.Description})
	tx.Commit()
	httpx.OK(c, true)
}

// UserLevel godoc
// @Summary Change a member's level
// @Tags Member User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/user/update-level [put]
func (h *Handler) UserLevel(c *gin.Context) {
	var req struct {
		ID      uint64 `json:"id" binding:"required"`
		LevelID uint64 `json:"levelId" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	h.db.Model(&User{}).Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).Update("level_id", req.LevelID)
	httpx.OK(c, true)
}

// LevelPage godoc
// @Summary Page member levels
// @Tags Member Level
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/level/page [get]
func (h *Handler) LevelPage(c *gin.Context) { h.pageLevels(c, false) }

// LevelSimple godoc
// @Summary List enabled member levels
// @Tags Member Level
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/level/simple-list [get]
func (h *Handler) LevelSimple(c *gin.Context) { h.pageLevels(c, true) }

// LevelList godoc
// @Summary List member levels
// @Tags Member Level
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/level/list [get]
func (h *Handler) LevelList(c *gin.Context) {
	var rows []Level
	h.db.Where("tenant_id = ?", tenantID(c)).Order("level,id").Find(&rows)
	httpx.OK(c, rows)
}

// LevelGet godoc
// @Summary Get a member level
// @Tags Member Level
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/level/get [get]
func (h *Handler) LevelGet(c *gin.Context) { h.getEntity(c, &Level{}, "会员等级不存在") }

// LevelCreate godoc
// @Summary Create a member level
// @Tags Member Level
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LevelSaveRequest true "Level"
// @Success 200 {object} httpx.Response
// @Router /member/level/create [post]
func (h *Handler) LevelCreate(c *gin.Context) { h.saveLevel(c, false) }

// LevelUpdate godoc
// @Summary Update a member level
// @Tags Member Level
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LevelSaveRequest true "Level"
// @Success 200 {object} httpx.Response
// @Router /member/level/update [put]
func (h *Handler) LevelUpdate(c *gin.Context) { h.saveLevel(c, true) }

// GroupPage godoc
// @Summary Page member groups
// @Tags Member Group
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/group/page [get]
func (h *Handler) GroupPage(c *gin.Context) { h.pageGroups(c, false) }

// GroupSimple godoc
// @Summary List enabled member groups
// @Tags Member Group
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/group/simple-list [get]
func (h *Handler) GroupSimple(c *gin.Context) { h.pageGroups(c, true) }

// GroupGet godoc
// @Summary Get a member group
// @Tags Member Group
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/group/get [get]
func (h *Handler) GroupGet(c *gin.Context) { h.getEntity(c, &Group{}, "会员分组不存在") }

// GroupCreate godoc
// @Summary Create a member group
// @Tags Member Group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GroupSaveRequest true "Group"
// @Success 200 {object} httpx.Response
// @Router /member/group/create [post]
func (h *Handler) GroupCreate(c *gin.Context) { h.saveGroup(c, false) }

// GroupUpdate godoc
// @Summary Update a member group
// @Tags Member Group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GroupSaveRequest true "Group"
// @Success 200 {object} httpx.Response
// @Router /member/group/update [put]
func (h *Handler) GroupUpdate(c *gin.Context) { h.saveGroup(c, true) }

// TagPage godoc
// @Summary Page member tags
// @Tags Member Tag
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/tag/page [get]
func (h *Handler) TagPage(c *gin.Context) { h.pageTags(c, false) }

// TagSimple godoc
// @Summary List enabled member tags
// @Tags Member Tag
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/tag/simple-list [get]
func (h *Handler) TagSimple(c *gin.Context) { h.pageTags(c, true) }

// TagGet godoc
// @Summary Get a member tag
// @Tags Member Tag
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/tag/get [get]
func (h *Handler) TagGet(c *gin.Context) { h.getEntity(c, &Tag{}, "会员标签不存在") }

// TagCreate godoc
// @Summary Create a member tag
// @Tags Member Tag
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body TagSaveRequest true "Tag"
// @Success 200 {object} httpx.Response
// @Router /member/tag/create [post]
func (h *Handler) TagCreate(c *gin.Context) { h.saveTag(c, false) }

// TagUpdate godoc
// @Summary Update a member tag
// @Tags Member Tag
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body TagSaveRequest true "Tag"
// @Success 200 {object} httpx.Response
// @Router /member/tag/update [put]
func (h *Handler) TagUpdate(c *gin.Context) { h.saveTag(c, true) }

// EntityDelete godoc
// @Summary Delete a member level, group or tag
// @Tags Member Administration
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /member/level/delete [delete]
// @Router /member/group/delete [delete]
// @Router /member/tag/delete [delete]
func (h *Handler) EntityDelete(model any) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).Delete(model)
		httpx.OK(c, true)
	}
}

func (h *Handler) pageLevels(c *gin.Context, simple bool) {
	query := h.db.Where("tenant_id = ?", tenantID(c))
	if simple {
		var rows []Level
		query.Where("status = 0").Order("level,id").Find(&rows)
		httpx.OK(c, rows)
		return
	}
	var total int64
	h.db.Model(&Level{}).Where("tenant_id = ?", tenantID(c)).Count(&total)
	pageNo, pageSize := page(c)
	var rows []Level
	query.Order("level,id").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

func (h *Handler) pageGroups(c *gin.Context, simple bool) {
	query := h.db.Where("tenant_id = ?", tenantID(c))
	var rows []Group
	if simple {
		query.Where("status = 0").Order("id").Find(&rows)
		httpx.OK(c, rows)
		return
	}
	var total int64
	h.db.Model(&Group{}).Where("tenant_id = ?", tenantID(c)).Count(&total)
	pageNo, pageSize := page(c)
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

func (h *Handler) pageTags(c *gin.Context, simple bool) {
	query := h.db.Where("tenant_id = ?", tenantID(c))
	var rows []Tag
	if simple {
		query.Where("status = 0").Order("id").Find(&rows)
		httpx.OK(c, rows)
		return
	}
	var total int64
	h.db.Model(&Tag{}).Where("tenant_id = ?", tenantID(c)).Count(&total)
	pageNo, pageSize := page(c)
	query.Order("id DESC").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

func (h *Handler) saveLevel(c *gin.Context, update bool) {
	var req LevelSaveRequest
	if c.ShouldBindJSON(&req) != nil || (update && req.ID == 0) {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	row := Level{ID: req.ID, TenantID: tenantID(c), Name: req.Name, Level: req.Level, Experience: req.Experience, Discount: req.Discount, Icon: req.Icon, Background: req.Background, Status: req.Status}
	if update {
		h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).Updates(&row)
		httpx.OK(c, true)
		return
	}
	h.db.Create(&row)
	httpx.OK(c, row.ID)
}

func (h *Handler) saveGroup(c *gin.Context, update bool) {
	var req GroupSaveRequest
	if c.ShouldBindJSON(&req) != nil || (update && req.ID == 0) {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	row := Group{ID: req.ID, TenantID: tenantID(c), Name: req.Name, Remark: req.Remark, Status: req.Status}
	if update {
		h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).Updates(&row)
		httpx.OK(c, true)
		return
	}
	h.db.Create(&row)
	httpx.OK(c, row.ID)
}

func (h *Handler) saveTag(c *gin.Context, update bool) {
	var req TagSaveRequest
	if c.ShouldBindJSON(&req) != nil || (update && req.ID == 0) {
		httpx.Fail(c, http.StatusBadRequest, 400, "请求参数错误")
		return
	}
	row := Tag{ID: req.ID, TenantID: tenantID(c), Name: req.Name, Status: req.Status}
	if update {
		h.db.Where("tenant_id = ? AND id = ?", tenantID(c), req.ID).Updates(&row)
		httpx.OK(c, true)
		return
	}
	h.db.Create(&row)
	httpx.OK(c, row.ID)
}

func (h *Handler) replaceTags(userID uint64, tagIDs []uint64) {
	h.db.Where("user_id = ?", userID).Delete(&UserTag{})
	for _, tagID := range tagIDs {
		h.db.Create(&UserTag{UserID: userID, TagID: tagID})
	}
}

func (h *Handler) getEntity(c *gin.Context, model any, notFound string) {
	if h.db.Where("tenant_id = ? AND id = ?", tenantID(c), queryID(c)).First(model).Error != nil {
		httpx.Fail(c, http.StatusNotFound, 404, notFound)
		return
	}
	httpx.OK(c, model)
}

func applyUser(row *User, req UserSaveRequest) {
	row.Mobile, row.Nickname, row.Avatar = req.Mobile, req.Nickname, req.Avatar
	row.Sex, row.Status, row.LevelID, row.GroupID = req.Sex, req.Status, req.LevelID, req.GroupID
	row.Remark = req.Remark
}
