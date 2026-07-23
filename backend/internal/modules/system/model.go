package system

import "time"

const ModuleName = "system"

type Tenant struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	ContactName   string    `gorm:"size:64" json:"contactName"`
	ContactMobile string    `gorm:"size:32" json:"contactMobile"`
	Domain        string    `gorm:"size:128" json:"domain"`
	PackageID     uint64    `gorm:"not null;default:0" json:"packageId"`
	Status        int       `gorm:"not null;default:0" json:"status"`
	ExpireTime    time.Time `json:"expireTime"`
	AccountCount  int       `gorm:"not null;default:100" json:"accountCount"`
	CreatedAt     time.Time `json:"createTime"`
	UpdatedAt     time.Time `json:"updateTime"`
}

type TenantPackage struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	Remark    string    `gorm:"size:512" json:"remark"`
	MenuIDs   string    `gorm:"type:text" json:"-"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
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

type Role struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	TenantID         uint64    `gorm:"index;not null" json:"tenantId"`
	Name             string    `gorm:"size:64;not null" json:"name"`
	Code             string    `gorm:"size:64;not null" json:"code"`
	Sort             int       `gorm:"not null;default:0" json:"sort"`
	Status           int       `gorm:"not null;default:0" json:"status"`
	Type             int       `gorm:"not null;default:2" json:"type"`
	DataScope        int       `gorm:"not null;default:1" json:"dataScope"`
	DataScopeDeptIDs string    `gorm:"type:text" json:"-"`
	Remark           string    `gorm:"size:512" json:"remark"`
	CreatedAt        time.Time `json:"createTime"`
	UpdatedAt        time.Time `json:"updateTime"`
}

type UserRole struct {
	UserID uint64 `gorm:"primaryKey" json:"userId"`
	RoleID uint64 `gorm:"primaryKey" json:"roleId"`
}

type UserPost struct {
	UserID uint64 `gorm:"primaryKey" json:"userId"`
	PostID uint64 `gorm:"primaryKey" json:"postId"`
}

type SystemMenu struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	TenantID      uint64    `gorm:"index;not null" json:"tenantId"`
	Name          string    `gorm:"size:64;not null" json:"name"`
	Permission    string    `gorm:"size:128;index" json:"permission"`
	Type          int       `gorm:"not null;default:2" json:"type"`
	Sort          int       `gorm:"not null;default:0" json:"sort"`
	ParentID      uint64    `gorm:"index;not null;default:0" json:"parentId"`
	Path          string    `gorm:"size:256" json:"path"`
	Icon          string    `gorm:"size:128" json:"icon"`
	Component     string    `gorm:"size:256" json:"component"`
	ComponentName string    `gorm:"size:128" json:"componentName"`
	Status        int       `gorm:"not null;default:0" json:"status"`
	Visible       bool      `gorm:"not null;default:true" json:"visible"`
	KeepAlive     bool      `gorm:"not null;default:true" json:"keepAlive"`
	AlwaysShow    bool      `gorm:"not null;default:false" json:"alwaysShow"`
	CreatedAt     time.Time `json:"createTime"`
	UpdatedAt     time.Time `json:"updateTime"`
}

type RoleMenu struct {
	RoleID uint64 `gorm:"primaryKey" json:"roleId"`
	MenuID uint64 `gorm:"primaryKey" json:"menuId"`
}

type DictType struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Type      string    `gorm:"size:128;uniqueIndex;not null" json:"type"`
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

type Notice struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	TenantID  uint64     `gorm:"index;not null" json:"tenantId"`
	Title     string     `gorm:"size:256;not null" json:"title"`
	Type      int        `gorm:"not null" json:"type"`
	Content   string     `gorm:"type:text;not null" json:"content"`
	Status    int        `gorm:"not null;default:0" json:"status"`
	Remark    string     `gorm:"size:512" json:"remark"`
	Creator   string     `gorm:"size:64" json:"creator"`
	PushedAt  *time.Time `json:"pushedAt"`
	CreatedAt time.Time  `json:"createTime"`
	UpdatedAt time.Time  `json:"updateTime"`
}

type NotifyTemplate struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Nickname  string    `gorm:"size:128" json:"nickname"`
	Code      string    `gorm:"size:128;uniqueIndex:uk_notify_template;not null" json:"code"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Type      int       `gorm:"not null;default:0" json:"type"`
	Params    string    `gorm:"type:text" json:"params"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	Remark    string    `gorm:"size:512" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type MailAccount struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	TenantID       uint64    `gorm:"index;not null" json:"tenantId"`
	Mail           string    `gorm:"size:256;not null" json:"mail"`
	Username       string    `gorm:"size:256" json:"username"`
	Password       string    `gorm:"size:512" json:"password"`
	Host           string    `gorm:"size:256;not null" json:"host"`
	Port           int       `gorm:"not null" json:"port"`
	SSLEnable      bool      `gorm:"not null;default:false" json:"sslEnable"`
	StartTLSEnable bool      `gorm:"not null;default:false" json:"starttlsEnable"`
	CreatedAt      time.Time `json:"createTime"`
	UpdatedAt      time.Time `json:"updateTime"`
}

