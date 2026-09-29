package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/sentiment"
	"github.com/pulsegraph/pulsegraph-backend/internal/topics"
)

const defaultWindowSpan = 7 * 24 * time.Hour

// Sentiment serves the sentiment analytics endpoints.
type Sentiment struct {
	service   *sentiment.Service
	topicsSvc *topics.Service
	logger    *zap.Logger
}

// NewSentiment wires the sentiment handler.
func NewSentiment(service *sentiment.Service, topicsSvc *topics.Service, logger *zap.Logger) *Sentiment {
	return &Sentiment{service: service, topicsSvc: topicsSvc, logger: logger}
}

// RegisterRoutes mounts the sentiment endpoint.
func (h *Sentiment) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/topics/:id/sentiment", h.Get)
}

// Get handles GET /api/v1/topics/:id/sentiment.
//
// @Summary      Topic sentiment
// @Description  Emotion distribution and sentiment timeline.
// @Tags         sentiment
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Topic id"
// @Success      200  {object}  sentiment.TopicSentiment
// @Failure      404  {object}  validators.ErrorResponse
// @Router       /topics/{id}/sentiment [get]
func (h *Sentiment) Get(c *gin.Context) {
	topicID, ok := validators.ParseIDParam(c, "id")
	if !ok {
		return
	}
	from, to, err := resolveWindow(c, defaultWindowSpan)
	if err != nil {
		validators.RespondValidationError(c, err)
		return
	}
	if !requireTopic(c, h.topicsSvc, topicID) {
		return
	}
	res, err := h.service.GetTopicSentiment(c.Request.Context(), topicID, from, to)
	if err != nil {
		h.logger.Error("http: sentiment failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "could not compute sentiment")
		return
	}
	c.JSON(http.StatusOK, res)
}
