package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger emits one structured access-log line per request with status,
// latency and correlation id.
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", c.GetString(RequestIDKey)),
		}
		if errs := c.Errors.String(); errs != "" {
			fields = append(fields, zap.String("errors", errs))
		}

		switch {
		case status >= 500:
			logger.Error("http: request completed", fields...)
		case status >= 400:
			logger.Warn("http: request completed", fields...)
		default:
			logger.Info("http: request completed", fields...)
		}
	}
}
