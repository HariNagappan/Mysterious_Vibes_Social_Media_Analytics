package trends

import (
	"context"
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Repository persists and reads trend measurements.
type Repository struct {
	queries *generated.Queries
}

// NewRepository wraps the generated queries.
func NewRepository(queries *generated.Queries) *Repository {
	return &Repository{queries: queries}
}

// Buckets returns the post-count series for a window.
func (r *Repository) Buckets(ctx context.Context, topicID int64, since, until time.Time, bucketMinutes int32) ([]generated.PostsBucket, error) {
	return r.queries.PostsPerBucket(ctx, generated.PostsPerBucketParams{
		TopicID:       topicID,
		Since:         since,
		Until:         until,
		BucketMinutes: bucketMinutes,
	})
}

// RecentPosts returns compact post projections for keyword extraction.
func (r *Repository) RecentPosts(ctx context.Context, topicID int64, limit int32) ([]generated.RecentPostBrief, error) {
	return r.queries.RecentPostsBrief(ctx, topicID, limit)
}

// Insert stores one trend measurement.
func (r *Repository) Insert(ctx context.Context, arg generated.InsertTrendSnapshotParams) (generated.TrendAnalytic, error) {
	return r.queries.InsertTrendSnapshot(ctx, arg)
}

// Latest returns rows from the most recent detection batch.
func (r *Repository) Latest(ctx context.Context, topicID int64, limit int32) ([]generated.TrendAnalytic, error) {
	return r.queries.GetLatestTrends(ctx, topicID, limit)
}

// History returns the measurement history (newest first).
func (r *Repository) History(ctx context.Context, topicID int64, limit int32) ([]generated.TrendAnalytic, error) {
	return r.queries.TrendHistory(ctx, topicID, limit)
}
