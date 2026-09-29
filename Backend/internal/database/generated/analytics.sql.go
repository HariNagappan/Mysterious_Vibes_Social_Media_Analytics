package generated

import (
	"context"
	"time"
)

const topicOverview = `SELECT
	COUNT(*)::bigint AS total_posts,
	COALESCE(SUM(p.engagement_count), 0)::bigint AS total_engagement,
	COUNT(DISTINCT p.author_reference)::bigint AS distinct_authors,
	COALESCE(MIN(p."timestamp"), to_timestamp(0)) AS first_post_at,
	COALESCE(MAX(p."timestamp"), to_timestamp(0)) AS last_post_at,
	COALESCE(AVG(s.sentiment_score), 0)::double precision AS avg_sentiment
FROM social_posts p
LEFT JOIN sentiment_results s ON s.post_id = p.id
WHERE p.topic_id = $1`

// TopicOverview computes headline statistics for a topic.
func (q *Queries) TopicOverview(ctx context.Context, topicID int64) (TopicOverview, error) {
	row := q.db.QueryRow(ctx, topicOverview, topicID)
	var i TopicOverview
	err := row.Scan(&i.TotalPosts, &i.TotalEngagement, &i.DistinctAuthors, &i.FirstPostAt, &i.LastPostAt, &i.AvgSentiment)
	return i, err
}

const platformBreakdown = `SELECT platform, COUNT(*)::bigint AS count
FROM social_posts
WHERE topic_id = $1
GROUP BY platform
ORDER BY count DESC`

// PlatformBreakdown returns post volume per connected platform.
func (q *Queries) PlatformBreakdown(ctx context.Context, topicID int64) ([]PlatformCount, error) {
	rows, err := q.db.Query(ctx, platformBreakdown, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PlatformCount{}
	for rows.Next() {
		var i PlatformCount
		if err := rows.Scan(&i.Platform, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const topAuthors = `SELECT author_reference,
	COUNT(*)::bigint AS post_count,
	COALESCE(SUM(engagement_count), 0)::bigint AS total_engagement
FROM social_posts
WHERE topic_id = $1
GROUP BY author_reference
ORDER BY total_engagement DESC
LIMIT $2`

// TopAuthors ranks the most engaging authors participating in a topic.
func (q *Queries) TopAuthors(ctx context.Context, topicID int64, limit int32) ([]TopAuthor, error) {
	rows, err := q.db.Query(ctx, topAuthors, topicID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TopAuthor{}
	for rows.Next() {
		var i TopAuthor
		if err := rows.Scan(&i.AuthorReference, &i.PostCount, &i.TotalEngagement); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const postsPerDay = `SELECT date_trunc('day', "timestamp") AS day, COUNT(*)::bigint AS post_count
FROM social_posts
WHERE topic_id = $1 AND "timestamp" >= $2
GROUP BY day
ORDER BY day ASC`

// PostsPerDay returns daily post volume since the given time.
func (q *Queries) PostsPerDay(ctx context.Context, topicID int64, since time.Time) ([]DayCount, error) {
	rows, err := q.db.Query(ctx, postsPerDay, topicID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []DayCount{}
	for rows.Next() {
		var i DayCount
		if err := rows.Scan(&i.Day, &i.PostCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
