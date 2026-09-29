package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/topics"
	"github.com/pulsegraph/pulsegraph-backend/internal/trends"
)

// Trends serves the trend analytics endpoints.
type Trends struct {
	service   *trends.Service
	topicsSvc *topics.Service
	logger    *zap.Logger
}

// NewTrends wires the trends handler.
func NewTrends(service *trends.Service, topicsSvc *topics.Service, logger *zap.Logger) *Trends {
	return &Trends{service: service, topicsSvc: topicsSvc, logger: logger}
}

// RegisterRoutes mounts the trends endpoint.
func (h *Trends) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/topics/:id/trends", h.Get)
}

// Get handles GET /api/v1/topics/:id/trends.
//
// @Summary      Rising trends
// @Description  Rising keywords, growth scores and the topic-level trend series.
// @Tags         trends
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Topic id"
// @Success      200  {object}  trends.TopicTrends
// @Failure      404  {object}  validators.ErrorResponse
// @Router       /topics/{id}/trends [get]
func (h *Trends) Get(c *gin.Context) {
	topicID, ok := validators.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !requireTopic(c, h.topicsSvc, topicID) {
		return
	}
	res, err := h.service.Get(c.Request.Context(), topicID)
	if err != nil {
		h.logger.Error("http: trends failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "could not load trends")
		return
	}
	c.JSON(http.StatusOK, res)
}
