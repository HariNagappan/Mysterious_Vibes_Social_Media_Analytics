// Command api serves the PulseGraph REST API.
//
// It owns the HTTP surface: authentication, topics, timeline, sentiment,
// demographics, trends, network, dashboard, search, plus operational
// endpoints (/healthz, /readyz, /metrics, /swagger). It applies embedded
// database migrations on boot when RUN_MIGRATIONS=true (default in dev), and
// supports `-migrate-only` for deployment pipelines.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/ai"
	"github.com/pulsegraph/pulsegraph-backend/internal/analytics"
	"github.com/pulsegraph/pulsegraph-backend/internal/api/handlers"
	"github.com/pulsegraph/pulsegraph-backend/internal/api/routes"
	"github.com/pulsegraph/pulsegraph-backend/internal/auth"
	"github.com/pulsegraph/pulsegraph-backend/internal/cache"
	"github.com/pulsegraph/pulsegraph-backend/internal/config"
	"github.com/pulsegraph/pulsegraph-backend/internal/database"
	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
	"github.com/pulsegraph/pulsegraph-backend/internal/demographics"
	"github.com/pulsegraph/pulsegraph-backend/internal/graph"
	"github.com/pulsegraph/pulsegraph-backend/internal/logger"
	"github.com/pulsegraph/pulsegraph-backend/internal/search"
	"github.com/pulsegraph/pulsegraph-backend/internal/sentiment"
	"github.com/pulsegraph/pulsegraph-backend/internal/timeline"
	"github.com/pulsegraph/pulsegraph-backend/internal/topics"
	"github.com/pulsegraph/pulsegraph-backend/internal/trends"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "apply database migrations and exit")
	flag.Parse()

	if err := run(*migrateOnly); err != nil {
		fmt.Fprintf(os.Stderr, "api: fatal: %v\n", err)
		os.Exit(1)
	}
}

func run(migrateOnly bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, err := logger.New("api", cfg.App.Env, os.Getenv("LOG_LEVEL"))
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Postgres.URL, cfg.Postgres.MaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.App.RunMigrations || migrateOnly {
		applied, err := database.Migrate(ctx, pool)
		if err != nil {
			return err
		}
		log.Info("database: migrations applied", zap.Strings("versions", applied))
	}
	if migrateOnly {
		log.Info("api: migrate-only mode complete")
		return nil
	}

	redisCache, err := cache.New(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, log)
	if err != nil {
		return err
	}
	defer func() { _ = redisCache.Close() }()

	queries := generated.New(pool)

	// ── Domain wiring (manual dependency injection) ─────────────────────────
	jwtManager := auth.NewManager(cfg.JWT.Secret, cfg.JWT.TTL, cfg.JWT.Issuer)
	authService := auth.NewService(auth.NewRepository(queries), jwtManager, log)
	authHandler := auth.NewHandler(authService, log)

	topicsService := topics.NewService(topics.NewRepository(queries))
	timelineService := timeline.NewService(timeline.NewRepository(queries))
	sentimentService := sentiment.NewService(
		sentiment.NewRepository(queries),
		sentiment.NewClient(cfg.ML.SentimentURL, cfg.ML.Timeout, log),
		log,
	)
	demographicsService := demographics.NewService(demographics.NewRepository(queries))
	trendsService := trends.NewService(trends.NewRepository(queries), log)
	graphService := graph.NewService(queries)
	aggregator := analytics.NewAggregator(queries)
	summarizer := ai.NewSummarizer(cfg.ML.SummarizerURL, cfg.ML.Timeout, log)
	dashboardService := analytics.NewService(
		topicsService,
		timelineService,
		sentimentService,
		demographicsService,
		trendsService,
		graphService,
		summarizer,
		aggregator,
		redisCache,
		cfg.Redis.CacheTTL,
		log,
	)

	indexer := search.New(cfg.Search, log)
	defer func() { _ = indexer.Close() }()

	engine := routes.New(routes.Options{
		Config:       cfg,
		Logger:       log,
		Auth:         authHandler,
		JWT:          jwtManager,
		Health:       handlers.NewHealth(pool, redisCache),
		Topics:       handlers.NewTopics(topicsService, log),
		Timeline:     handlers.NewTimeline(timelineService, log),
		Sentiment:    handlers.NewSentiment(sentimentService, topicsService, log),
		Demographics: handlers.NewDemographics(demographicsService, topicsService, log),
		Trends:       handlers.NewTrends(trendsService, topicsService, log),
		Network:      handlers.NewNetwork(graphService, topicsService, log),
		Dashboard:    handlers.NewDashboard(dashboardService, log),
		Search:       handlers.NewSearch(indexer, log),
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("api: listening", zap.Int("port", cfg.App.Port), zap.String("env", cfg.App.Env))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("api: shutdown signal received")
	case err := <-serverErr:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("api: graceful shutdown failed", zap.Error(err))
		return err
	}
	log.Info("api: stopped cleanly")
	return nil
}
