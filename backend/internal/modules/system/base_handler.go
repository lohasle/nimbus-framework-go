package system

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
)

// MenuSimpleList godoc
// @Summary List all menus for permission assignment
// @Tags System Menu
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/menu/simple-list [get]
func (h *Handler) MenuSimpleList(c *gin.Context) {
	var rows []SystemMenu
	query := h.service.db.Where("status = 0")
	if ids := h.allowedMenuIDs(tenantIDFromContext(c)); len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}
	query.Order("sort,id").Find(&rows)
	httpx.OK(c, rows)
}

// MenuList godoc
// @Summary List menus
// @Tags System Menu
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/menu/list [get]
func (h *Handler) MenuList(c *gin.Context) {
	query := h.service.db.Model(&SystemMenu{})
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var rows []SystemMenu
	query.Order("sort,id").Find(&rows)
	httpx.OK(c, rows)
}

// MenuGet godoc
// @Summary Get a menu
// @Tags System Menu
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/menu/get [get]
func (h *Handler) MenuGet(c *gin.Context) {
	var row SystemMenu
	if h.service.db.First(&row, queryID(c)).Error != nil {
		httpx.Fail(c, 404, 404, "菜单不存在")
		return
	}
	httpx.OK(c, row)
}

func normalizeMenu(row *SystemMenu) {
	row.Name, row.Permission, row.Path = strings.TrimSpace(row.Name), strings.TrimSpace(row.Permission), strings.TrimSpace(row.Path)
	row.Component, row.ComponentName, row.Icon = strings.TrimSpace(row.Component), strings.TrimSpace(row.ComponentName), strings.TrimSpace(row.Icon)
}

// MenuCreate godoc
// @Summary Create a menu
// @Tags System Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SystemMenu true "Menu"
// @Success 200 {object} httpx.Response
// @Router /system/menu/create [post]
func (h *Handler) MenuCreate(c *gin.Context) {
	var row SystemMenu
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Name) == "" || row.Type < 1 || row.Type > 3 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, 0
	normalizeMenu(&row)
	if err := h.service.db.Create(&row).Error; err != nil {
		httpx.Fail(c, 409, 409, "创建菜单失败")
		return
	}
	httpx.OK(c, row.ID)
}

// MenuUpdate godoc
// @Summary Update a menu
// @Tags System Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SystemMenu true "Menu"
// @Success 200 {object} httpx.Response
// @Router /system/menu/update [put]
func (h *Handler) MenuUpdate(c *gin.Context) {
	var req SystemMenu
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 || req.ID == req.ParentID {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row SystemMenu
	if h.service.db.First(&row, req.ID).Error != nil {
		httpx.Fail(c, 404, 404, "菜单不存在")
		return
	}
	created := row.CreatedAt
	row = req
	row.TenantID = 0
	row.CreatedAt = created
	normalizeMenu(&row)
	if err := h.service.db.Save(&row).Error; err != nil {
		httpx.Fail(c, 409, 409, "更新菜单失败")
		return
	}
	httpx.OK(c, true)
}

// MenuDelete godoc
// @Summary Delete a menu
// @Tags System Menu
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/menu/delete [delete]
func (h *Handler) MenuDelete(c *gin.Context) {
	id := queryID(c)
	var count int64
	h.service.db.Model(&SystemMenu{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		httpx.Fail(c, 400, 400, "请先删除子菜单")
		return
	}
	tx := h.service.db.Begin()
	tx.Where("menu_id = ?", id).Delete(&RoleMenu{})
	tx.Delete(&SystemMenu{}, id)
	tx.Commit()
	httpx.OK(c, true)
}

// DeptList godoc
// @Summary List departments
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dept/list [get]
func (h *Handler) DeptList(c *gin.Context) {
	query := h.service.db.Where("tenant_id = ?", tenantIDFromContext(c))
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var rows []Department
	query.Order("sort,id").Find(&rows)
	httpx.OK(c, rows)
}

// DeptGet godoc
// @Summary Get a department
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dept/get [get]
func (h *Handler) DeptGet(c *gin.Context) {
	var row Department
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "部门不存在")
		return
	}
	httpx.OK(c, row)
}

// DeptCreate godoc
// @Summary Create a department
// @Tags System Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Department true "Department"
// @Success 200 {object} httpx.Response
// @Router /system/dept/create [post]
func (h *Handler) DeptCreate(c *gin.Context) {
	var row Department
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Name) == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantIDFromContext(c)
	if err := h.service.db.Create(&row).Error; err != nil {
		httpx.Fail(c, 500, 500, "创建部门失败")
		return
	}
	httpx.OK(c, row.ID)
}

