package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/analytics"
	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
)

// Dashboard serves the complete-intelligence endpoint.
type Dashboard struct {
	service *analytics.Service
	logger  *zap.Logger
}

// NewDashboard wires the dashboard handler.
func NewDashboard(service *analytics.Service, logger *zap.Logger) *Dashboard {
	return &Dashboard{service: service, logger: logger}
}

// RegisterRoutes mounts the dashboard endpoint.
func (h *Dashboard) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/topics/:id/dashboard", h.Get)
}

// Get handles GET /api/v1/topics/:id/dashboard.
//
// @Summary      Complete topic intelligence
// @Description  Timeline + sentiment + demographics + trends + network + AI summary in one payload.
// @Tags         dashboard
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Topic id"
// @Success      200  {object}  analytics.Dashboard
// @Failure      404  {object}  validators.ErrorResponse
// @Router       /topics/{id}/dashboard [get]
func (h *Dashboard) Get(c *gin.Context) {
	topicID, ok := validators.ParseIDParam(c, "id")
	if !ok {
		return
	}
	res, err := h.service.Get(c.Request.Context(), topicID)
	if err != nil {
		if errors.Is(err, analytics.ErrTopicNotFound) {
			validators.RespondError(c, http.StatusNotFound, "topic_not_found", "topic was not found")
			return
		}
		h.logger.Error("http: dashboard failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "could not build dashboard")
		return
	}
	c.JSON(http.StatusOK, res)
}
