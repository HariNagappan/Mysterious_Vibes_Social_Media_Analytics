// Command scheduler runs PulseGraph's periodic work:
//
//   - the analytics cycle (trend recomputation, demographic snapshots, graph
//     rebuild + influence scoring) for every monitored topic, and
//   - in demo mode (DEMO_INGEST_ENABLED=true), the built-in ingestion
//     generator that keeps the full pipeline flowing end-to-end without
//     external credentials.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/analytics"
	"github.com/pulsegraph/pulsegraph-backend/internal/cache"
	"github.com/pulsegraph/pulsegraph-backend/internal/config"
	"github.com/pulsegraph/pulsegraph-backend/internal/database"
	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
	"github.com/pulsegraph/pulsegraph-backend/internal/demographics"
	"github.com/pulsegraph/pulsegraph-backend/internal/graph"
	"github.com/pulsegraph/pulsegraph-backend/internal/ingestion"
	"github.com/pulsegraph/pulsegraph-backend/internal/ingestion/connectors/telegram"
	"github.com/pulsegraph/pulsegraph-backend/internal/ingestion/connectors/twitter"
	"github.com/pulsegraph/pulsegraph-backend/internal/logger"
	"github.com/pulsegraph/pulsegraph-backend/internal/pipeline"
	"github.com/pulsegraph/pulsegraph-backend/internal/trends"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "scheduler: fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log, err := logger.New("scheduler", cfg.App.Env, os.Getenv("LOG_LEVEL"))
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

	redisCache, err := cache.New(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, log)
	if err != nil {
		return err
	}
	defer func() { _ = redisCache.Close() }()

	queries := generated.New(pool)
	producer := pipeline.NewRedisProducer(redisCache.Client(), cfg.Pipeline.StreamPrefix, log)
	trendsService := trends.NewService(trends.NewRepository(queries), log)
	demographicsService := demographics.NewService(demographics.NewRepository(queries))
	graphBuilder := graph.NewBuilder(queries, log)

	registry := ingestion.NewRegistry()
	registry.Register(twitter.NewClient(os.Getenv("TWITTER_BEARER_TOKEN"), log))
	registry.Register(telegram.NewClient(os.Getenv("TELEGRAM_BOT_TOKEN"), log))

	metricsAddr := os.Getenv("SCHEDULER_METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = ":9092"
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Info("scheduler: metrics endpoint listening", zap.String("addr", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			log.Warn("scheduler: metrics server stopped", zap.Error(err))
		}
	}()

	log.Info("scheduler: starting",
		zap.Duration("analytics_interval", cfg.Demo.ScheduleEvery),
		zap.Bool("demo_ingest_enabled", cfg.Demo.IngestEnabled))

	// Kick off one cycle immediately so a fresh deployment starts producing
	// analytics as soon as it comes up.
	runAnalyticsCycle(ctx, log, queries, trendsService, demographicsService, graphBuilder, redisCache)
	if cfg.Demo.IngestEnabled {
		runIngestionCycle(ctx, log, queries, registry, producer)
	}

	schedule := time.NewTicker(cfg.Demo.ScheduleEvery)
	defer schedule.Stop()

	var ingest <-chan time.Time
	if cfg.Demo.IngestEnabled {
		ingestTicker := time.NewTicker(cfg.Demo.IngestInterval)
		defer ingestTicker.Stop()
		ingest = ingestTicker.C
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("scheduler: stopped cleanly")
			return nil
		case <-schedule.C:
			runAnalyticsCycle(ctx, log, queries, trendsService, demographicsService, graphBuilder, redisCache)
		case <-ingest:
			runIngestionCycle(ctx, log, queries, registry, producer)
		}
	}
}

