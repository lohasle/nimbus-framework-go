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

type APIErrorLog struct {
	ID                        uint64     `gorm:"primaryKey" json:"id"`
	TenantID                  uint64     `gorm:"index;not null" json:"tenantId"`
	TraceID                   string     `gorm:"size:64;index" json:"traceId"`
	UserID                    uint64     `gorm:"index;not null;default:0" json:"userId"`
	UserType                  int        `gorm:"not null;default:2" json:"userType"`
	ApplicationName           string     `gorm:"size:128" json:"applicationName"`
	RequestMethod             string     `gorm:"size:16" json:"requestMethod"`
	RequestParams             string     `gorm:"type:text" json:"requestParams"`
	RequestURL                string     `gorm:"size:1024" json:"requestUrl"`
	UserIP                    string     `gorm:"size:64" json:"userIp"`
	UserAgent                 string     `gorm:"size:512" json:"userAgent"`
	ExceptionTime             time.Time  `json:"exceptionTime"`
	ExceptionName             string     `gorm:"size:512" json:"exceptionName"`
	ExceptionMessage          string     `gorm:"type:text" json:"exceptionMessage"`
	ExceptionRootCauseMessage string     `gorm:"type:text" json:"exceptionRootCauseMessage"`
	ExceptionStackTrace       string     `gorm:"type:longtext" json:"exceptionStackTrace"`
	ExceptionClassName        string     `gorm:"size:512" json:"exceptionClassName"`
	ExceptionFileName         string     `gorm:"size:512" json:"exceptionFileName"`
	ExceptionMethodName       string     `gorm:"size:512" json:"exceptionMethodName"`
	ExceptionLineNumber       int        `json:"exceptionLineNumber"`
	ProcessUserID             uint64     `gorm:"index" json:"processUserId"`
	ProcessStatus             int        `gorm:"not null;default:0" json:"processStatus"`
	ProcessTime               *time.Time `json:"processTime"`
	ResultCode                int        `json:"resultCode"`
	CreatedAt                 time.Time  `json:"createTime"`
}

type DataSourceConfig struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	URL       string    `gorm:"size:1024;not null" json:"url"`
	Username  string    `gorm:"size:256" json:"username"`
	Password  string    `gorm:"size:512" json:"password"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type Job struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	TenantID       uint64    `gorm:"index;not null" json:"tenantId"`
	Name           string    `gorm:"size:128;not null" json:"name"`
	Status         int       `gorm:"not null;default:0" json:"status"`
	HandlerName    string    `gorm:"size:128;not null" json:"handlerName"`
	HandlerParam   string    `gorm:"type:text" json:"handlerParam"`
	CronExpression string    `gorm:"size:128;not null" json:"cronExpression"`
	RetryCount     int       `gorm:"not null;default:0" json:"retryCount"`
	RetryInterval  int       `gorm:"not null;default:0" json:"retryInterval"`
	MonitorTimeout int       `gorm:"not null;default:0" json:"monitorTimeout"`
	CreatedAt      time.Time `json:"createTime"`
	UpdatedAt      time.Time `json:"updateTime"`
}

type JobLog struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	TenantID       uint64    `gorm:"index;not null" json:"tenantId"`
	JobID          uint64    `gorm:"index;not null" json:"jobId"`
	HandlerName    string    `gorm:"size:128" json:"handlerName"`
	HandlerParam   string    `gorm:"type:text" json:"handlerParam"`
	CronExpression string    `gorm:"size:128" json:"cronExpression"`
	ExecuteIndex   int       `json:"executeIndex"`
	BeginTime      time.Time `json:"beginTime"`
	EndTime        time.Time `json:"endTime"`
	Duration       int64     `json:"duration"`
	Status         int       `gorm:"not null;default:0" json:"status"`
	Result         string    `gorm:"type:text" json:"result"`
	CreatedAt      time.Time `json:"createTime"`
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

func (Config) TableName() string           { return "infra_config" }
func (FileConfig) TableName() string       { return "infra_file_config" }
func (FileRecord) TableName() string       { return "infra_file" }
func (APIAccessLog) TableName() string     { return "infra_api_access_log" }
func (APIErrorLog) TableName() string      { return "infra_api_error_log" }
func (DataSourceConfig) TableName() string { return "infra_data_source_config" }
func (Job) TableName() string              { return "infra_job" }
func (JobLog) TableName() string           { return "infra_job_log" }
