-- name: InsertTrendSnapshot :one
INSERT INTO trend_analytics (topic_id, keyword, trend_score, growth_rate, velocity)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, topic_id, keyword, trend_score, growth_rate, velocity, detected_at;

-- name: GetLatestTrends :many
SELECT id, topic_id, keyword, trend_score, growth_rate, velocity, detected_at
FROM trend_analytics
WHERE topic_id = $1
  AND detected_at >= (SELECT COALESCE(MAX(detected_at), to_timestamp(0)) FROM trend_analytics WHERE topic_id = $1) - interval '10 minutes'
ORDER BY trend_score DESC
LIMIT $2;

-- name: TrendHistory :many
SELECT id, topic_id, keyword, trend_score, growth_rate, velocity, detected_at
FROM trend_analytics
WHERE topic_id = $1
ORDER BY detected_at DESC
LIMIT $2;
