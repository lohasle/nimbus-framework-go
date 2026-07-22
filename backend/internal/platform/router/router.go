package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appmodule "github.com/lohasle/nimbus-framework-go/internal/modules/app"
	"github.com/lohasle/nimbus-framework-go/internal/modules/application"
	"github.com/lohasle/nimbus-framework-go/internal/modules/im"
	"github.com/lohasle/nimbus-framework-go/internal/modules/infra"
	"github.com/lohasle/nimbus-framework-go/internal/modules/member"
	"github.com/lohasle/nimbus-framework-go/internal/modules/pay"
	"github.com/lohasle/nimbus-framework-go/internal/modules/system"
	"github.com/lohasle/nimbus-framework-go/internal/platform/httpx"
	"github.com/lohasle/nimbus-framework-go/internal/platform/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func New(handler *system.Handler, db *gorm.DB) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RequestContext(func(record middleware.RequestLog) {
		if record.TenantID == 0 {
			return
		}
		db.Create(&infra.APIAccessLog{
			TenantID: record.TenantID, TraceID: record.TraceID, UserID: record.UserID,
			Method: record.Method, Path: record.Path, Status: record.Status, Duration: record.Duration,
			IP: record.IP, UserAgent: record.UserAgent,
		})
	}))
	r.GET("/health", health("nimbus-server"))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	admin := r.Group("/admin-api")
	admin.GET("/system/tenant/get-id-by-name", handler.TenantID)
	admin.GET("/system/tenant/get-by-website", handler.TenantByWebsite)
	admin.POST("/system/auth/login", handler.Login)
	admin.GET("/system/auth/get-permission-info", handler.Auth(), handler.PermissionInfo)
	admin.GET("/system/dict-data/simple-list", handler.Auth(), handler.SimpleDictData)
	admin.GET("/system/notify-message/get-unread-count", handler.Auth(), handler.UnreadNotifyMessageCount)
	admin.GET("/system/notify-message/get-unread-list", handler.Auth(), handler.UnreadNotifyMessageList)
	admin.POST("/system/auth/logout", handler.Auth(), handler.Logout)
	systemAdmin := admin.Group("/system", handler.Auth())
	systemAdmin.GET("/user/page", handler.UserPage)
	systemAdmin.GET("/user/list", handler.UserList)
	systemAdmin.GET("/user/simple-list", handler.SimpleUsers)
	systemAdmin.GET("/user/get-simple", handler.SimpleUser)
	systemAdmin.GET("/user/list-by-nickname", handler.UsersByNickname)
	systemAdmin.GET("/user/get", handler.UserGet)
	systemAdmin.POST("/user/create", handler.UserCreate)
	systemAdmin.PUT("/user/update", handler.UserUpdate)
	systemAdmin.DELETE("/user/delete", handler.UserDelete)
	systemAdmin.DELETE("/user/delete-list", handler.UserDeleteList)
	systemAdmin.PUT("/user/update-password", handler.UserPassword)
	systemAdmin.PUT("/user/update-status", handler.UserStatus)
	systemAdmin.GET("/user/export-excel", handler.UserExport)
	systemAdmin.GET("/user/get-import-template", handler.UserImportTemplate)
	systemAdmin.POST("/user/import", handler.UserImport)
	systemAdmin.GET("/role/simple-list", handler.SimpleRoles)
	systemAdmin.GET("/permission/list-user-roles", handler.UserRoleList)
	systemAdmin.POST("/permission/assign-user-role", handler.AssignUserRole)
	systemAdmin.GET("/dept/simple-list", handler.SimpleDepartments)
	systemAdmin.GET("/post/simple-list", handler.SimplePosts)
	systemAdmin.GET("/area/tree", handler.AreaTree)
	systemAdmin.GET("/area/get-by-ip", handler.AreaByIP)
	infra.Register(admin, db, handler.Auth())
	member.Register(admin, db, handler.Auth())
	pay.Register(admin, db, handler.Auth())
	for _, module := range []string{
		system.ModuleName,
		infra.ModuleName,
		member.ModuleName,
		pay.ModuleName,
		application.ModuleName,
		im.ModuleName,
		appmodule.ModuleName,
	} {
		admin.GET("/"+module+"/health", health(module))
	}
	r.NoRoute(func(c *gin.Context) {
		httpx.Fail(c, http.StatusNotFound, 404, "请求地址不存在:"+c.Request.URL.Path)
	})
	return r
}

// health godoc
// @Summary Service health
// @Description Health endpoint used by deployment probes and scaffold modules.
// @Tags Health
// @Produce json
// @Success 200 {object} httpx.Response
// @Router /health [get]
// @Router /system/health [get]
// @Router /infra/health [get]
// @Router /member/health [get]
// @Router /pay/health [get]
// @Router /application/health [get]
// @Router /im/health [get]
// @Router /app/health [get]
func health(module string) gin.HandlerFunc {
	return func(c *gin.Context) { httpx.OK(c, gin.H{"status": "UP", "service": module}) }
}
