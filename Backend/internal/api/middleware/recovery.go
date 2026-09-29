package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
)

// Recovery converts panics into the canonical 500 envelope and logs a stack
// trace so the request is never left with a half-written response.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("http: panic recovered",
			zap.Any("panic", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("request_id", c.GetString(RequestIDKey)),
			zap.Stack("stack"),
		)
		if !c.Writer.Written() {
			validators.RespondError(c, http.StatusInternalServerError, "internal_error", "unexpected server error")
		}
		c.Abort()
	})
}
