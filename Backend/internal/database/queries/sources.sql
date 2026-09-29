-- name: CreateSource :one
INSERT INTO social_sources (platform, source_name, configuration, status)
VALUES ($1, $2, $3, $4)
RETURNING id, platform, source_name, configuration, status, created_at;

-- name: ListSources :many
SELECT id, platform, source_name, configuration, status, created_at
FROM social_sources
ORDER BY created_at DESC;

-- name: UpdateSourceStatus :one
UPDATE social_sources
SET status = $2
WHERE id = $1
RETURNING id, platform, source_name, configuration, status, created_at;
