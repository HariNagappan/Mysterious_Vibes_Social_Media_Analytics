package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/timeline"
)

// Timeline serves the timeline endpoints.
type Timeline struct {
	service *timeline.Service
	logger  *zap.Logger
}

// NewTimeline wires the timeline handler.
func NewTimeline(service *timeline.Service, logger *zap.Logger) *Timeline {
	return &Timeline{service: service, logger: logger}
}

// RegisterRoutes mounts the timeline endpoint.
func (h *Timeline) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/topics/:id/timeline", h.Get)
}

// Get handles GET /api/v1/topics/:id/timeline.
//
// @Summary      Topic timeline
// @Description  Chronological posts, discussion growth and important events.
// @Tags         timeline
// @Security     BearerAuth
// @Produce      json
// @Param        id      path   int     true   "Topic id"
// @Param        from    query  string  false  "RFC3339 window start"
// @Param        to      query  string  false  "RFC3339 window end"
// @Param        limit   query  int     false  "Page size"
// @Param        offset  query  int     false  "Page offset"
// @Success      200  {object}  timeline.Timeline
// @Failure      404  {object}  validators.ErrorResponse
// @Router       /topics/{id}/timeline [get]
func (h *Timeline) Get(c *gin.Context) {
	topicID, ok := validators.ParseIDParam(c, "id")
	if !ok {
		return
	}
	from, hasFrom, err := validators.ParseOptionalTime(c, "from")
	if err != nil {
		validators.RespondValidationError(c, err)
		return
	}
	to, hasTo, err := validators.ParseOptionalTime(c, "to")
	if err != nil {
		validators.RespondValidationError(c, err)
		return
	}
	page := validators.ParsePagination(c, 100, 500)

	res, err := h.service.Get(c.Request.Context(), timeline.Query{
		TopicID: topicID,
		From:    from,
		HasFrom: hasFrom,
		To:      to,
		HasTo:   hasTo,
		Limit:   page.Limit,
		Offset:  page.Offset,
	})
	if err != nil {
		switch {
		case errors.Is(err, timeline.ErrTopicNotFound):
			validators.RespondError(c, http.StatusNotFound, "topic_not_found", "topic was not found")
		case errors.Is(err, timeline.ErrInvalidWindow):
			validators.RespondError(c, http.StatusBadRequest, "invalid_window", "from must be earlier than to")
		default:
			h.logger.Error("http: timeline failed", zap.Error(err))
			validators.RespondError(c, http.StatusInternalServerError, "internal_error", "could not load timeline")
		}
		return
	}
	c.JSON(http.StatusOK, res)
}
