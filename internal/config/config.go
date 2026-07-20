package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr        string
	DBDSN           string
	JWTSecret       string
	TokenTTL        time.Duration
	BootstrapTenant string
	BootstrapUser   string
	BootstrapPass   string
}

func Load() Config {
	return Config{
		HTTPAddr:        env("NIMBUS_HTTP_ADDR", ":58080"),
		DBDSN:           env("NIMBUS_DB_DSN", "nimbus:nimbus_dev@tcp(127.0.0.1:23316)/nimbus_platform_go?charset=utf8mb4&parseTime=True&loc=Local"),
		JWTSecret:       env("NIMBUS_JWT_SECRET", "nimbus-local-development-secret"),
		TokenTTL:        duration("NIMBUS_TOKEN_TTL", 2*time.Hour),
		BootstrapTenant: env("NIMBUS_BOOTSTRAP_TENANT", "Nimbus Framework"),
		BootstrapUser:   env("NIMBUS_BOOTSTRAP_USERNAME", "admin"),
		BootstrapPass:   env("NIMBUS_BOOTSTRAP_PASSWORD", "admin123"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(env(key, fallback.String()))
	if err != nil {
		return fallback
	}
	return value
}
