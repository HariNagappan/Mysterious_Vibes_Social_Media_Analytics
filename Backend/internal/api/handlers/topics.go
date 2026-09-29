package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/topics"
)

// Topics serves topic management endpoints.
type Topics struct {
	service *topics.Service
	logger  *zap.Logger
}

// NewTopics wires the topics handler.
func NewTopics(service *topics.Service, logger *zap.Logger) *Topics {
	return &Topics{service: service, logger: logger}
}

// RegisterRoutes mounts the topics endpoints.
func (h *Topics) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/topics", h.Create)
	group.GET("/topics", h.List)
}

// Create handles POST /api/v1/topics.
//
// @Summary      Create a monitored topic
// @Description  Registers a new conversation to monitor (e.g. "Flood misinformation").
// @Tags         topics
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload  body      topics.CreateRequest  true  "Topic payload"
// @Success      201      {object}  generated.Topic
// @Failure      400      {object}  validators.ErrorResponse
// @Router       /topics [post]
func (h *Topics) Create(c *gin.Context) {
	var req topics.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validators.RespondValidationError(c, err)
		return
	}
	topic, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, topics.ErrInvalidInput) {
			validators.RespondError(c, http.StatusBadRequest, "invalid_request", "topic name is required")
			return
		}
		h.logger.Error("http: create topic failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "could not create topic")
		return
	}
	c.JSON(http.StatusCreated, topic)
}

// List handles GET /api/v1/topics.
//
// @Summary      List monitored topics
// @Tags         topics
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  topics.ListResponse
// @Router       /topics [get]
func (h *Topics) List(c *gin.Context) {
	page := validators.ParsePagination(c, 20, 100)
	res, err := h.service.List(c.Request.Context(), page.Limit, page.Offset)
	if err != nil {
		h.logger.Error("http: list topics failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "could not list topics")
		return
	}
	c.JSON(http.StatusOK, res)
}
