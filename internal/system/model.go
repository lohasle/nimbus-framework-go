package system

import "time"

type Tenant struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AdminUser struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"uniqueIndex:uk_tenant_username;not null" json:"tenantId"`
	Username     string    `gorm:"size:64;uniqueIndex:uk_tenant_username;not null" json:"username"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	Nickname     string    `gorm:"size:64;not null" json:"nickname"`
	Email        string    `gorm:"size:128" json:"email"`
	DeptID       uint64    `gorm:"not null;default:1" json:"deptId"`
	Status       int       `gorm:"not null;default:0" json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

type TokenResponse struct {
	UserID       uint64 `json:"userId" example:"1"`
	UserType     int    `json:"userType" example:"2"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresTime  int64  `json:"expiresTime"`
}

type Menu struct {
	ID            uint64  `json:"id"`
	ParentID      uint64  `json:"parentId"`
	Name          string  `json:"name"`
	Path          string  `json:"path"`
	Component     *string `json:"component"`
	ComponentName *string `json:"componentName"`
	Icon          string  `json:"icon"`
	Visible       bool    `json:"visible"`
	KeepAlive     bool    `json:"keepAlive"`
	AlwaysShow    bool    `json:"alwaysShow"`
	Children      []Menu  `json:"children"`
}
