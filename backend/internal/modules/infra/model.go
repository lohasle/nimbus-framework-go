package infra

import "time"

type Config struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"uniqueIndex:uk_infra_config;not null" json:"tenantId"`
	Category  string    `gorm:"size:64;not null" json:"category"`
	Type      int       `gorm:"not null;default:2" json:"type"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Key       string    `gorm:"size:128;uniqueIndex:uk_infra_config;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Visible   bool      `gorm:"not null;default:true" json:"visible"`
	Remark    string    `gorm:"size:512" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type FileConfig struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Storage   int       `gorm:"not null;default:10" json:"storage"`
	Master    bool      `gorm:"not null;default:false" json:"master"`
	Config    string    `gorm:"type:text" json:"config"`
	Remark    string    `gorm:"size:512" json:"remark"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type FileRecord struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	TenantID    uint64    `gorm:"index;not null" json:"tenantId"`
	ConfigID    uint64    `gorm:"index;not null" json:"configId"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Path        string    `gorm:"size:512;not null" json:"path"`
	URL         string    `gorm:"size:1024;not null" json:"url"`
	ContentType string    `gorm:"size:128" json:"type"`
	Size        int64     `gorm:"not null;default:0" json:"size"`
	CreatedAt   time.Time `json:"createTime"`
	UpdatedAt   time.Time `json:"updateTime"`
}

type APIAccessLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	TraceID   string    `gorm:"size:64;index" json:"traceId"`
	UserID    uint64    `gorm:"index;not null;default:0" json:"userId"`
	Method    string    `gorm:"size:16;not null" json:"requestMethod"`
	Path      string    `gorm:"size:512;not null" json:"requestUrl"`
	Status    int       `gorm:"not null" json:"responseCode"`
	Duration  int64     `gorm:"not null;default:0" json:"duration"`
	IP        string    `gorm:"size:64" json:"userIp"`
	UserAgent string    `gorm:"size:512" json:"userAgent"`
	CreatedAt time.Time `json:"createTime"`
}

type ConfigSaveRequest struct {
	ID       uint64 `json:"id"`
	Category string `json:"category" binding:"required"`
	Type     int    `json:"type"`
	Name     string `json:"name" binding:"required"`
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value"`
	Visible  bool   `json:"visible"`
	Remark   string `json:"remark"`
}

type FileConfigSaveRequest struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name" binding:"required"`
	Storage int    `json:"storage"`
	Master  bool   `json:"master"`
	Config  string `json:"config"`
	Remark  string `json:"remark"`
}

func (Config) TableName() string       { return "infra_config" }
func (FileConfig) TableName() string   { return "infra_file_config" }
func (FileRecord) TableName() string   { return "infra_file" }
func (APIAccessLog) TableName() string { return "infra_api_access_log" }
