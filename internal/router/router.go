package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lohasle/nimbus-framework-go/internal/httpx"
	"github.com/lohasle/nimbus-framework-go/internal/middleware"
	"github.com/lohasle/nimbus-framework-go/internal/system"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(handler *system.Handler) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RequestContext())
	r.GET("/health", health("nimbus-server"))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	admin := r.Group("/admin-api")
	admin.GET("/system/tenant/get-id-by-name", handler.TenantID)
	admin.POST("/system/auth/login", handler.Login)
	admin.GET("/system/auth/get-permission-info", handler.Auth(), handler.PermissionInfo)
	admin.POST("/system/auth/logout", handler.Auth(), handler.Logout)
	for _, module := range []string{"system", "infra", "member", "pay", "application", "im", "app"} {
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
