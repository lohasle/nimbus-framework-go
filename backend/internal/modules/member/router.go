package member

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Level{}, &Group{}, &Tag{}, &UserTag{}, &PointRecord{}, &ExperienceRecord{})
}

func Seed(db *gorm.DB, tenant uint64) error {
	level := Level{TenantID: tenant, Name: "普通会员", Level: 1, Experience: 0, Discount: 100, Status: 0}
	if err := db.Where("tenant_id = ? AND level = ?", tenant, level.Level).FirstOrCreate(&level).Error; err != nil {
		return err
	}
	group := Group{TenantID: tenant, Name: "默认分组", Status: 0}
	return db.Where("tenant_id = ? AND name = ?", tenant, group.Name).FirstOrCreate(&group).Error
}

func Register(group *gin.RouterGroup, db *gorm.DB, auth gin.HandlerFunc) {
	h := NewHandler(db)
	users := group.Group("/member/user", auth)
	users.GET("/page", h.UserPage)
	users.GET("/get", h.UserGet)
	users.POST("/create", h.UserCreate)
	users.PUT("/update", h.UserUpdate)
	users.PUT("/update-status", h.UserStatus)
	users.PUT("/update-point", h.UserPoint)
	users.PUT("/update-level", h.UserLevel)
	users.DELETE("/delete", h.UserDelete)

	levels := group.Group("/member/level", auth)
	levels.GET("/page", h.LevelPage)
	levels.GET("/list", h.LevelList)
	levels.GET("/get", h.LevelGet)
	levels.GET("/simple-list", h.LevelSimple)
	levels.GET("/list-all-simple", h.LevelSimple)
	levels.POST("/create", h.LevelCreate)
	levels.PUT("/update", h.LevelUpdate)
	levels.DELETE("/delete", h.EntityDelete(&Level{}))

	groups := group.Group("/member/group", auth)
	groups.GET("/page", h.GroupPage)
	groups.GET("/get", h.GroupGet)
	groups.GET("/simple-list", h.GroupSimple)
	groups.GET("/list-all-simple", h.GroupSimple)
	groups.POST("/create", h.GroupCreate)
	groups.PUT("/update", h.GroupUpdate)
	groups.DELETE("/delete", h.EntityDelete(&Group{}))

	tags := group.Group("/member/tag", auth)
	tags.GET("/page", h.TagPage)
	tags.GET("/get", h.TagGet)
	tags.GET("/simple-list", h.TagSimple)
	tags.GET("/list-all-simple", h.TagSimple)
	tags.POST("/create", h.TagCreate)
	tags.PUT("/update", h.TagUpdate)
	tags.DELETE("/delete", h.EntityDelete(&Tag{}))
}
