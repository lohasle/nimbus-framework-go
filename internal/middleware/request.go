package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
		slog.Info("http request", "trace_id", traceID, "method", c.Request.Method, "path", c.FullPath(), "status", c.Writer.Status(), "latency", time.Since(started))
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,tenant-id,X-Trace-ID")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Vary", "Origin")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
