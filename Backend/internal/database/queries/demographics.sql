-- name: InsertDemographicSnapshot :one
INSERT INTO demographic_analytics (
    topic_id, language_distribution, region_distribution, interest_distribution, audience_groups
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, topic_id, snapshot_at, language_distribution, region_distribution, interest_distribution, audience_groups;

-- name: GetLatestDemographicSnapshot :one
SELECT id, topic_id, snapshot_at, language_distribution, region_distribution, interest_distribution, audience_groups
FROM demographic_analytics
WHERE topic_id = $1
ORDER BY snapshot_at DESC
LIMIT 1;
