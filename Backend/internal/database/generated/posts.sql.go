package generated

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// InsertPostParams carries the insert arguments for InsertPost.
type InsertPostParams struct {
	TopicID           int64
	Platform          string
	ExternalID        string
	AuthorReference   string
	AuthorDisplayName string
	Content           string
	Language          string
	RegionHint        string
	Timestamp         time.Time
	EngagementCount   int32
	ReplyToExternalID string
	MentionRefs       []string
}

const insertPost = `INSERT INTO social_posts (
	topic_id, platform, external_id, author_reference, author_display_name,
	content, language, region_hint, "timestamp", engagement_count,
	reply_to_external_id, mention_refs
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (platform, external_id) DO UPDATE
SET engagement_count = GREATEST(social_posts.engagement_count, EXCLUDED.engagement_count)
RETURNING id, topic_id, platform, external_id, author_reference, author_display_name,
	content, language, region_hint, "timestamp", engagement_count,
	reply_to_external_id, mention_refs, ingested_at`

// InsertPost stores a normalized post, making ingestion idempotent per
// (platform, external_id).
func (q *Queries) InsertPost(ctx context.Context, arg InsertPostParams) (SocialPost, error) {
	row := q.db.QueryRow(ctx, insertPost,
		arg.TopicID,
		arg.Platform,
		arg.ExternalID,
		arg.AuthorReference,
		arg.AuthorDisplayName,
		arg.Content,
		arg.Language,
		arg.RegionHint,
		arg.Timestamp,
		arg.EngagementCount,
		arg.ReplyToExternalID,
		arg.MentionRefs,
	)
	var i SocialPost
	err := row.Scan(
		&i.ID,
		&i.TopicID,
		&i.Platform,
		&i.ExternalID,
		&i.AuthorReference,
		&i.AuthorDisplayName,
		&i.Content,
		&i.Language,
		&i.RegionHint,
		&i.Timestamp,
		&i.EngagementCount,
		&i.ReplyToExternalID,
		&i.MentionRefs,
		&i.IngestedAt,
	)
	return i, err
}

const getPostByID = `SELECT id, topic_id, platform, external_id, author_reference, author_display_name,
	content, language, region_hint, "timestamp", engagement_count,
	reply_to_external_id, mention_refs, ingested_at
FROM social_posts
WHERE id = $1`

// GetPostByID loads a single post.
func (q *Queries) GetPostByID(ctx context.Context, id int64) (SocialPost, error) {
	row := q.db.QueryRow(ctx, getPostByID, id)
	var i SocialPost
	err := row.Scan(
		&i.ID,
		&i.TopicID,
		&i.Platform,
		&i.ExternalID,
		&i.AuthorReference,
		&i.AuthorDisplayName,
		&i.Content,
		&i.Language,
		&i.RegionHint,
		&i.Timestamp,
		&i.EngagementCount,
		&i.ReplyToExternalID,
		&i.MentionRefs,
		&i.IngestedAt,
	)
	return i, err
}

// ListPostsByTopicParams carries filter + pagination arguments for
// ListPostsByTopic. Since/Until are optional (pass pgtype.Timestamptz{} for
// unbounded).
type ListPostsByTopicParams struct {
	TopicID int64
	Since   pgtype.Timestamptz
	Until   pgtype.Timestamptz
	Limit   int32
	Offset  int32
}

const listPostsByTopic = `SELECT id, topic_id, platform, external_id, author_reference, author_display_name,
	content, language, region_hint, "timestamp", engagement_count,
	reply_to_external_id, mention_refs, ingested_at
FROM social_posts
WHERE topic_id = $1
  AND ($2::timestamptz IS NULL OR "timestamp" >= $2)
  AND ($3::timestamptz IS NULL OR "timestamp" <= $3)
ORDER BY "timestamp" DESC
LIMIT $4 OFFSET $5`

