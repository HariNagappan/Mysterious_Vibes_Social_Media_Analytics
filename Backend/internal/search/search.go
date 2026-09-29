// Package search defines the Elasticsearch/OpenSearch abstraction. The
// pipeline always talks to the Indexer interface; a no-op implementation keeps
// every environment bootable when no search cluster is configured.
package search

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/pulsegraph/pulsegraph-backend/internal/config"
)

// Doc is the indexed representation of a post.
type Doc struct {
	ID         int64     `json:"id"`
	TopicID    int64     `json:"topic_id"`
	Platform   string    `json:"platform"`
	Author     string    `json:"author"`
	Content    string    `json:"content"`
	Language   string    `json:"language"`
	Timestamp  time.Time `json:"timestamp"`
	Engagement int32     `json:"engagement_count"`
}

// Hit is one search result.
type Hit struct {
	ID        int64     `json:"id"`
	Score     float64   `json:"score"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// Indexer abstracts the search backend so Elasticsearch can be swapped for
// OpenSearch (or anything else) without touching the pipeline.
type Indexer interface {
	// Enabled reports whether a real backend is configured.
	Enabled() bool
	// Index stores or refreshes one post document.
	Index(ctx context.Context, doc Doc) error
	// Search runs a full-text query and returns ranked hits.
	Search(ctx context.Context, query string, limit int) ([]Hit, error)
	// Close releases resources.
	Close() error
}

// New builds the configured indexer (Elasticsearch/OpenSearch or no-op).
func New(cfg config.SearchConfig, logger *zap.Logger) Indexer {
	if !cfg.Enabled {
		return &noopIndexer{logger: logger}
	}
	return newESIndexer(cfg, logger)
}
