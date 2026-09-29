package topics

import (
	"context"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Repository persists topics via the generated query layer.
type Repository struct {
	queries *generated.Queries
}

// NewRepository wraps the generated queries.
func NewRepository(queries *generated.Queries) *Repository {
	return &Repository{queries: queries}
}

// Create inserts a topic.
func (r *Repository) Create(ctx context.Context, arg generated.CreateTopicParams) (generated.Topic, error) {
	return r.queries.CreateTopic(ctx, arg)
}

// Get loads one topic.
func (r *Repository) Get(ctx context.Context, id int64) (generated.Topic, error) {
	return r.queries.GetTopicByID(ctx, id)
}

// List returns topics with pagination.
func (r *Repository) List(ctx context.Context, limit, offset int32) ([]generated.Topic, error) {
	return r.queries.ListTopics(ctx, generated.ListTopicsParams{Limit: limit, Offset: offset})
}

// Count returns the total number of topics.
func (r *Repository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountTopics(ctx)
}
