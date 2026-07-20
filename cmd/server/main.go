package main

import (
	"log/slog"
	"os"

	_ "github.com/lohasle/nimbus-framework-go/docs"
	"github.com/lohasle/nimbus-framework-go/internal/config"
	"github.com/lohasle/nimbus-framework-go/internal/database"
	"github.com/lohasle/nimbus-framework-go/internal/router"
	"github.com/lohasle/nimbus-framework-go/internal/system"
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
	service := system.NewService(db, cfg)
	if err = router.New(system.NewHandler(service)).Run(cfg.HTTPAddr); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
