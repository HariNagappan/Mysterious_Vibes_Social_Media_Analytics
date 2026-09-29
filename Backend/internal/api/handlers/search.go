package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/validators"
	"github.com/pulsegraph/pulsegraph-backend/internal/search"
)

// Search serves full-text post search backed by the search abstraction.
type Search struct {
	indexer search.Indexer
	logger  *zap.Logger
}

// NewSearch wires the search handler.
func NewSearch(indexer search.Indexer, logger *zap.Logger) *Search {
	return &Search{indexer: indexer, logger: logger}
}

// RegisterRoutes mounts the search endpoint.
func (h *Search) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/search/posts", h.SearchPosts)
}

// SearchPosts handles GET /api/v1/search/posts.
//
// @Summary      Full-text post search
// @Description  Multi-match search over content and author (requires SEARCH_ENABLED).
// @Tags         search
// @Security     BearerAuth
// @Produce      json
// @Param        q      query  string  true   "Search query"
// @Param        limit  query  int     false  "Result limit"
// @Success      200  {object}  map[string]any
// @Router       /search/posts [get]
func (h *Search) SearchPosts(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		validators.RespondError(c, http.StatusBadRequest, "missing_query", "the q query parameter is required")
		return
	}
	page := validators.ParsePagination(c, 20, 100)

	if !h.indexer.Enabled() {
		c.JSON(http.StatusOK, gin.H{
			"search_enabled": false,
			"hits":           []search.Hit{},
			"message":        "search backend is not configured (set SEARCH_ENABLED=true)",
		})
		return
	}

	hits, err := h.indexer.Search(c.Request.Context(), query, int(page.Limit))
	if err != nil {
		h.logger.Error("http: search failed", zap.Error(err))
		validators.RespondError(c, http.StatusInternalServerError, "internal_error", "search failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"search_enabled": true, "hits": hits})
}
