package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/demographics"
	"github.com/pulsegraph/pulsegraph-backend/internal/topics"
)

// Demographics serves the audience analytics endpoints.
type Demographics struct {
	service   *demographics.Service
	topicsSvc *topics.Service
	logger    *zap.Logger
}

// NewDemographics wires the demographics handler.
func NewDemographics(service *demographics.Service, topicsSvc *topics.Service, logger *zap.Logger) *Demographics {
	return &Demographics{service: service, topicsSvc: topicsSvc, logger: logger}
}

// RegisterRoutes mounts the demographics endpoint.
func (h *Demographics) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/topics/:id/demographics", h.Get)
}

// Get handles GET /api/v1/topics/:id/demographics.
//
// @Summary      Audience demographics
// @Description  Aggregated audience groups, language distribution and regional patterns.
// @Tags         demographics
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Topic id"
// @Success      200  {object}  demographics.Snapshot
// @Failure      404  {object}  validators.ErrorResponse
// @Router       /topics/{id}/demographics [get]
func (h *Demographics) Get(c *gin.Context) {
	topicID, ok := validators.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !requireTopic(c, h.topicsSvc, topicID) {
		return
	}
	snapshot, err := h.service.Get(c.Request.Context(), topicID)
	if err != nil {
		h.logger.Error("http: demographics failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "could not load demographics")
		return
	}
	c.JSON(http.StatusOK, snapshot)
}