type MailTemplate struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Code      string    `gorm:"size:128;uniqueIndex:uk_mail_template;not null" json:"code"`
	AccountID uint64    `gorm:"index;not null" json:"accountId"`
	Nickname  string    `gorm:"size:128" json:"nickname"`
	Title     string    `gorm:"size:512;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type MailLog struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	TenantID         uint64     `gorm:"index;not null" json:"tenantId"`
	UserID           uint64     `gorm:"index;not null;default:0" json:"userId"`
	UserType         int        `gorm:"not null;default:2" json:"userType"`
	ToMails          string     `gorm:"type:text" json:"-"`
	CCMails          string     `gorm:"type:text" json:"-"`
	BCCMails         string     `gorm:"type:text" json:"-"`
	AccountID        uint64     `gorm:"index;not null" json:"accountId"`
	FromMail         string     `gorm:"size:256" json:"fromMail"`
	TemplateID       uint64     `gorm:"index;not null" json:"templateId"`
	TemplateCode     string     `gorm:"size:128" json:"templateCode"`
	TemplateNickname string     `gorm:"size:128" json:"templateNickname"`
	TemplateTitle    string     `gorm:"size:512" json:"templateTitle"`
	TemplateContent  string     `gorm:"type:text" json:"templateContent"`
	TemplateParams   string     `gorm:"type:text" json:"templateParams"`
	SendStatus       int        `gorm:"not null;default:0" json:"sendStatus"`
	SendTime         *time.Time `json:"sendTime"`
	SendMessageID    string     `gorm:"size:256" json:"sendMessageId"`
	SendException    string     `gorm:"type:text" json:"sendException"`
	CreatedAt        time.Time  `json:"createTime"`
}

type SMSChannel struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	TenantID    uint64    `gorm:"index;not null" json:"tenantId"`
	Code        string    `gorm:"size:64;not null" json:"code"`
	Status      int       `gorm:"not null;default:0" json:"status"`
	Signature   string    `gorm:"size:128" json:"signature"`
	Remark      string    `gorm:"size:512" json:"remark"`
	APIKey      string    `gorm:"size:512" json:"apiKey"`
	APISecret   string    `gorm:"size:512" json:"apiSecret"`
	CallbackURL string    `gorm:"size:1024" json:"callbackUrl"`
	CreatedAt   time.Time `json:"createTime"`
	UpdatedAt   time.Time `json:"updateTime"`
}

type SMSTemplate struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	TenantID      uint64    `gorm:"index;not null" json:"tenantId"`
	Type          int       `gorm:"not null;default:0" json:"type"`
	Status        int       `gorm:"not null;default:0" json:"status"`
	Code          string    `gorm:"size:128;uniqueIndex:uk_sms_template;not null" json:"code"`
	Name          string    `gorm:"size:128;not null" json:"name"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	Remark        string    `gorm:"size:512" json:"remark"`
	APITemplateID string    `gorm:"size:256" json:"apiTemplateId"`
	ChannelID     uint64    `gorm:"index;not null" json:"channelId"`
	Params        string    `gorm:"type:text" json:"params"`
	CreatedAt     time.Time `json:"createTime"`
	UpdatedAt     time.Time `json:"updateTime"`
}

