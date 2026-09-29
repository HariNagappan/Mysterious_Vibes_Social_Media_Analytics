package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestIDHeader is the header carrying the request correlation id.
const RequestIDHeader = "X-Request-ID"

// RequestIDKey is the gin context key holding the correlation id.
const RequestIDKey = "request_id"

// RequestID assigns a correlation id to every request (reusing an inbound
// X-Request-ID when present) and echoes it back on the response so a single
// request can be traced across logs.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "req-fallback"
	}
	return hex.EncodeToString(buf)
}