func runAnalyticsCycle(
	ctx context.Context,
	log *zap.Logger,
	queries *generated.Queries,
	trendsService *trends.Service,
	demographicsService *demographics.Service,
	graphBuilder *graph.Builder,
	redisCache *cache.Cache,
) {
	topicsList, err := queries.ListTopics(ctx, generated.ListTopicsParams{Limit: 100, Offset: 0})
	if err != nil {
		log.Warn("scheduler: list topics failed", zap.Error(err))
		return
	}
	for _, topic := range topicsList {
		if ctx.Err() != nil {
			return
		}
		if _, err := trendsService.ComputeAndStore(ctx, topic.ID); err != nil {
			analytics.IncPipelineFailure("trends")
			log.Warn("scheduler: trend computation failed", zap.Int64("topic_id", topic.ID), zap.Error(err))
		} else {
			analytics.IncAnalyticsRun("trends")
		}

		if _, err := demographicsService.ComputeAndStore(ctx, topic.ID); err != nil {
			analytics.IncPipelineFailure("demographics")
			log.Warn("scheduler: demographics computation failed", zap.Int64("topic_id", topic.ID), zap.Error(err))
		} else {
			analytics.IncAnalyticsRun("demographics")
		}

		if _, _, err := graphBuilder.BuildForTopic(ctx, topic.ID); err != nil {
			analytics.IncPipelineFailure("graph")
			log.Warn("scheduler: graph rebuild failed", zap.Int64("topic_id", topic.ID), zap.Error(err))
		} else {
			analytics.IncAnalyticsRun("graph")
		}

		if err := redisCache.Delete(ctx, fmt.Sprintf("dashboard:%d", topic.ID)); err != nil {
			log.Debug("scheduler: dashboard cache invalidation failed", zap.Error(err))
		}
	}
	if len(topicsList) > 0 {
		log.Info("scheduler: analytics cycle complete", zap.Int("topics", len(topicsList)))
	}
}

func runIngestionCycle(
	ctx context.Context,
	log *zap.Logger,
	queries *generated.Queries,
	registry *ingestion.Registry,
	producer pipeline.Producer,
) {
	topicsList, err := queries.ListTopics(ctx, generated.ListTopicsParams{Limit: 20, Offset: 0})
	if err != nil {
		log.Warn("scheduler: list topics failed", zap.Error(err))
		return
	}

	total := 0
	for _, topic := range topicsList {
		for _, platform := range registry.Platforms() {
			if ctx.Err() != nil {
				return
			}
			connector, ok := registry.Get(platform)
			if !ok {
				continue
			}
			rawPosts, err := connector.Fetch(ctx, ingestion.FetchOptions{
				TopicID:  topic.ID,
				Keywords: topic.Keywords,
				Limit:    20,
			})
			if err != nil {
				log.Warn("scheduler: connector fetch failed", zap.String("platform", platform), zap.Error(err))
				continue
			}

			ingested := 0
			for _, rawPost := range rawPosts {
				normalized, err := ingestion.Normalize(rawPost, topic.ID)
				if err != nil {
					continue
				}
				post, err := queries.InsertPost(ctx, normalized.ToInsertParams())
				if err != nil {
					log.Debug("scheduler: insert post failed", zap.Error(err))
					continue
				}
				event, err := pipeline.NewEvent(pipeline.EventPostIngested, pipeline.PostIngestedPayload{
					PostID:   post.ID,
					TopicID:  topic.ID,
					Platform: post.Platform,
				})
				if err != nil {
					continue
				}
				if err := producer.Publish(ctx, pipeline.SubjectForEvent(pipeline.EventPostIngested), event); err != nil {
					log.Warn("scheduler: publish failed", zap.Error(err))
					continue
				}
				analytics.IncPostsIngested()
				ingested++
				total++
			}
			if ingested > 0 {
				log.Info("scheduler: ingested posts",
					zap.String("platform", platform),
					zap.Int("count", ingested),
					zap.Int64("topic_id", topic.ID))
			}
		}
	}
	if total > 0 {
		log.Info("scheduler: ingestion cycle complete", zap.Int("posts", total))
	}
}
