package sentiment

import (
	"context"
	"time"

	"github.com/pulsegraph/pulsegraph-backend/internal/database/generated"
)

// Repository persists and reads sentiment data via the generated query layer.
type Repository struct {
	queries *generated.Queries
}

// NewRepository wraps the generated queries.
func NewRepository(queries *generated.Queries) *Repository {
	return &Repository{queries: queries}
}

// GetPost loads a post for scoring.
func (r *Repository) GetPost(ctx context.Context, postID int64) (generated.SocialPost, error) {
	return r.queries.GetPostByID(ctx, postID)
}

// UpsertResult stores the verdict for a post.
func (r *Repository) UpsertResult(ctx context.Context, arg generated.UpsertSentimentResultParams) (generated.SentimentResult, error) {
	return r.queries.UpsertSentimentResult(ctx, arg)
}

// Distribution returns the emotion histogram for a window.
func (r *Repository) Distribution(ctx context.Context, topicID int64, since, until time.Time) ([]generated.EmotionCount, error) {
	return r.queries.SentimentDistribution(ctx, generated.SentimentDistributionParams{TopicID: topicID, Since: since, Until: until})
}

// Timeline returns the sentiment evolution series.
func (r *Repository) Timeline(ctx context.Context, topicID int64, since, until time.Time, bucketMinutes int32) ([]generated.SentimentBucket, error) {
	return r.queries.SentimentTimeline(ctx, generated.SentimentTimelineParams{
		TopicID:       topicID,
		Since:         since,
		Until:         until,
		BucketMinutes: bucketMinutes,
	})
}

// CountByTopic counts scored posts.
func (r *Repository) CountByTopic(ctx context.Context, topicID int64) (int64, error) {
	return r.queries.CountSentimentByTopic(ctx, topicID)
}

// CountPostsByTopic counts all collected posts (scored or not).
func (r *Repository) CountPostsByTopic(ctx context.Context, topicID int64) (int64, error) {
	return r.queries.CountPostsByTopic(ctx, topicID)
}

// Average returns the mean sentiment score for a window.
func (r *Repository) Average(ctx context.Context, topicID int64, since, until time.Time) (float64, error) {
	return r.queries.AvgSentimentByTopic(ctx, topicID, since, until)
}
