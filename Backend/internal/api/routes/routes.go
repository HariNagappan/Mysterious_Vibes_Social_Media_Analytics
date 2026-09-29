// Package routes wires the gin engine: middleware stack, operational
// endpoints and the versioned API surface.
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/api/handlers"
	"github.com/pulsegraph/pulsegraph-backend/internal/api/middleware"
	"github.com/pulsegraph/pulsegraph-backend/internal/api/openapi"
	"github.com/pulsegraph/pulsegraph-backend/internal/auth"
	"github.com/pulsegraph/pulsegraph-backend/internal/config"
)

// Options bundles every handler the router mounts.
type Options struct {
	Config       *config.Config
	Logger       *zap.Logger
	Auth         *auth.Handler
	JWT          *auth.Manager
	Health       *handlers.Health
	Topics       *handlers.Topics
	Timeline     *handlers.Timeline
	Sentiment    *handlers.Sentiment
	Demographics *handlers.Demographics
	Trends       *handlers.Trends
	Network      *handlers.Network
	Dashboard    *handlers.Dashboard
	Search       *handlers.Search
}

// New builds the gin engine with the full middleware stack and routes.
func New(opts Options) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.Logger(opts.Logger),
		middleware.Recovery(opts.Logger),
		middleware.Metrics(),
		middleware.CORS(opts.Config.App.CORSOrigins),
	)

	// Operational endpoints (outside the versioned API).
	engine.GET("/healthz", opts.Health.Healthz)
	engine.GET("/readyz", opts.Health.Readyz)
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	openapi.RegisterRoutes(engine)

	// Public API: authentication.
	api := engine.Group("/api/v1")
	opts.Auth.RegisterRoutes(api)

	// Authenticated API surface.
	secured := api.Group("")
	secured.Use(middleware.RequireAuth(opts.JWT, opts.Logger))

	opts.Topics.RegisterRoutes(secured)
	opts.Timeline.RegisterRoutes(secured)
	opts.Sentiment.RegisterRoutes(secured)
	opts.Demographics.RegisterRoutes(secured)
	opts.Trends.RegisterRoutes(secured)
	opts.Network.RegisterRoutes(secured)
	opts.Dashboard.RegisterRoutes(secured)
	if opts.Search != nil {
		opts.Search.RegisterRoutes(secured)
	}
	return engine
}