// DeptUpdate godoc
// @Summary Update a department
// @Tags System Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Department true "Department"
// @Success 200 {object} httpx.Response
// @Router /system/dept/update [put]
func (h *Handler) DeptUpdate(c *gin.Context) {
	var req Department
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 || req.ID == req.ParentID {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row Department
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "部门不存在")
		return
	}
	row.Name, row.ParentID, row.Sort, row.LeaderUserID, row.Phone, row.Email, row.Status = req.Name, req.ParentID, req.Sort, req.LeaderUserID, req.Phone, req.Email, req.Status
	h.service.db.Save(&row)
	httpx.OK(c, true)
}

// DeptDelete godoc
// @Summary Delete a department
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dept/delete [delete]
func (h *Handler) DeptDelete(c *gin.Context) { h.deleteDepartments(c, []uint64{queryID(c)}) }

// DeptDeleteList godoc
// @Summary Delete departments in batch
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dept/delete-list [delete]
func (h *Handler) DeptDeleteList(c *gin.Context) { h.deleteDepartments(c, splitIDs(c.Query("ids"))) }

func (h *Handler) deleteDepartments(c *gin.Context, ids []uint64) {
	tenant := tenantIDFromContext(c)
	var children, users int64
	h.service.db.Model(&Department{}).Where("tenant_id = ? AND parent_id IN ?", tenant, ids).Count(&children)
	h.service.db.Model(&AdminUser{}).Where("tenant_id = ? AND dept_id IN ?", tenant, ids).Count(&users)
	if children > 0 || users > 0 {
		httpx.Fail(c, 400, 400, "部门存在子部门或用户，不能删除")
		return
	}
	h.service.db.Where("tenant_id = ? AND id IN ?", tenant, ids).Delete(&Department{})
	httpx.OK(c, true)
}

// PostPage godoc
// @Summary Page posts
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/post/page [get]
func (h *Handler) PostPage(c *gin.Context) {
	query := h.service.db.Model(&Post{}).Where("tenant_id = ?", tenantIDFromContext(c))
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
	var rows []Post
	query.Order("sort,id").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// PostGet godoc
// @Summary Get a post
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/post/get [get]
func (h *Handler) PostGet(c *gin.Context) {
	var row Post
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), queryID(c)).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "岗位不存在")
		return
	}
	httpx.OK(c, row)
}

// PostCreate godoc
// @Summary Create a post
// @Tags System Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Post true "Post"
// @Success 200 {object} httpx.Response
// @Router /system/post/create [post]
func (h *Handler) PostCreate(c *gin.Context) {
	var row Post
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Name) == "" || strings.TrimSpace(row.Code) == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID, row.TenantID = 0, tenantIDFromContext(c)
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 409, 409, "岗位编码已存在")
		return
	}
	httpx.OK(c, row.ID)
}

// PostUpdate godoc
// @Summary Update a post
// @Tags System Organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Post true "Post"
// @Success 200 {object} httpx.Response
// @Router /system/post/update [put]
func (h *Handler) PostUpdate(c *gin.Context) {
	var req Post
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row Post
	if h.service.db.Where("tenant_id = ? AND id = ?", tenantIDFromContext(c), req.ID).First(&row).Error != nil {
		httpx.Fail(c, 404, 404, "岗位不存在")
		return
	}
	row.Name, row.Code, row.Sort, row.Status, row.Remark = req.Name, req.Code, req.Sort, req.Status, req.Remark
	if h.service.db.Save(&row).Error != nil {
		httpx.Fail(c, 409, 409, "岗位编码已存在")
		return
	}
	httpx.OK(c, true)
}

// PostDelete godoc
// @Summary Delete a post
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/post/delete [delete]
func (h *Handler) PostDelete(c *gin.Context) { h.deletePosts(c, []uint64{queryID(c)}) }

// PostDeleteList godoc
// @Summary Delete posts in batch
// @Tags System Organization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/post/delete-list [delete]
func (h *Handler) PostDeleteList(c *gin.Context) { h.deletePosts(c, splitIDs(c.Query("ids"))) }

func (h *Handler) deletePosts(c *gin.Context, ids []uint64) {
	var count int64
	h.service.db.Model(&UserPost{}).Where("post_id IN ?", ids).Count(&count)
	if count > 0 {
		httpx.Fail(c, http.StatusBadRequest, 400, "岗位已分配用户，不能删除")
		return
	}
	h.service.db.Where("tenant_id = ? AND id IN ?", tenantIDFromContext(c), ids).Delete(&Post{})
	httpx.OK(c, true)
}
