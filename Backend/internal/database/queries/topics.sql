-- name: CreateTopic :one
INSERT INTO topics (name, keywords, description)
VALUES ($1, $2, $3)
RETURNING id, name, keywords, description, created_at, updated_at;

-- name: GetTopicByID :one
SELECT id, name, keywords, description, created_at, updated_at
FROM topics
WHERE id = $1;

-- name: ListTopics :many
SELECT id, name, keywords, description, created_at, updated_at
FROM topics
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountTopics :one
SELECT COUNT(*)::bigint FROM topics;
