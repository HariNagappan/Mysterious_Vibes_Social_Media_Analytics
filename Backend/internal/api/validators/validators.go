// Package validators centralises request parsing helpers, the canonical JSON
// error envelope and the reusable validation rules shared by every handler.
package validators

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorBody is the canonical error envelope payload.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps every error returned by the API.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// RespondError writes the canonical error envelope.
func RespondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}

// RespondValidationError writes a 400 envelope for a malformed payload.
func RespondValidationError(c *gin.Context, err error) {
	RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
}

// ParseIDParam reads a positive int64 path parameter.
func ParseIDParam(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		RespondError(c, http.StatusBadRequest, "invalid_id", fmt.Sprintf("%s must be a positive integer", name))
		return 0, false
	}
	return id, true
}

// Pagination is the normalized pagination window.
type Pagination struct {
	Limit  int32
	Offset int32
}

// ParsePagination reads limit/offset query parameters with sane bounds.
func ParsePagination(c *gin.Context, defaultLimit, maxLimit int32) Pagination {
	limit := defaultLimit
	offset := int32(0)
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	return Pagination{Limit: limit, Offset: offset}
}

// ParseOptionalTime parses an RFC3339 timestamp query parameter. A missing or
// empty value returns (zero, false, nil); a malformed value returns an error.
func ParseOptionalTime(c *gin.Context, name string) (time.Time, bool, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return time.Time{}, false, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("%s must be an RFC3339 timestamp (e.g. 2026-09-01T12:00:00Z)", name)
	}
	return t.UTC(), true, nil
}

// ValidRole reports whether role is one of the supported user roles.
func ValidRole(role string) bool {
	switch role {
	case "admin", "analyst", "viewer":
		return true
	default:
		return false
	}
}

// ValidPlatform reports whether platform is a supported social platform.
func ValidPlatform(platform string) bool {
	switch platform {
	case "twitter", "telegram", "instagram", "facebook", "reddit", "youtube":
		return true
	default:
		return false
	}
}
