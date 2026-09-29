-- name: TopicOverview :one
SELECT
    COUNT(*)::bigint AS total_posts,
    COALESCE(SUM(p.engagement_count), 0)::bigint AS total_engagement,
    COUNT(DISTINCT p.author_reference)::bigint AS distinct_authors,
    COALESCE(MIN(p."timestamp"), to_timestamp(0)) AS first_post_at,
    COALESCE(MAX(p."timestamp"), to_timestamp(0)) AS last_post_at,
    COALESCE(AVG(s.sentiment_score), 0)::double precision AS avg_sentiment
FROM social_posts p
LEFT JOIN sentiment_results s ON s.post_id = p.id
WHERE p.topic_id = $1;

-- name: PlatformBreakdown :many
SELECT platform, COUNT(*)::bigint AS count
FROM social_posts
WHERE topic_id = $1
GROUP BY platform
ORDER BY count DESC;

-- name: TopAuthors :many
SELECT author_reference,
    COUNT(*)::bigint AS post_count,
    COALESCE(SUM(engagement_count), 0)::bigint AS total_engagement
FROM social_posts
WHERE topic_id = $1
GROUP BY author_reference
ORDER BY total_engagement DESC
LIMIT $2;

-- name: PostsPerDay :many
SELECT date_trunc('day', "timestamp") AS day, COUNT(*)::bigint AS post_count
FROM social_posts
WHERE topic_id = $1 AND "timestamp" >= $2
GROUP BY day
ORDER BY day ASC;
