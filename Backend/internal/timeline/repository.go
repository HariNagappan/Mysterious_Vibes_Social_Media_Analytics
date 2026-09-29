package timeline

import (
	"context"
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Repository reads timeline projections via the generated query layer.
type Repository struct {
	queries *generated.Queries
}

// NewRepository wraps the generated queries.
func NewRepository(queries *generated.Queries) *Repository {
	return &Repository{queries: queries}
}

// GetTopic loads the monitored topic.
func (r *Repository) GetTopic(ctx context.Context, id int64) (generated.Topic, error) {
	return r.queries.GetTopicByID(ctx, id)
}

// ListPosts returns the chronological post feed.
func (r *Repository) ListPosts(ctx context.Context, arg generated.ListPostsByTopicParams) ([]generated.SocialPost, error) {
	return r.queries.ListPostsByTopic(ctx, arg)
}

// PostsPerBucket returns the discussion growth series.
func (r *Repository) PostsPerBucket(ctx context.Context, arg generated.PostsPerBucketParams) ([]generated.PostsBucket, error) {
	return r.queries.PostsPerBucket(ctx, arg)
}

// TopPosts returns the most engaged posts inside a window.
func (r *Repository) TopPosts(ctx context.Context, topicID int64, since, until time.Time, limit int32) ([]generated.SocialPost, error) {
	return r.queries.TopPostsByEngagement(ctx, topicID, since, until, limit)
}
