package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/auth"
)

// Context keys for the authenticated principal.
const (
	// ContextUserIDKey stores the authenticated user id.
	ContextUserIDKey = "auth_user_id"
	// ContextRoleKey stores the authenticated user role.
	ContextRoleKey = "auth_role"
	// ContextEmailKey stores the authenticated user email.
	ContextEmailKey = "auth_email"
)

// RequireAuth validates the bearer token and stores the principal claims in
// the request context.
func RequireAuth(tokens *auth.Manager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			validators.RespondError(c, http.StatusUnauthorized, "missing_token", "an Authorization: Bearer <token> header is required")
			c.Abort()
			return
		}
		raw := strings.TrimSpace(header[len("bearer "):])
		claims, err := tokens.ParseToken(raw)
		if err != nil {
			logger.Debug("http: rejected token", zap.Error(err))
			validators.RespondError(c, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
			c.Abort()
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Set(ContextEmailKey, claims.Email)
		c.Next()
	}
}

// RequireRole restricts a route group to the given roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := c.GetString(ContextRoleKey)
		if _, ok := allowed[role]; !ok {
			validators.RespondError(c, http.StatusForbidden, "forbidden", "your role does not permit this action")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ContextUserID extracts the authenticated user id from the gin context.
func ContextUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := value.(int64)
	return id, ok
}
