package database

import (
	"fmt"
	"log/slog"

	"github.com/lohasle/nimbus-framework-go/internal/modules/system"
	"github.com/lohasle/nimbus-framework-go/internal/platform/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	if err = db.AutoMigrate(
		&system.Tenant{},
		&system.AdminUser{},
		&system.Department{},
		&system.Post{},
		&system.Role{},
		&system.UserRole{},
		&system.UserPost{},
		&system.DictData{},
		&system.NotifyMessage{},
	); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}
	if err = bootstrap(db, cfg); err != nil {
		return nil, err
	}
	slog.Info("mysql initialized", "database", "nimbus_platform_go")
	return db, nil
}

func bootstrap(db *gorm.DB, cfg config.Config) error {
	tenant := system.Tenant{
		Name: cfg.BootstrapTenant, Status: 0, ContactName: "Nimbus 管理员",
		ContactMobile: "13800000000", Domain: "localhost", AccountCount: 100,
	}
	if err := db.Where("name = ?", tenant.Name).FirstOrCreate(&tenant).Error; err != nil {
		return fmt.Errorf("bootstrap tenant: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.BootstrapPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap password: %w", err)
	}
	user := system.AdminUser{TenantID: tenant.ID, Username: cfg.BootstrapUser}
	if err = db.Where("tenant_id = ? AND username = ?", tenant.ID, user.Username).Attrs(system.AdminUser{
		PasswordHash: string(hash), Nickname: "平台管理员", Email: "admin@nimbus.local", DeptID: 1, Status: 0,
	}).FirstOrCreate(&user).Error; err != nil {
		return fmt.Errorf("bootstrap admin: %w", err)
	}
	department := system.Department{ID: 1, TenantID: tenant.ID}
	if err = db.Where("id = ? AND tenant_id = ?", department.ID, tenant.ID).Attrs(system.Department{
		Name: "Nimbus 总部", ParentID: 0, Sort: 0, Status: 0, LeaderUserID: user.ID,
	}).FirstOrCreate(&department).Error; err != nil {
		return fmt.Errorf("bootstrap department: %w", err)
	}
	post := system.Post{ID: 1, TenantID: tenant.ID}
	if err = db.Where("id = ? AND tenant_id = ?", post.ID, tenant.ID).Attrs(system.Post{
		Name: "平台管理员", Code: "platform_admin", Sort: 0, Status: 0,
	}).FirstOrCreate(&post).Error; err != nil {
		return fmt.Errorf("bootstrap post: %w", err)
	}
	role := system.Role{TenantID: tenant.ID, Code: "super_admin"}
	if err = db.Where("tenant_id = ? AND code = ?", tenant.ID, role.Code).Attrs(system.Role{
		Name: "超级管理员", Sort: 1, Status: 0, Type: 1, DataScope: 1, Remark: "系统内置角色",
	}).FirstOrCreate(&role).Error; err != nil {
		return fmt.Errorf("bootstrap role: %w", err)
	}
	if err = db.Where(system.UserRole{UserID: user.ID, RoleID: role.ID}).FirstOrCreate(&system.UserRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
		return fmt.Errorf("bootstrap user role: %w", err)
	}
	if err = db.Where(system.UserPost{UserID: user.ID, PostID: post.ID}).FirstOrCreate(&system.UserPost{UserID: user.ID, PostID: post.ID}).Error; err != nil {
		return fmt.Errorf("bootstrap user post: %w", err)
	}
	dicts := []system.DictData{
		{DictType: "common_status", Label: "开启", Value: "0", Sort: 1, Status: 0, ColorType: "success"},
		{DictType: "common_status", Label: "关闭", Value: "1", Sort: 2, Status: 0, ColorType: "danger"},
		{DictType: "system_user_sex", Label: "男", Value: "1", Sort: 1, Status: 0, ColorType: "primary"},
		{DictType: "system_user_sex", Label: "女", Value: "2", Sort: 2, Status: 0, ColorType: "danger"},
		{DictType: "system_user_sex", Label: "未知", Value: "0", Sort: 3, Status: 0, ColorType: "info"},
	}
	for _, dict := range dicts {
		if err = db.Where("dict_type = ? AND value = ?", dict.DictType, dict.Value).FirstOrCreate(&dict).Error; err != nil {
			return fmt.Errorf("bootstrap dictionary: %w", err)
		}
	}
	return nil
}
