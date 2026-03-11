package middleware

import (
	"bookadmin/global"
	"bookadmin/observability"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}

		start := time.Now()
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		duration := time.Since(start)
		if global.GVA_DB != nil {
			if sqlDB, err := global.GVA_DB.DB(); err == nil {
				stats := sqlDB.Stats()
				observability.SetDBStats(stats.OpenConnections, stats.Idle, stats.InUse)
			}
		}

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("route", route),
			zap.Int("status", c.Writer.Status()),
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("latency", duration),
		}
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}
		if username, exists := c.Get("username"); exists {
			fields = append(fields, zap.Any("username", username))
		}

		observability.RecordHTTPRequest(c.Request.Method, route, c.Writer.Status(), duration)
		global.GVA_LOG.Info("http_request", fields...)
	}
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}