type SMSLog struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	TenantID        uint64     `gorm:"index;not null" json:"tenantId"`
	ChannelID       uint64     `gorm:"index" json:"channelId"`
	ChannelCode     string     `gorm:"size:64" json:"channelCode"`
	TemplateID      uint64     `gorm:"index" json:"templateId"`
	TemplateCode    string     `gorm:"size:128" json:"templateCode"`
	TemplateType    int        `json:"templateType"`
	TemplateContent string     `gorm:"type:text" json:"templateContent"`
	TemplateParams  string     `gorm:"type:text" json:"templateParams"`
	APITemplateID   string     `gorm:"size:256" json:"apiTemplateId"`
	Mobile          string     `gorm:"size:32;index" json:"mobile"`
	UserID          uint64     `gorm:"index" json:"userId"`
	UserType        int        `json:"userType"`
	SendStatus      int        `json:"sendStatus"`
	SendTime        *time.Time `json:"sendTime"`
	APISendCode     string     `gorm:"size:128" json:"apiSendCode"`
	APISendMsg      string     `gorm:"size:1024" json:"apiSendMsg"`
	APIRequestID    string     `gorm:"size:256" json:"apiRequestId"`
	APISerialNo     string     `gorm:"size:256" json:"apiSerialNo"`
	ReceiveStatus   int        `json:"receiveStatus"`
	ReceiveTime     *time.Time `json:"receiveTime"`
	APIReceiveCode  string     `gorm:"size:128" json:"apiReceiveCode"`
	APIReceiveMsg   string     `gorm:"size:1024" json:"apiReceiveMsg"`
	CreatedAt       time.Time  `json:"createTime"`
}

type AuthSMSCode struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Mobile    string    `gorm:"index;size:32;not null" json:"mobile"`
	Scene     int       `gorm:"index;not null" json:"scene"`
	CodeHash  string    `gorm:"size:128;not null" json:"-"`
	Used      bool      `gorm:"index;not null;default:false" json:"used"`
	ExpireAt  time.Time `gorm:"index;not null" json:"expireTime"`
	CreatedAt time.Time `json:"createTime"`
}

type SocialClient struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"index;not null" json:"tenantId"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	SocialType   int       `gorm:"index;not null" json:"socialType"`
	UserType     int       `gorm:"index;not null" json:"userType"`
	ClientID     string    `gorm:"size:256" json:"clientId"`
	ClientSecret string    `gorm:"size:512" json:"clientSecret"`
	AgentID      string    `gorm:"size:256" json:"agentId"`
	PublicKey    string    `gorm:"type:text" json:"publicKey"`
	Status       int       `gorm:"not null;default:0" json:"status"`
	CreatedAt    time.Time `json:"createTime"`
	UpdatedAt    time.Time `json:"updateTime"`
}

type SocialUser struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"index;not null" json:"tenantId"`
	Type         int       `gorm:"index;not null" json:"type"`
	OpenID       string    `gorm:"size:256;index" json:"openid"`
	Token        string    `gorm:"size:1024" json:"token"`
	RawTokenInfo string    `gorm:"type:text" json:"rawTokenInfo"`
	Nickname     string    `gorm:"size:128" json:"nickname"`
	Avatar       string    `gorm:"size:1024" json:"avatar"`
	RawUserInfo  string    `gorm:"type:text" json:"rawUserInfo"`
	Code         string    `gorm:"size:512" json:"code"`
	State        string    `gorm:"size:512" json:"state"`
	CreatedAt    time.Time `json:"createTime"`
	UpdatedAt    time.Time `json:"updateTime"`
}

type SocialUserBind struct {
	UserID       uint64 `gorm:"primaryKey" json:"userId"`
	UserType     int    `gorm:"primaryKey" json:"userType"`
	SocialUserID uint64 `gorm:"primaryKey" json:"socialUserId"`
}

type LoginLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	LogType   int       `gorm:"not null;default:100" json:"logType"`
	TraceID   string    `gorm:"size:64;index" json:"traceId"`
	UserID    uint64    `gorm:"index;not null;default:0" json:"userId"`
	UserType  int       `gorm:"not null;default:2" json:"userType"`
	Username  string    `gorm:"size:64;index" json:"username"`
	Result    int       `gorm:"not null;default:0" json:"result"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	UserIP    string    `gorm:"size:64" json:"userIp"`
	UserAgent string    `gorm:"size:512" json:"userAgent"`
	CreatedAt time.Time `json:"createTime"`
}

