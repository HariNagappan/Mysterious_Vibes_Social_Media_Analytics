package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/topics"
)

var errInvalidWindow = errors.New("from must be earlier than to")

// resolveWindow reads optional RFC3339 from/to query parameters with a
// default span ending now.
func resolveWindow(c *gin.Context, defaultSpan time.Duration) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	to := now
	if parsed, ok, err := validators.ParseOptionalTime(c, "to"); err != nil {
		return time.Time{}, time.Time{}, err
	} else if ok {
		to = parsed
	}
	from := to.Add(-defaultSpan)
	if parsed, ok, err := validators.ParseOptionalTime(c, "from"); err != nil {
		return time.Time{}, time.Time{}, err
	} else if ok {
		from = parsed
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, errInvalidWindow
	}
	return from, to, nil
}

// requireTopic writes a 404 envelope when the topic does not exist.
func requireTopic(c *gin.Context, topicsSvc *topics.Service, topicID int64) bool {
	if _, err := topicsSvc.Get(c.Request.Context(), topicID); err != nil {
		validators.RespondError(c, http.StatusNotFound, "topic_not_found", "topic was not found")
		return false
	}
	return true
}
