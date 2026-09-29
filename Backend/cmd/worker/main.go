// Command worker consumes the Redis Streams pipeline and executes the AI
// processing stages: sentiment scoring, search indexing and embedding
// generation. Each stage publishes the next stage's event on success;
// failures are retried and ultimately dead-lettered by the consumer.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/ai"
	"github.com/pulsegraph/pulsegraph-backend/internal/analytics"
	"github.com/pulsegraph/pulsegraph-backend/internal/cache"
	"github.com/pulsegraph/pulsegraph-backend/internal/config"
	"github.com/pulsegraph/pulsegraph-backend/internal/database"
	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
	"github.com/pulsegraph/pulsegraph-backend/internal/logger"
	"github.com/pulsegraph/pulsegraph-backend/internal/pipeline"
	"github.com/pulsegraph/pulsegraph-backend/internal/search"
	"github.com/pulsegraph/pulsegraph-backend/internal/sentiment"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "worker: fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log, err := logger.New("worker", cfg.App.Env, os.Getenv("LOG_LEVEL"))
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
	sentimentService := sentiment.NewService(
		sentiment.NewRepository(queries),
		sentiment.NewClient(cfg.ML.SentimentURL, cfg.ML.Timeout, log),
		log,
	)
	embeddingClient := ai.NewEmbeddingClient(cfg.ML.EmbeddingURL, cfg.ML.Timeout, log)
	indexer := search.New(cfg.Search, log)
	defer func() { _ = indexer.Close() }()

	// Prometheus endpoint for pipeline-stage metrics.
	metricsAddr := os.Getenv("WORKER_METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = ":9091"
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Info("worker: metrics endpoint listening", zap.String("addr", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			log.Warn("worker: metrics server stopped", zap.Error(err))
		}
	}()

	// Stage 1: post.ingested → sentiment scoring + search indexing.
	postsHandler := func(ctx context.Context, evt pipeline.Event) error {
		if evt.Type != pipeline.EventPostIngested {
			return nil
		}
		var payload pipeline.PostIngestedPayload
		if err := evt.Decode(&payload); err != nil {
			return err
		}
		res, err := sentimentService.ProcessPost(ctx, payload.PostID)
		if err != nil {
			analytics.IncPipelineFailure("sentiment")
			return fmt.Errorf("worker: sentiment stage: %w", err)
		}
		analytics.IncSentimentProcessed()

		post, err := queries.GetPostByID(ctx, payload.PostID)
		if err != nil {
			return fmt.Errorf("worker: load post for indexing: %w", err)
		}
		if err := indexer.Index(ctx, search.Doc{
			ID:         post.ID,
			TopicID:    post.TopicID,
			Platform:   post.Platform,
			Author:     post.AuthorReference,
			Content:    post.Content,
			Language:   post.Language,
			Timestamp:  post.Timestamp,
			Engagement: post.EngagementCount,
		}); err != nil {
			log.Warn("worker: search indexing failed", zap.Int64("post_id", post.ID), zap.Error(err))
		}

		next, err := pipeline.NewEvent(pipeline.EventSentimentCompleted, pipeline.SentimentCompletedPayload{
			PostID:  payload.PostID,
			TopicID: payload.TopicID,
			Emotion: res.Emotion,
			Score:   res.Score,
		})
		if err != nil {
			return err
		}
		return producer.Publish(ctx, pipeline.SubjectForEvent(pipeline.EventSentimentCompleted), next)
	}

	// Stage 2: sentiment.completed → embedding generation + storage.
	sentimentHandler := func(ctx context.Context, evt pipeline.Event) error {
		if evt.Type != pipeline.EventSentimentCompleted {
			return nil
		}
		var payload pipeline.SentimentCompletedPayload
		if err := evt.Decode(&payload); err != nil {
			return err
		}
		post, err := queries.GetPostByID(ctx, payload.PostID)
		if err != nil {
			return fmt.Errorf("worker: load post for embedding: %w", err)
		}
		vector, modelVersion, err := embeddingClient.Embed(ctx, post.Content)
		if err != nil {
			analytics.IncPipelineFailure("embedding")
			return fmt.Errorf("worker: embedding stage: %w", err)
		}
		if _, err := queries.UpsertEmbedding(ctx, generated.UpsertEmbeddingParams{
			PostID:       post.ID,
			ModelVersion: modelVersion,
			Embedding:    ai.FormatVector(vector),
		}); err != nil {
			return fmt.Errorf("worker: store embedding: %w", err)
		}
		analytics.IncEmbeddingGenerated()

		next, err := pipeline.NewEvent(pipeline.EventEmbeddingCompleted, pipeline.EmbeddingCompletedPayload{
			PostID:  post.ID,
			TopicID: payload.TopicID,
			Dims:    len(vector),
		})
		if err != nil {
			return err
		}
		return producer.Publish(ctx, pipeline.SubjectForEvent(pipeline.EventEmbeddingCompleted), next)
	}

	// Stage 3: embedding.completed → trail marker (feeds downstream consumers).
	analyticsHandler := func(_ context.Context, evt pipeline.Event) error {
		if evt.Type != pipeline.EventEmbeddingCompleted {
			return nil
		}
		var payload pipeline.EmbeddingCompletedPayload
		if err := evt.Decode(&payload); err != nil {
			return err
		}
		log.Debug("worker: embedding trail event",
			zap.Int64("post_id", payload.PostID),
			zap.Int("dims", payload.Dims))
		return nil
	}

	group := cfg.Pipeline.ConsumerGroup
	type consumerSpec struct {
		name     string
		consumer *pipeline.Consumer
	}
	consumers := []consumerSpec{
		{
			name:     "posts",
			consumer: pipeline.NewConsumer(redisCache.Client(), producer.StreamName(pipeline.SubjectPosts), group, "worker-posts", postsHandler, log, cfg.Pipeline.BatchSize, cfg.Pipeline.Block, int64(cfg.Pipeline.MaxRetries)),
		},
		{
			name:     "sentiment",
			consumer: pipeline.NewConsumer(redisCache.Client(), producer.StreamName(pipeline.SubjectSentiment), group, "worker-sentiment", sentimentHandler, log, cfg.Pipeline.BatchSize, cfg.Pipeline.Block, int64(cfg.Pipeline.MaxRetries)),
		},
		{
			name:     "analytics",
			consumer: pipeline.NewConsumer(redisCache.Client(), producer.StreamName(pipeline.SubjectAnalytics), group, "worker-analytics", analyticsHandler, log, cfg.Pipeline.BatchSize, cfg.Pipeline.Block, int64(cfg.Pipeline.MaxRetries)),
		},
	}

	var wg sync.WaitGroup
	for _, spec := range consumers {
		wg.Add(1)
		go func(name string, consumer *pipeline.Consumer) {
			defer wg.Done()
			log.Info("worker: consumer started", zap.String("stage", name))
			if err := consumer.Run(ctx); err != nil {
				log.Error("worker: consumer stopped with error", zap.String("stage", name), zap.Error(err))
			}
		}(spec.name, spec.consumer)
	}
	log.Info("worker: consumers running", zap.String("group", group))

	<-ctx.Done()
	wg.Wait()
	log.Info("worker: stopped cleanly")
	return nil
}
