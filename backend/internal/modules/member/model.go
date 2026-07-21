package member

import "time"

type User struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	TenantID   uint64     `gorm:"uniqueIndex:uk_member_mobile;not null" json:"tenantId"`
	Mobile     string     `gorm:"size:32;uniqueIndex:uk_member_mobile" json:"mobile"`
	Nickname   string     `gorm:"size:64;not null" json:"nickname"`
	Avatar     string     `gorm:"size:512" json:"avatar"`
	Sex        int        `gorm:"not null;default:0" json:"sex"`
	Status     int        `gorm:"not null;default:0" json:"status"`
	LevelID    uint64     `gorm:"index;not null;default:0" json:"levelId"`
	GroupID    uint64     `gorm:"index;not null;default:0" json:"groupId"`
	Point      int64      `gorm:"not null;default:0" json:"point"`
	Experience int64      `gorm:"not null;default:0" json:"experience"`
	Balance    int64      `gorm:"not null;default:0" json:"balance"`
	RegisterIP string     `gorm:"size:64" json:"registerIp"`
	LoginIP    string     `gorm:"size:64" json:"loginIp"`
	LoginDate  *time.Time `json:"loginDate"`
	Remark     string     `gorm:"size:512" json:"remark"`
	CreatedAt  time.Time  `json:"createTime"`
	UpdatedAt  time.Time  `json:"updateTime"`
}

type Level struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	TenantID   uint64    `gorm:"index;not null" json:"tenantId"`
	Name       string    `gorm:"size:64;not null" json:"name"`
	Level      int       `gorm:"not null" json:"level"`
	Experience int64     `gorm:"not null;default:0" json:"experience"`
	Discount   int       `gorm:"not null;default:100" json:"discountPercent"`
	Icon       string    `gorm:"size:512" json:"icon"`
	Background string    `gorm:"size:512" json:"backgroundUrl"`
	Status     int       `gorm:"not null;default:0" json:"status"`
	CreatedAt  time.Time `json:"createTime"`
	UpdatedAt  time.Time `json:"updateTime"`
}

type Group struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Remark    string    `gorm:"size:512" json:"remark"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type Tag struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Status    int       `gorm:"not null;default:0" json:"status"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type UserTag struct {
	UserID uint64 `gorm:"primaryKey" json:"userId"`
	TagID  uint64 `gorm:"primaryKey" json:"tagId"`
}

type PointRecord struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	TenantID    uint64    `gorm:"index;not null" json:"tenantId"`
	UserID      uint64    `gorm:"index;not null" json:"userId"`
	BizType     string    `gorm:"size:64" json:"bizType"`
	BizID       string    `gorm:"size:128" json:"bizId"`
	Point       int64     `gorm:"not null" json:"point"`
	TotalPoint  int64     `gorm:"not null" json:"totalPoint"`
	Description string    `gorm:"size:512" json:"description"`
	CreatedAt   time.Time `json:"createTime"`
}

type ExperienceRecord struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	TenantID        uint64    `gorm:"index;not null" json:"tenantId"`
	UserID          uint64    `gorm:"index;not null" json:"userId"`
	BizType         string    `gorm:"size:64" json:"bizType"`
	BizID           string    `gorm:"size:128" json:"bizId"`
	Experience      int64     `gorm:"not null" json:"experience"`
	TotalExperience int64     `gorm:"not null" json:"totalExperience"`
	Description     string    `gorm:"size:512" json:"description"`
	CreatedAt       time.Time `json:"createTime"`
}

type UserSaveRequest struct {
	ID       uint64   `json:"id"`
	Mobile   string   `json:"mobile"`
	Nickname string   `json:"nickname" binding:"required"`
	Avatar   string   `json:"avatar"`
	Sex      int      `json:"sex"`
	Status   int      `json:"status"`
	LevelID  uint64   `json:"levelId"`
	GroupID  uint64   `json:"groupId"`
	TagIDs   []uint64 `json:"tagIds"`
	Remark   string   `json:"remark"`
}

type LevelSaveRequest struct {
	ID         uint64 `json:"id"`
	Name       string `json:"name" binding:"required"`
	Level      int    `json:"level" binding:"required"`
	Experience int64  `json:"experience"`
	Discount   int    `json:"discountPercent"`
	Icon       string `json:"icon"`
	Background string `json:"backgroundUrl"`
	Status     int    `json:"status"`
}

type GroupSaveRequest struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
	Status int    `json:"status"`
}

type TagSaveRequest struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name" binding:"required"`
	Status int    `json:"status"`
}

func (User) TableName() string             { return "member_user" }
func (Level) TableName() string            { return "member_level" }
func (Group) TableName() string            { return "member_group" }
func (Tag) TableName() string              { return "member_tag" }
func (UserTag) TableName() string          { return "member_user_tag" }
func (PointRecord) TableName() string      { return "member_point_record" }
func (ExperienceRecord) TableName() string { return "member_experience_record" }
