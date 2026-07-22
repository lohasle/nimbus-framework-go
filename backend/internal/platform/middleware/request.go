package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RequestLog struct {
	TraceID   string
	TenantID  uint64
	UserID    uint64
	Method    string
	Path      string
	Status    int
	Duration  int64
	IP        string
	UserAgent string
}

type RequestLogRecorder func(RequestLog)

func RequestContext(recorders ...RequestLogRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
		latency := time.Since(started)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		slog.Info("http request", "trace_id", traceID, "method", c.Request.Method, "path", path, "status", c.Writer.Status(), "latency", latency)
		record := RequestLog{
			TraceID: traceID, TenantID: c.GetUint64("tenant_id"), UserID: c.GetUint64("user_id"),
			Method: c.Request.Method, Path: path, Status: c.Writer.Status(), Duration: latency.Milliseconds(),
			IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
		}
		for _, recorder := range recorders {
			recorder(record)
		}
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
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,tenant-id,visit-tenant-id,X-Trace-ID,Cache-Control,Pragma,X-Api-Encrypt")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Vary", "Origin")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
