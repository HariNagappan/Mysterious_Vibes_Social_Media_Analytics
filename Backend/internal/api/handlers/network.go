package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/graph"
	"github.com/pulsegraph/pulsegraph-backend/internal/topics"
)

const (
	defaultNodeLimit = 200
	defaultEdgeLimit = 400
)

// Network serves the influence-network endpoints.
type Network struct {
	service   *graph.Service
	topicsSvc *topics.Service
	logger    *zap.Logger
}

// NewNetwork wires the network handler.
func NewNetwork(service *graph.Service, topicsSvc *topics.Service, logger *zap.Logger) *Network {
	return &Network{service: service, topicsSvc: topicsSvc, logger: logger}
}

// RegisterRoutes mounts the network endpoint.
func (h *Network) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/topics/:id/network", h.Get)
}

// Get handles GET /api/v1/topics/:id/network.
//
// @Summary      Influence network
// @Description  Graph nodes, edges, influence scores and communities.
// @Tags         network
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Topic id"
// @Success      200  {object}  graph.Network
// @Failure      404  {object}  validators.ErrorResponse
// @Router       /topics/{id}/network [get]
func (h *Network) Get(c *gin.Context) {
	topicID, ok := validators.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !requireTopic(c, h.topicsSvc, topicID) {
		return
	}
	res, err := h.service.Network(c.Request.Context(), topicID, defaultNodeLimit, defaultEdgeLimit)
	if err != nil {
		h.logger.Error("http: network failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "could not load network")
		return
	}
	c.JSON(http.StatusOK, res)
}
