package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pulsegraph/pulsegraph-backend/internal/cache"
)

// Health serves liveness and readiness probes.
type Health struct {
	db    *pgxpool.Pool
	cache *cache.Cache
}

// NewHealth wires the health handler.
func NewHealth(db *pgxpool.Pool, redisCache *cache.Cache) *Health {
	return &Health{db: db, cache: redisCache}
}

// Healthz handles GET /healthz (liveness).
func (h *Health) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz handles GET /readyz (checks PostgreSQL and Redis).
func (h *Health) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	checks := gin.H{"database": "ok", "redis": "ok"}
	healthy := true

	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = "unavailable"
		healthy = false
	}
	if err := h.cache.Ping(ctx); err != nil {
		checks["redis"] = "unavailable"
		healthy = false
	}

	if !healthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "checks": checks})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready", "checks": checks})
}
