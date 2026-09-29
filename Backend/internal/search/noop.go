package search

import (
	"context"

	"go.uber.org/zap"
)

// noopIndexer keeps the pipeline bootable without a search cluster: indexing
// is skipped and searches return an empty result set.
type noopIndexer struct {
	logger *zap.Logger
}

// Enabled reports false — no real backend is configured.
func (n *noopIndexer) Enabled() bool { return false }

// Index skips indexing.
func (n *noopIndexer) Index(_ context.Context, doc Doc) error {
	n.logger.Debug("search: disabled; skipping document", zap.Int64("post_id", doc.ID))
	return nil
}

// Search returns an empty result set.
func (n *noopIndexer) Search(_ context.Context, _ string, _ int) ([]Hit, error) {
	return []Hit{}, nil
}

// Close releases resources (none).
func (n *noopIndexer) Close() error { return nil }
