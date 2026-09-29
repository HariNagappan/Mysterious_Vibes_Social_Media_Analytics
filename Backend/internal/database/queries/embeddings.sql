-- name: UpsertEmbedding :one
INSERT INTO post_embeddings (post_id, model_version, embedding)
VALUES ($1, $2, $3::vector)
ON CONFLICT (post_id) DO UPDATE
SET model_version = EXCLUDED.model_version,
    embedding = EXCLUDED.embedding
RETURNING post_id, model_version, embedding::text AS embedding, created_at;

-- name: GetEmbeddingByPost :one
SELECT post_id, model_version, embedding::text AS embedding, created_at
FROM post_embeddings
WHERE post_id = $1;

-- name: SimilarPosts :many
SELECT pe.post_id,
    (pe.embedding <=> $2::vector)::double precision AS distance,
    p.content,
    p."timestamp" AS timestamp
FROM post_embeddings pe
JOIN social_posts p ON p.id = pe.post_id
WHERE p.topic_id = $1
ORDER BY pe.embedding <=> $2::vector ASC
LIMIT $3;
