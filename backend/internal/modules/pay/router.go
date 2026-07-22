package pay

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&App{}, &Channel{}, &Order{}, &Refund{}, &Wallet{}, &WalletTransaction{})
}

func Seed(db *gorm.DB, tenant uint64) error {
	app := App{TenantID: tenant, Name: "Nimbus 默认应用", AppKey: "nimbus-default", Status: 0, Remark: "可删除的开发环境初始化应用"}
	return db.Where("tenant_id = ? AND app_key = ?", tenant, app.AppKey).FirstOrCreate(&app).Error
}

func Register(group *gin.RouterGroup, db *gorm.DB, auth gin.HandlerFunc) {
	h := NewHandler(db)
	apps := group.Group("/pay/app", auth)
	apps.GET("/page", h.AppPage)
	apps.GET("/simple-list", h.AppSimple)
	apps.GET("/list", h.AppList)
	apps.GET("/get", h.AppGet)
	apps.POST("/create", h.AppCreate)
	apps.PUT("/update", h.AppUpdate)
	apps.PUT("/update-status", h.AppStatus)
	apps.DELETE("/delete", h.AppDelete)

	channels := group.Group("/pay/channel", auth)
	channels.GET("/page", h.ChannelPage)
	channels.GET("/get-enable-channel-code-list", h.ChannelCodes)
	channels.GET("/get-enable-code-list", h.ChannelCodes)
	channels.GET("/list", h.ChannelList)
	channels.GET("/get", h.ChannelGet)
	channels.POST("/create", h.ChannelCreate)
	channels.PUT("/update", h.ChannelUpdate)
	channels.DELETE("/delete", h.ChannelDelete)
	channels.GET("/export-excel", h.ChannelExport)

	orders := group.Group("/pay/order", auth)
	orders.GET("/page", h.OrderPage)
	orders.GET("/get", h.OrderGet)
	orders.GET("/get-detail", h.OrderDetail)
	orders.POST("/create", h.OrderCreate)
	orders.POST("/submit", h.OrderCreate)
	orders.GET("/export-excel", h.OrderExport)

	refunds := group.Group("/pay/refund", auth)
	refunds.GET("/page", h.RefundPage)
	refunds.GET("/get", h.RefundGet)
	refunds.POST("/create", h.RefundCreate)
	refunds.PUT("/update", h.RefundUpdate)
	refunds.DELETE("/delete", h.RefundDelete)
	refunds.GET("/export-excel", h.RefundExport)

	wallets := group.Group("/pay/wallet", auth)
	wallets.GET("/get", h.WalletGet)
	wallets.GET("/page", h.WalletPage)
	wallets.PUT("/update-balance", h.WalletBalanceUpdate)
	group.GET("/pay/wallet-transaction/page", auth, h.WalletTransactionPage)
}
