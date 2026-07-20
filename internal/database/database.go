package database

import (
	"fmt"
	"log/slog"

	"github.com/lohasle/nimbus-framework-go/internal/config"
	"github.com/lohasle/nimbus-framework-go/internal/system"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	if err = db.AutoMigrate(&system.Tenant{}, &system.AdminUser{}); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}
	if err = bootstrap(db, cfg); err != nil {
		return nil, err
	}
	slog.Info("mysql initialized", "database", "nimbus_platform_go")
	return db, nil
}

func bootstrap(db *gorm.DB, cfg config.Config) error {
	tenant := system.Tenant{Name: cfg.BootstrapTenant, Status: 0}
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
	return nil
}