// ListPostsByTopic returns the chronological (newest first) post feed for a
// topic, optionally bounded to a time window.
func (q *Queries) ListPostsByTopic(ctx context.Context, arg ListPostsByTopicParams) ([]SocialPost, error) {
	rows, err := q.db.Query(ctx, listPostsByTopic, arg.TopicID, arg.Since, arg.Until, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SocialPost{}
	for rows.Next() {
		var i SocialPost
		if err := rows.Scan(
			&i.ID,
			&i.TopicID,
			&i.Platform,
			&i.ExternalID,
			&i.AuthorReference,
			&i.AuthorDisplayName,
			&i.Content,
			&i.Language,
			&i.RegionHint,
			&i.Timestamp,
			&i.EngagementCount,
			&i.ReplyToExternalID,
			&i.MentionRefs,
			&i.IngestedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const countPostsByTopic = `SELECT COUNT(*)::bigint FROM social_posts WHERE topic_id = $1`

// CountPostsByTopic returns the number of collected posts for a topic.
func (q *Queries) CountPostsByTopic(ctx context.Context, topicID int64) (int64, error) {
	row := q.db.QueryRow(ctx, countPostsByTopic, topicID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const countPostsByTopicInRange = `SELECT COUNT(*)::bigint FROM social_posts
WHERE topic_id = $1 AND "timestamp" >= $2 AND "timestamp" <= $3`

// CountPostsByTopicInRange counts posts inside a closed time window.
func (q *Queries) CountPostsByTopicInRange(ctx context.Context, topicID int64, since time.Time, until time.Time) (int64, error) {
	row := q.db.QueryRow(ctx, countPostsByTopicInRange, topicID, since, until)
	var count int64
	err := row.Scan(&count)
	return count, err
}

// PostsPerBucketParams carries the time-series arguments for PostsPerBucket.
type PostsPerBucketParams struct {
	TopicID       int64
	Since         time.Time
	Until         time.Time
	BucketMinutes int32
}

const postsPerBucket = `SELECT
	to_timestamp(floor((extract(epoch FROM "timestamp")::double precision) / ($4::double precision * 60)) * ($4::double precision * 60)) AS bucket,
	COUNT(*)::bigint AS post_count
FROM social_posts
WHERE topic_id = $1 AND "timestamp" >= $2 AND "timestamp" <= $3
GROUP BY bucket
ORDER BY bucket ASC`

// PostsPerBucket returns the discussion growth series: post counts bucketed
// into fixed-size time windows.
func (q *Queries) PostsPerBucket(ctx context.Context, arg PostsPerBucketParams) ([]PostsBucket, error) {
	rows, err := q.db.Query(ctx, postsPerBucket, arg.TopicID, arg.Since, arg.Until, arg.BucketMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PostsBucket{}
	for rows.Next() {
		var i PostsBucket
		if err := rows.Scan(&i.Bucket, &i.PostCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const topPostsByEngagement = `SELECT id, topic_id, platform, external_id, author_reference, author_display_name,
	content, language, region_hint, "timestamp", engagement_count,
	reply_to_external_id, mention_refs, ingested_at
FROM social_posts
WHERE topic_id = $1 AND "timestamp" >= $2 AND "timestamp" <= $3
ORDER BY engagement_count DESC
LIMIT $4`

// TopPostsByEngagement returns the most engaged posts inside a window; the
// timeline API surfaces these as important events.
func (q *Queries) TopPostsByEngagement(ctx context.Context, topicID int64, since time.Time, until time.Time, limit int32) ([]SocialPost, error) {
	rows, err := q.db.Query(ctx, topPostsByEngagement, topicID, since, until, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SocialPost{}
	for rows.Next() {
		var i SocialPost
		if err := rows.Scan(
			&i.ID,
			&i.TopicID,
			&i.Platform,
			&i.ExternalID,
			&i.AuthorReference,
			&i.AuthorDisplayName,
			&i.Content,
			&i.Language,
			&i.RegionHint,
			&i.Timestamp,
			&i.EngagementCount,
			&i.ReplyToExternalID,
			&i.MentionRefs,
			&i.IngestedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const recentPostsBrief = `SELECT id, content, language, "timestamp", engagement_count
FROM social_posts
WHERE topic_id = $1
ORDER BY "timestamp" DESC
LIMIT $2`

// RecentPostsBrief returns a compact newest-first projection used for
// keyword extraction and important-event detection.
func (q *Queries) RecentPostsBrief(ctx context.Context, topicID int64, limit int32) ([]RecentPostBrief, error) {
	rows, err := q.db.Query(ctx, recentPostsBrief, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []RecentPostBrief{}
	for rows.Next() {
		var i RecentPostBrief
		if err := rows.Scan(&i.ID, &i.Content, &i.Language, &i.Timestamp, &i.EngagementCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listPostsMissingSentiment = `SELECT p.id, p.topic_id, p.platform, p.external_id, p.author_reference, p.author_display_name,
	p.content, p.language, p.region_hint, p."timestamp", p.engagement_count,
	p.reply_to_external_id, p.mention_refs, p.ingested_at
FROM social_posts p
LEFT JOIN sentiment_results s ON s.post_id = p.id
WHERE p.topic_id = $1 AND s.id IS NULL
ORDER BY p."timestamp" DESC
LIMIT $2`

// ListPostsMissingSentiment returns posts that have not been scored yet —
// used by backfill jobs and the worker pipeline.
func (q *Queries) ListPostsMissingSentiment(ctx context.Context, topicID int64, limit int32) ([]SocialPost, error) {
	rows, err := q.db.Query(ctx, listPostsMissingSentiment, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SocialPost{}
	for rows.Next() {
		var i SocialPost
		if err := rows.Scan(
			&i.ID,
			&i.TopicID,
			&i.Platform,
			&i.ExternalID,
			&i.AuthorReference,
			&i.AuthorDisplayName,
			&i.Content,
			&i.Language,
			&i.RegionHint,
			&i.Timestamp,
			&i.EngagementCount,
			&i.ReplyToExternalID,
			&i.MentionRefs,
			&i.IngestedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listPostsMissingEmbedding = `SELECT p.id, p.topic_id, p.platform, p.external_id, p.author_reference, p.author_display_name,
	p.content, p.language, p.region_hint, p."timestamp", p.engagement_count,
	p.reply_to_external_id, p.mention_refs, p.ingested_at
FROM social_posts p
LEFT JOIN post_embeddings e ON e.post_id = p.id
WHERE p.topic_id = $1 AND e.post_id IS NULL
ORDER BY p."timestamp" DESC
LIMIT $2`

// ListPostsMissingEmbedding returns posts that have not been embedded yet.
func (q *Queries) ListPostsMissingEmbedding(ctx context.Context, topicID int64, limit int32) ([]SocialPost, error) {
	rows, err := q.db.Query(ctx, listPostsMissingEmbedding, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SocialPost{}
	for rows.Next() {
		var i SocialPost
		if err := rows.Scan(
			&i.ID,
			&i.TopicID,
			&i.Platform,
			&i.ExternalID,
			&i.AuthorReference,
			&i.AuthorDisplayName,
			&i.Content,
			&i.Language,
			&i.RegionHint,
			&i.Timestamp,
			&i.EngagementCount,
			&i.ReplyToExternalID,
			&i.MentionRefs,
			&i.IngestedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listPostSignals = `SELECT p.language, p.region_hint, p.content, p.author_reference,
	p.engagement_count,
	COALESCE(s.sentiment_score, 0)::double precision AS sentiment_score
FROM social_posts p
LEFT JOIN sentiment_results s ON s.post_id = p.id
WHERE p.topic_id = $1
ORDER BY p."timestamp" DESC
LIMIT $2`

// ListPostSignals returns the compact projection consumed by the
// demographics estimator and the graph builder.
func (q *Queries) ListPostSignals(ctx context.Context, topicID int64, limit int32) ([]PostSignal, error) {
	rows, err := q.db.Query(ctx, listPostSignals, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PostSignal{}
	for rows.Next() {
		var i PostSignal
		if err := rows.Scan(&i.Language, &i.RegionHint, &i.Content, &i.AuthorReference, &i.EngagementCount, &i.SentimentScore); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
