package generated

import (
	"context"
	"time"
)

// UpsertSentimentResultParams carries the insert arguments for
// UpsertSentimentResult.
type UpsertSentimentResultParams struct {
	PostID         int64
	Emotion        string
	SentimentScore float64
	Confidence     float64
	ModelVersion   string
}

const upsertSentimentResult = `INSERT INTO sentiment_results (post_id, emotion, sentiment_score, confidence, model_version)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (post_id) DO UPDATE
SET emotion = EXCLUDED.emotion,
    sentiment_score = EXCLUDED.sentiment_score,
    confidence = EXCLUDED.confidence,
    model_version = EXCLUDED.model_version
RETURNING id, post_id, emotion, sentiment_score, confidence, model_version, created_at`

// UpsertSentimentResult stores (or refreshes) the AI sentiment verdict for a
// post.
func (q *Queries) UpsertSentimentResult(ctx context.Context, arg UpsertSentimentResultParams) (SentimentResult, error) {
	row := q.db.QueryRow(ctx, upsertSentimentResult,
		arg.PostID,
		arg.Emotion,
		arg.SentimentScore,
		arg.Confidence,
		arg.ModelVersion,
	)
	var i SentimentResult
	err := row.Scan(&i.ID, &i.PostID, &i.Emotion, &i.SentimentScore, &i.Confidence, &i.ModelVersion, &i.CreatedAt)
	return i, err
}

const getSentimentByPost = `SELECT id, post_id, emotion, sentiment_score, confidence, model_version, created_at
FROM sentiment_results
WHERE post_id = $1`

// GetSentimentByPost loads the sentiment verdict for one post.
func (q *Queries) GetSentimentByPost(ctx context.Context, postID int64) (SentimentResult, error) {
	row := q.db.QueryRow(ctx, getSentimentByPost, postID)
	var i SentimentResult
	err := row.Scan(&i.ID, &i.PostID, &i.Emotion, &i.SentimentScore, &i.Confidence, &i.ModelVersion, &i.CreatedAt)
	return i, err
}

// SentimentDistributionParams carries the window arguments for
// SentimentDistribution.
type SentimentDistributionParams struct {
	TopicID int64
	Since   time.Time
	Until   time.Time
}

const sentimentDistribution = `SELECT s.emotion, COUNT(*)::bigint AS count
FROM sentiment_results s
JOIN social_posts p ON p.id = s.post_id
WHERE p.topic_id = $1 AND p."timestamp" >= $2 AND p."timestamp" <= $3
GROUP BY s.emotion
ORDER BY count DESC`

// SentimentDistribution returns the emotion histogram for a topic window.
func (q *Queries) SentimentDistribution(ctx context.Context, arg SentimentDistributionParams) ([]EmotionCount, error) {
	rows, err := q.db.Query(ctx, sentimentDistribution, arg.TopicID, arg.Since, arg.Until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmotionCount{}
	for rows.Next() {
		var i EmotionCount
		if err := rows.Scan(&i.Emotion, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// SentimentTimelineParams carries the window + bucket arguments for
// SentimentTimeline.
type SentimentTimelineParams struct {
	TopicID       int64
	Since         time.Time
	Until         time.Time
	BucketMinutes int32
}

const sentimentTimeline = `SELECT
	to_timestamp(floor((extract(epoch FROM p."timestamp")::double precision) / ($4::double precision * 60)) * ($4::double precision * 60)) AS bucket,
	AVG(s.sentiment_score)::double precision AS avg_score,
	COUNT(*)::bigint AS post_count,
	MIN(s.sentiment_score)::double precision AS min_score,
	MAX(s.sentiment_score)::double precision AS max_score
FROM sentiment_results s
JOIN social_posts p ON p.id = s.post_id
WHERE p.topic_id = $1 AND p."timestamp" >= $2 AND p."timestamp" <= $3
GROUP BY bucket
ORDER BY bucket ASC`

// SentimentTimeline returns the sentiment evolution series for a topic.
func (q *Queries) SentimentTimeline(ctx context.Context, arg SentimentTimelineParams) ([]SentimentBucket, error) {
	rows, err := q.db.Query(ctx, sentimentTimeline, arg.TopicID, arg.Since, arg.Until, arg.BucketMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SentimentBucket{}
	for rows.Next() {
		var i SentimentBucket
		if err := rows.Scan(&i.Bucket, &i.AvgScore, &i.PostCount, &i.MinScore, &i.MaxScore); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const countSentimentByTopic = `SELECT COUNT(*)::bigint
FROM sentiment_results s
JOIN social_posts p ON p.id = s.post_id
WHERE p.topic_id = $1`

// CountSentimentByTopic counts scored posts for a topic.
func (q *Queries) CountSentimentByTopic(ctx context.Context, topicID int64) (int64, error) {
	row := q.db.QueryRow(ctx, countSentimentByTopic, topicID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const avgSentimentByTopic = `SELECT COALESCE(AVG(s.sentiment_score), 0)::double precision
FROM sentiment_results s
JOIN social_posts p ON p.id = s.post_id
WHERE p.topic_id = $1 AND p."timestamp" >= $2 AND p."timestamp" <= $3`

// AvgSentimentByTopic returns the mean sentiment score for a window.
func (q *Queries) AvgSentimentByTopic(ctx context.Context, topicID int64, since time.Time, until time.Time) (float64, error) {
	row := q.db.QueryRow(ctx, avgSentimentByTopic, topicID, since, until)
	var avg float64
	err := row.Scan(&avg)
	return avg, err
}
