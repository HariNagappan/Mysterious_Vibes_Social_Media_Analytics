-- name: UpsertSentimentResult :one
INSERT INTO sentiment_results (post_id, emotion, sentiment_score, confidence, model_version)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (post_id) DO UPDATE
SET emotion = EXCLUDED.emotion,
    sentiment_score = EXCLUDED.sentiment_score,
    confidence = EXCLUDED.confidence,
    model_version = EXCLUDED.model_version
RETURNING id, post_id, emotion, sentiment_score, confidence, model_version, created_at;

-- name: GetSentimentByPost :one
SELECT id, post_id, emotion, sentiment_score, confidence, model_version, created_at
FROM sentiment_results
WHERE post_id = $1;

-- name: SentimentDistribution :many
SELECT s.emotion, COUNT(*)::bigint AS count
FROM sentiment_results s
JOIN social_posts p ON p.id = s.post_id
WHERE p.topic_id = $1 AND p."timestamp" >= $2 AND p."timestamp" <= $3
GROUP BY s.emotion
ORDER BY count DESC;

-- name: SentimentTimeline :many
SELECT
    to_timestamp(floor((extract(epoch FROM p."timestamp")::double precision) / ($4::double precision * 60)) * ($4::double precision * 60)) AS bucket,
    AVG(s.sentiment_score)::double precision AS avg_score,
    COUNT(*)::bigint AS post_count,
    MIN(s.sentiment_score)::double precision AS min_score,
    MAX(s.sentiment_score)::double precision AS max_score
FROM sentiment_results s
JOIN social_posts p ON p.id = s.post_id
WHERE p.topic_id = $1 AND p."timestamp" >= $2 AND p."timestamp" <= $3
GROUP BY bucket
ORDER BY bucket ASC;

-- name: CountSentimentByTopic :one
SELECT COUNT(*)::bigint
FROM sentiment_results s
JOIN social_posts p ON p.id = s.post_id
WHERE p.topic_id = $1;

-- name: AvgSentimentByTopic :one
SELECT COALESCE(AVG(s.sentiment_score), 0)::double precision
FROM sentiment_results s
JOIN social_posts p ON p.id = s.post_id
WHERE p.topic_id = $1 AND p."timestamp" >= $2 AND p."timestamp" <= $3;
