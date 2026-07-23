package infra

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Config{}, &FileConfig{}, &FileRecord{}, &APIAccessLog{}, &APIErrorLog{}, &DataSourceConfig{}, &Job{}, &JobLog{})
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
	if db != nil {
		newJobScheduler(db).start()
	}
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

	errorLogs := group.Group("/infra/api-error-log", auth)
	errorLogs.GET("/page", h.APIErrorLogPage)
	errorLogs.PUT("/update-status", h.APIErrorLogStatus)
	errorLogs.GET("/export-excel", h.APIErrorLogExport)

	dataSources := group.Group("/infra/data-source-config", auth)
	dataSources.GET("/list", h.DataSourceList)
	dataSources.GET("/get", h.DataSourceGet)
	dataSources.POST("/create", h.DataSourceCreate)
	dataSources.PUT("/update", h.DataSourceUpdate)
	dataSources.DELETE("/delete", h.DataSourceDelete)
	dataSources.DELETE("/delete-list", h.DataSourceDeleteList)

	group.GET("/infra/file/content/:id", h.FileContent)
	files := group.Group("/infra/file", auth)
	files.GET("/page", h.FilePage)
	files.POST("/upload", h.FileUpload)
	files.POST("/create", h.FileCreate)
	files.GET("/presigned-url", h.FilePresignedURL)
	files.DELETE("/delete", h.FileDelete)
	files.DELETE("/delete-list", h.FileDeleteList)

	jobs := group.Group("/infra/job", auth)
	jobs.GET("/page", h.JobPage)
	jobs.GET("/get", h.JobGet)
	jobs.POST("/create", h.JobCreate)
	jobs.PUT("/update", h.JobUpdate)
	jobs.DELETE("/delete", h.JobDelete)
	jobs.DELETE("/delete-list", h.JobDeleteList)
	jobs.PUT("/update-status", h.JobStatus)
	jobs.PUT("/trigger", h.JobTrigger)
	jobs.GET("/get_next_times", h.JobNextTimes)
	jobs.POST("/sync", h.JobSync)
	jobs.GET("/export-excel", h.JobExport)

	jobLogs := group.Group("/infra/job-log", auth)
	jobLogs.GET("/page", h.JobLogPage)
	jobLogs.GET("/get", h.JobLogGet)
	jobLogs.GET("/export-excel", h.JobLogExport)

	group.GET("/infra/redis/get-monitor-info", auth, h.RedisMonitor)
}
