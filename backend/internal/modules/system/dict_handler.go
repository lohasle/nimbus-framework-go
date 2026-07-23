package system

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
)

// DictTypeSimpleList godoc
// @Summary List enabled dictionary types
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-type/simple-list [get]
func (h *Handler) DictTypeSimpleList(c *gin.Context) {
	var rows []DictType
	h.service.db.Where("status = 0").Order("name,id").Find(&rows)
	httpx.OK(c, rows)
}

// DictTypePage godoc
// @Summary Page dictionary types
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-type/page [get]
func (h *Handler) DictTypePage(c *gin.Context) {
	query := h.service.db.Model(&DictType{})
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if typ := strings.TrimSpace(c.Query("type")); typ != "" {
		query = query.Where("type LIKE ?", "%"+typ+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []DictType
	query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// DictTypeGet godoc
// @Summary Get a dictionary type
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-type/get [get]
func (h *Handler) DictTypeGet(c *gin.Context) {
	var row DictType
	if h.service.db.First(&row, queryID(c)).Error != nil {
		httpx.Fail(c, 404, 404, "字典类型不存在")
		return
	}
	httpx.OK(c, row)
}

// DictTypeCreate godoc
// @Summary Create a dictionary type
// @Tags System Dictionary
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DictType true "Dictionary type"
// @Success 200 {object} httpx.Response
// @Router /system/dict-type/create [post]
func (h *Handler) DictTypeCreate(c *gin.Context) {
	var row DictType
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Name) == "" || strings.TrimSpace(row.Type) == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	row.ID = 0
	row.Type = strings.TrimSpace(row.Type)
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 409, 409, "字典类型已存在")
		return
	}
	httpx.OK(c, row.ID)
}

// DictTypeUpdate godoc
// @Summary Update a dictionary type
// @Tags System Dictionary
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DictType true "Dictionary type"
// @Success 200 {object} httpx.Response
// @Router /system/dict-type/update [put]
func (h *Handler) DictTypeUpdate(c *gin.Context) {
	var req DictType
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row DictType
	if h.service.db.First(&row, req.ID).Error != nil {
		httpx.Fail(c, 404, 404, "字典类型不存在")
		return
	}
	oldType := row.Type
	row.Name, row.Type, row.Status, row.Remark = req.Name, strings.TrimSpace(req.Type), req.Status, req.Remark
	tx := h.service.db.Begin()
	if tx.Save(&row).Error != nil {
		tx.Rollback()
		httpx.Fail(c, 409, 409, "字典类型已存在")
		return
	}
	if oldType != row.Type {
		tx.Model(&DictData{}).Where("dict_type = ?", oldType).Update("dict_type", row.Type)
	}
	tx.Commit()
	httpx.OK(c, true)
}

// DictTypeDelete godoc
// @Summary Delete a dictionary type
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-type/delete [delete]
func (h *Handler) DictTypeDelete(c *gin.Context) { h.deleteDictTypes(c, []uint64{queryID(c)}) }

// DictTypeDeleteList godoc
// @Summary Delete dictionary types in batch
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-type/delete-list [delete]
func (h *Handler) DictTypeDeleteList(c *gin.Context) { h.deleteDictTypes(c, splitIDs(c.Query("ids"))) }

func (h *Handler) deleteDictTypes(c *gin.Context, ids []uint64) {
	var rows []DictType
	h.service.db.Where("id IN ?", ids).Find(&rows)
	types := make([]string, 0, len(rows))
	for _, row := range rows {
		types = append(types, row.Type)
	}
	tx := h.service.db.Begin()
	tx.Where("dict_type IN ?", types).Delete(&DictData{})
	tx.Where("id IN ?", ids).Delete(&DictType{})
	tx.Commit()
	httpx.OK(c, true)
}

// DictDataPage godoc
// @Summary Page dictionary data
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-data/page [get]
func (h *Handler) DictDataPage(c *gin.Context) {
	query := h.service.db.Model(&DictData{})
	if label := strings.TrimSpace(c.Query("label")); label != "" {
		query = query.Where("label LIKE ?", "%"+label+"%")
	}
	if typ := strings.TrimSpace(c.Query("dictType")); typ != "" {
		query = query.Where("dict_type = ?", typ)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	pn, ps := pageParams(c)
	var rows []DictData
	query.Order("sort,id").Offset((pn - 1) * ps).Limit(ps).Find(&rows)
	httpx.OK(c, gin.H{"list": rows, "total": total})
}

// DictDataGet godoc
// @Summary Get dictionary data
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-data/get [get]
func (h *Handler) DictDataGet(c *gin.Context) {
	var row DictData
	if h.service.db.First(&row, queryID(c)).Error != nil {
		httpx.Fail(c, 404, 404, "字典数据不存在")
		return
	}
	httpx.OK(c, row)
}

// DictDataByType godoc
// @Summary List dictionary data by type
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-data/type [get]
func (h *Handler) DictDataByType(c *gin.Context) {
	var rows []DictData
	h.service.db.Where("dict_type = ? AND status = 0", c.Query("type")).Order("sort,id").Find(&rows)
	httpx.OK(c, rows)
}

// DictDataCreate godoc
// @Summary Create dictionary data
// @Tags System Dictionary
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DictData true "Dictionary data"
// @Success 200 {object} httpx.Response
// @Router /system/dict-data/create [post]
func (h *Handler) DictDataCreate(c *gin.Context) {
	var row DictData
	if c.ShouldBindJSON(&row) != nil || strings.TrimSpace(row.Label) == "" || strings.TrimSpace(row.DictType) == "" {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	if h.service.db.Where("type = ?", row.DictType).First(&DictType{}).Error != nil {
		httpx.Fail(c, 400, 400, "字典类型不存在")
		return
	}
	row.ID = 0
	if h.service.db.Create(&row).Error != nil {
		httpx.Fail(c, 500, 500, "创建字典数据失败")
		return
	}
	httpx.OK(c, row.ID)
}

// DictDataUpdate godoc
// @Summary Update dictionary data
// @Tags System Dictionary
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DictData true "Dictionary data"
// @Success 200 {object} httpx.Response
// @Router /system/dict-data/update [put]
func (h *Handler) DictDataUpdate(c *gin.Context) {
	var req DictData
	if c.ShouldBindJSON(&req) != nil || req.ID == 0 {
		httpx.Fail(c, 400, 400, "请求参数错误")
		return
	}
	var row DictData
	if h.service.db.First(&row, req.ID).Error != nil {
		httpx.Fail(c, 404, 404, "字典数据不存在")
		return
	}
	created := row.CreatedAt
	row = req
	row.CreatedAt = created
	h.service.db.Save(&row)
	httpx.OK(c, true)
}

// DictDataDelete godoc
// @Summary Delete dictionary data
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-data/delete [delete]
func (h *Handler) DictDataDelete(c *gin.Context) {
	h.service.db.Delete(&DictData{}, queryID(c))
	httpx.OK(c, true)
}

// DictDataDeleteList godoc
// @Summary Delete dictionary data in batch
// @Tags System Dictionary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Response
// @Router /system/dict-data/delete-list [delete]
func (h *Handler) DictDataDeleteList(c *gin.Context) {
	h.service.db.Where("id IN ?", splitIDs(c.Query("ids"))).Delete(&DictData{})
	httpx.OK(c, true)
}
