package infra

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Config{}, &FileConfig{}, &FileRecord{}, &APIAccessLog{})
}

func Seed(db *gorm.DB, tenant uint64) error {
	config := Config{TenantID: tenant, Category: "基础配置", Type: 2, Name: "平台名称", Key: "platform.name", Value: "Nimbus Framework", Visible: true}
	if err := db.Where("tenant_id = ? AND `key` = ?", tenant, config.Key).FirstOrCreate(&config).Error; err != nil {
		return err
	}
	fileConfig := FileConfig{TenantID: tenant, Name: "本地存储", Storage: 10, Master: true, Config: `{"basePath":"./data/uploads"}`}
	return db.Where("tenant_id = ? AND name = ?", tenant, fileConfig.Name).FirstOrCreate(&fileConfig).Error
}

func Register(group *gin.RouterGroup, db *gorm.DB, auth gin.HandlerFunc) {
	h := NewHandler(db)
	config := group.Group("/infra/config", auth)
	config.GET("/page", h.ConfigPage)
	config.GET("/get", h.ConfigGet)
	config.GET("/get-value-by-key", h.ConfigValue)
	config.POST("/create", h.ConfigCreate)
	config.PUT("/update", h.ConfigUpdate)
	config.DELETE("/delete", h.ConfigDelete)
	config.DELETE("/delete-list", h.ConfigDeleteList)
	config.GET("/export-excel", h.ConfigExport)

	fileConfig := group.Group("/infra/file-config", auth)
	fileConfig.GET("/page", h.FileConfigPage)
	fileConfig.GET("/get", h.FileConfigGet)
	fileConfig.GET("/test", h.FileConfigTest)
	fileConfig.POST("/create", h.FileConfigCreate)
	fileConfig.PUT("/update", h.FileConfigUpdate)
	fileConfig.PUT("/update-master", h.FileConfigMaster)
	fileConfig.DELETE("/delete", h.FileConfigDelete)
	fileConfig.DELETE("/delete-list", h.FileConfigDeleteList)

	logs := group.Group("/infra/api-access-log", auth)
	logs.GET("/page", h.AccessLogPage)
	logs.GET("/export-excel", h.AccessLogExport)
}