type OperateLog struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	TenantID      uint64    `gorm:"index;not null" json:"tenantId"`
	TraceID       string    `gorm:"size:64;index" json:"traceId"`
	UserType      int       `gorm:"not null;default:2" json:"userType"`
	UserID        uint64    `gorm:"index;not null;default:0" json:"userId"`
	UserName      string    `gorm:"size:64" json:"userName"`
	Type          string    `gorm:"size:64" json:"type"`
	SubType       string    `gorm:"size:64" json:"subType"`
	BizID         uint64    `gorm:"not null;default:0" json:"bizId"`
	Action        string    `gorm:"size:512" json:"action"`
	Extra         string    `gorm:"type:text" json:"extra"`
	RequestMethod string    `gorm:"size:16" json:"requestMethod"`
	RequestURL    string    `gorm:"size:512" json:"requestUrl"`
	UserIP        string    `gorm:"size:64" json:"userIp"`
	UserAgent     string    `gorm:"size:512" json:"userAgent"`
	CreatedAt     time.Time `json:"createTime"`
}

type OAuth2Client struct {
	ID                          uint64    `gorm:"primaryKey" json:"id"`
	TenantID                    uint64    `gorm:"index;not null" json:"tenantId"`
	ClientID                    string    `gorm:"size:128;uniqueIndex;not null" json:"clientId"`
	Secret                      string    `gorm:"size:256;not null" json:"secret"`
	Name                        string    `gorm:"size:128;not null" json:"name"`
	Logo                        string    `gorm:"size:512" json:"logo"`
	Description                 string    `gorm:"size:1024" json:"description"`
	Status                      int       `gorm:"not null;default:0" json:"status"`
	AccessTokenValiditySeconds  int       `gorm:"not null;default:7200" json:"accessTokenValiditySeconds"`
	RefreshTokenValiditySeconds int       `gorm:"not null;default:2592000" json:"refreshTokenValiditySeconds"`
	RedirectURIs                string    `gorm:"type:text" json:"-"`
	AutoApprove                 bool      `gorm:"not null;default:false" json:"autoApprove"`
	AuthorizedGrantTypes        string    `gorm:"type:text" json:"-"`
	Scopes                      string    `gorm:"type:text" json:"-"`
	Authorities                 string    `gorm:"type:text" json:"-"`
	ResourceIDs                 string    `gorm:"type:text" json:"-"`
	AdditionalInformation       string    `gorm:"type:text" json:"additionalInformation"`
	CreatedAt                   time.Time `json:"createTime"`
	UpdatedAt                   time.Time `json:"updateTime"`
}

type OAuth2AccessToken struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"index;not null" json:"tenantId"`
	AccessToken  string    `gorm:"size:512;uniqueIndex;not null" json:"accessToken"`
	RefreshToken string    `gorm:"size:512;index" json:"refreshToken"`
	UserID       uint64    `gorm:"index;not null;default:0" json:"userId"`
	UserType     int       `gorm:"not null;default:0" json:"userType"`
	ClientID     string    `gorm:"size:128;index" json:"clientId"`
	Scopes       string    `gorm:"type:text" json:"-"`
	CreatedAt    time.Time `json:"createTime"`
	ExpiresTime  time.Time `gorm:"index" json:"expiresTime"`
}

type OAuth2AuthorizationCode struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	TenantID    uint64    `gorm:"index;not null" json:"tenantId"`
	Code        string    `gorm:"size:128;uniqueIndex;not null" json:"code"`
	ClientID    string    `gorm:"size:128;index;not null" json:"clientId"`
	UserID      uint64    `gorm:"index;not null" json:"userId"`
	UserType    int       `gorm:"not null;default:2" json:"userType"`
	Scopes      string    `gorm:"type:text" json:"-"`
	RedirectURI string    `gorm:"size:1024" json:"redirectUri"`
	State       string    `gorm:"size:256" json:"state"`
	ExpiresTime time.Time `gorm:"index" json:"expiresTime"`
	CreatedAt   time.Time `json:"createTime"`
}

type OAuth2Approve struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	UserID    uint64    `gorm:"uniqueIndex:uk_oauth_approve;not null" json:"userId"`
	UserType  int       `gorm:"uniqueIndex:uk_oauth_approve;not null;default:2" json:"userType"`
	ClientID  string    `gorm:"uniqueIndex:uk_oauth_approve;size:128;not null" json:"clientId"`
	Scope     string    `gorm:"uniqueIndex:uk_oauth_approve;size:128;not null" json:"scope"`
	Approved  bool      `gorm:"not null" json:"approved"`
	ExpiresAt time.Time `json:"expiresTime"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
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
