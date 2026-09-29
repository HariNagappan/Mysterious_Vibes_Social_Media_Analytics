-- pgvector-backed post embeddings, used for semantic similarity queries and
-- as durable storage for the embedding pipeline stage.

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS post_embeddings (
    post_id       BIGINT PRIMARY KEY REFERENCES social_posts (id) ON DELETE CASCADE,
    model_version TEXT        NOT NULL,
    embedding     vector(384) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS post_embeddings_ivfflat_idx
    ON post_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
