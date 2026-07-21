package main

import (
	"log/slog"
	"os"

	_ "github.com/lohasle/nimbus-framework-go/docs"
	"github.com/lohasle/nimbus-framework-go/internal/modules/infra"
	"github.com/lohasle/nimbus-framework-go/internal/modules/member"
	"github.com/lohasle/nimbus-framework-go/internal/modules/pay"
	"github.com/lohasle/nimbus-framework-go/internal/modules/system"
	"github.com/lohasle/nimbus-framework-go/internal/platform/config"
	"github.com/lohasle/nimbus-framework-go/internal/platform/database"
	"github.com/lohasle/nimbus-framework-go/internal/platform/router"
)

// @title Nimbus Framework Go API
// @version 1.0
// @description Go modular-monolith scaffold for operations platforms. Default database: MySQL 8.4.
// @BasePath /admin-api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		slog.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	var tenant system.Tenant
	if err = db.Where("name = ?", cfg.BootstrapTenant).First(&tenant).Error; err == nil {
		err = infra.Migrate(db)
	}
	if err == nil {
		err = member.Migrate(db)
	}
	if err == nil {
		err = pay.Migrate(db)
	}
	if err == nil {
		err = infra.Seed(db, tenant.ID)
	}
	if err == nil {
		err = member.Seed(db, tenant.ID)
	}
	if err == nil {
		err = pay.Seed(db, tenant.ID)
	}
	if err != nil {
		slog.Error("module database initialization failed", "error", err)
		os.Exit(1)
	}
	service := system.NewService(db, cfg)
	if err = router.New(system.NewHandler(service), db).Run(cfg.HTTPAddr); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
