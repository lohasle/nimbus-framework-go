package pay

import "time"

type App struct {
	ID                uint64    `gorm:"primaryKey" json:"id"`
	TenantID          uint64    `gorm:"index;not null" json:"tenantId"`
	Name              string    `gorm:"size:128;not null" json:"name"`
	AppKey            string    `gorm:"size:64;uniqueIndex;not null" json:"appKey"`
	Status            int       `gorm:"not null;default:0" json:"status"`
	Remark            string    `gorm:"size:512" json:"remark"`
	OrderNotifyURL    string    `gorm:"size:1024" json:"orderNotifyUrl"`
	RefundNotifyURL   string    `gorm:"size:1024" json:"refundNotifyUrl"`
	TransferNotifyURL string    `gorm:"size:1024" json:"transferNotifyUrl"`
	CreatedAt         time.Time `json:"createTime"`
	UpdatedAt         time.Time `json:"updateTime"`
}

type Channel struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	AppID     uint64    `gorm:"uniqueIndex:uk_pay_channel;not null" json:"appId"`
	Code      string    `gorm:"size:64;uniqueIndex:uk_pay_channel;not null" json:"code"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	FeeRate   int       `gorm:"not null;default:0" json:"feeRate"`
	Config    string    `gorm:"type:text" json:"config"`
	Remark    string    `gorm:"size:512" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type Order struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	TenantID        uint64     `gorm:"index;not null" json:"tenantId"`
	AppID           uint64     `gorm:"index;not null" json:"appId"`
	ChannelCode     string     `gorm:"size:64;index" json:"channelCode"`
	MerchantOrderNo string     `gorm:"size:128;uniqueIndex;not null" json:"merchantOrderId"`
	ChannelOrderNo  string     `gorm:"size:128;index" json:"channelOrderNo"`
	Subject         string     `gorm:"size:256;not null" json:"subject"`
	Body            string     `gorm:"size:512" json:"body"`
	Price           int64      `gorm:"not null" json:"price"`
	Status          int        `gorm:"not null;default:0" json:"status"`
	RefundPrice     int64      `gorm:"not null;default:0" json:"refundPrice"`
	ClientIP        string     `gorm:"size:64" json:"userIp"`
	SuccessTime     *time.Time `json:"successTime"`
	CreatedAt       time.Time  `json:"createTime"`
	UpdatedAt       time.Time  `json:"updateTime"`
}

type Refund struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	TenantID         uint64     `gorm:"index;not null" json:"tenantId"`
	OrderID          uint64     `gorm:"index;not null" json:"orderId"`
	MerchantRefundNo string     `gorm:"size:128;uniqueIndex;not null" json:"merchantRefundId"`
	ChannelRefundNo  string     `gorm:"size:128;index" json:"channelRefundNo"`
	Price            int64      `gorm:"not null" json:"refundPrice"`
	Reason           string     `gorm:"size:512" json:"reason"`
	Status           int        `gorm:"not null;default:0" json:"status"`
	SuccessTime      *time.Time `json:"successTime"`
	CreatedAt        time.Time  `json:"createTime"`
	UpdatedAt        time.Time  `json:"updateTime"`
}

type Wallet struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	TenantID      uint64    `gorm:"index;not null" json:"tenantId"`
	UserID        uint64    `gorm:"uniqueIndex:uk_pay_wallet_user;not null" json:"userId"`
	UserType      int       `gorm:"uniqueIndex:uk_pay_wallet_user;not null;default:1" json:"userType"`
	Balance       int64     `gorm:"not null;default:0" json:"balance"`
	TotalExpense  int64     `gorm:"not null;default:0" json:"totalExpense"`
	TotalRecharge int64     `gorm:"not null;default:0" json:"totalRecharge"`
	FreezePrice   int64     `gorm:"not null;default:0" json:"freezePrice"`
	CreatedAt     time.Time `json:"createTime"`
	UpdatedAt     time.Time `json:"updateTime"`
}

type WalletTransaction struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	WalletID  uint64    `gorm:"index;not null" json:"walletId"`
	Title     string    `gorm:"size:128;not null" json:"title"`
	Price     int64     `gorm:"not null" json:"price"`
	Balance   int64     `gorm:"not null" json:"balance"`
	CreatedAt time.Time `json:"createTime"`
}

type AppSaveRequest struct {
	ID                uint64 `json:"id"`
	Name              string `json:"name" binding:"required"`
	AppKey            string `json:"appKey"`
	Status            int    `json:"status"`
	Remark            string `json:"remark"`
	OrderNotifyURL    string `json:"orderNotifyUrl"`
	RefundNotifyURL   string `json:"refundNotifyUrl"`
	TransferNotifyURL string `json:"transferNotifyUrl"`
}

type ChannelSaveRequest struct {
	ID      uint64 `json:"id"`
	AppID   uint64 `json:"appId" binding:"required"`
	Code    string `json:"code" binding:"required"`
	Status  int    `json:"status"`
	FeeRate int    `json:"feeRate"`
	Config  string `json:"config"`
	Remark  string `json:"remark"`
}

type OrderCreateRequest struct {
	AppID           uint64 `json:"appId" binding:"required"`
	ChannelCode     string `json:"channelCode"`
	MerchantOrderNo string `json:"merchantOrderId" binding:"required"`
	Subject         string `json:"subject" binding:"required"`
	Body            string `json:"body"`
	Price           int64  `json:"price" binding:"required"`
}

type RefundCreateRequest struct {
	OrderID          uint64 `json:"orderId" binding:"required"`
	MerchantRefundNo string `json:"merchantRefundId" binding:"required"`
	Price            int64  `json:"refundPrice" binding:"required"`
	Reason           string `json:"reason"`
}

func (App) TableName() string               { return "pay_app" }
func (Channel) TableName() string           { return "pay_channel" }
func (Order) TableName() string             { return "pay_order" }
func (Refund) TableName() string            { return "pay_refund" }
func (Wallet) TableName() string            { return "pay_wallet" }
func (WalletTransaction) TableName() string { return "pay_wallet_transaction" }
