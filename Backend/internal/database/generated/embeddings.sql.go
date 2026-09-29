package generated

import (
	"context"
)

// UpsertEmbeddingParams carries the insert arguments for UpsertEmbedding.
// Embedding carries the canonical pgvector literal, e.g. "[0.01,0.02,...]".
type UpsertEmbeddingParams struct {
	PostID       int64
	ModelVersion string
	Embedding    string
}

const upsertEmbedding = `INSERT INTO post_embeddings (post_id, model_version, embedding)
VALUES ($1, $2, $3::vector)
ON CONFLICT (post_id) DO UPDATE
SET model_version = EXCLUDED.model_version,
    embedding = EXCLUDED.embedding
RETURNING post_id, model_version, embedding::text AS embedding, created_at`

// UpsertEmbedding stores the embedding vector for a post.
func (q *Queries) UpsertEmbedding(ctx context.Context, arg UpsertEmbeddingParams) (PostEmbedding, error) {
	row := q.db.QueryRow(ctx, upsertEmbedding, arg.PostID, arg.ModelVersion, arg.Embedding)
	var i PostEmbedding
	err := row.Scan(&i.PostID, &i.ModelVersion, &i.Embedding, &i.CreatedAt)
	return i, err
}

const getEmbeddingByPost = `SELECT post_id, model_version, embedding::text AS embedding, created_at
FROM post_embeddings
WHERE post_id = $1`

// GetEmbeddingByPost loads the stored embedding for one post.
func (q *Queries) GetEmbeddingByPost(ctx context.Context, postID int64) (PostEmbedding, error) {
	row := q.db.QueryRow(ctx, getEmbeddingByPost, postID)
	var i PostEmbedding
	err := row.Scan(&i.PostID, &i.ModelVersion, &i.Embedding, &i.CreatedAt)
	return i, err
}

// SimilarPostsParams carries the arguments for SimilarPosts.
type SimilarPostsParams struct {
	TopicID   int64
	Embedding string
	Limit     int32
}

const similarPosts = `SELECT pe.post_id,
	(pe.embedding <=> $2::vector)::double precision AS distance,
	p.content,
	p."timestamp" AS timestamp
FROM post_embeddings pe
JOIN social_posts p ON p.id = pe.post_id
WHERE p.topic_id = $1
ORDER BY pe.embedding <=> $2::vector ASC
LIMIT $3`

// SimilarPosts returns the posts semantically closest to the given vector
// (cosine distance) within a topic — the core of narrative similarity.
func (q *Queries) SimilarPosts(ctx context.Context, arg SimilarPostsParams) ([]SimilarPost, error) {
	rows, err := q.db.Query(ctx, similarPosts, arg.TopicID, arg.Embedding, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SimilarPost{}
	for rows.Next() {
		var i SimilarPost
		if err := rows.Scan(&i.PostID, &i.Distance, &i.Content, &i.Timestamp); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
