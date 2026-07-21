package system

import "time"

const ModuleName = "system"

type Tenant struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	ContactName   string    `gorm:"size:64" json:"contactName"`
	ContactMobile string    `gorm:"size:32" json:"contactMobile"`
	Domain        string    `gorm:"size:128" json:"domain"`
	Status        int       `gorm:"not null;default:0" json:"status"`
	ExpireTime    time.Time `json:"expireTime"`
	AccountCount  int       `gorm:"not null;default:100" json:"accountCount"`
	CreatedAt     time.Time `json:"createTime"`
	UpdatedAt     time.Time `json:"updateTime"`
}

type AdminUser struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"uniqueIndex:uk_tenant_username;not null" json:"tenantId"`
	Username     string    `gorm:"size:64;uniqueIndex:uk_tenant_username;not null" json:"username"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	Nickname     string    `gorm:"size:64;not null" json:"nickname"`
	Email        string    `gorm:"size:128" json:"email"`
	Mobile       string    `gorm:"size:32" json:"mobile"`
	Sex          int       `gorm:"not null;default:0" json:"sex"`
	Avatar       string    `gorm:"size:512" json:"avatar"`
	DeptID       uint64    `gorm:"not null;default:1" json:"deptId"`
	Status       int       `gorm:"not null;default:0" json:"status"`
	Remark       string    `gorm:"size:512" json:"remark"`
	LoginIP      string    `gorm:"size:64" json:"loginIp"`
	LoginDate    time.Time `json:"loginDate"`
	CreatedAt    time.Time `json:"createTime"`
	UpdatedAt    time.Time `json:"updateTime"`
}

type Department struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"index;not null" json:"tenantId"`
	Name         string    `gorm:"size:64;not null" json:"name"`
	ParentID     uint64    `gorm:"not null;default:0" json:"parentId"`
	Sort         int       `gorm:"not null;default:0" json:"sort"`
	LeaderUserID uint64    `gorm:"not null;default:0" json:"leaderUserId"`
	Phone        string    `gorm:"size:32" json:"phone"`
	Email        string    `gorm:"size:128" json:"email"`
	Status       int       `gorm:"not null;default:0" json:"status"`
	CreatedAt    time.Time `json:"createTime"`
	UpdatedAt    time.Time `json:"updateTime"`
}

type Post struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Code      string    `gorm:"size:64;not null" json:"code"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	Remark    string    `gorm:"size:512" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type DictData struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	Label     string    `gorm:"size:128;not null" json:"label"`
	Value     string    `gorm:"size:128;not null" json:"value"`
	DictType  string    `gorm:"size:128;index;not null" json:"dictType"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	ColorType string    `gorm:"size:32" json:"colorType"`
	CSSClass  string    `gorm:"size:128" json:"cssClass"`
	Remark    string    `gorm:"size:512" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type NotifyMessage struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	TenantID         uint64     `gorm:"index;not null" json:"tenantId"`
	UserID           uint64     `gorm:"index;not null" json:"userId"`
	UserType         int        `gorm:"not null;default:2" json:"userType"`
	TemplateID       uint64     `gorm:"not null;default:0" json:"templateId"`
	TemplateCode     string     `gorm:"size:128" json:"templateCode"`
	TemplateNickname string     `gorm:"size:128" json:"templateNickname"`
	TemplateContent  string     `gorm:"size:2048" json:"templateContent"`
	TemplateType     int        `gorm:"not null;default:0" json:"templateType"`
	TemplateParams   string     `gorm:"type:text" json:"templateParams"`
	ReadStatus       bool       `gorm:"index;not null;default:false" json:"readStatus"`
	ReadTime         *time.Time `json:"readTime"`
	CreatedAt        time.Time  `json:"createTime"`
	UpdatedAt        time.Time  `json:"updateTime"`
}

type UserView struct {
	AdminUser
	DeptName string   `json:"deptName"`
	PostIDs  []uint64 `gorm:"-" json:"postIds"`
	RoleIDs  []uint64 `gorm:"-" json:"roleIds"`
}

type UserSaveRequest struct {
	ID       uint64   `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Nickname string   `json:"nickname" binding:"required"`
	DeptID   uint64   `json:"deptId"`
	Mobile   string   `json:"mobile"`
	Email    string   `json:"email"`
	Sex      int      `json:"sex"`
	PostIDs  []uint64 `json:"postIds"`
	RoleIDs  []uint64 `json:"roleIds"`
	Remark   string   `json:"remark"`
	Status   int      `json:"status"`
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
