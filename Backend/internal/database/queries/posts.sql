-- name: InsertPost :one
INSERT INTO social_posts (
    topic_id, platform, external_id, author_reference, author_display_name,
    content, language, region_hint, "timestamp", engagement_count,
    reply_to_external_id, mention_refs
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (platform, external_id) DO UPDATE
SET engagement_count = GREATEST(social_posts.engagement_count, EXCLUDED.engagement_count)
RETURNING id, topic_id, platform, external_id, author_reference, author_display_name,
    content, language, region_hint, "timestamp", engagement_count,
    reply_to_external_id, mention_refs, ingested_at;

-- name: GetPostByID :one
SELECT id, topic_id, platform, external_id, author_reference, author_display_name,
    content, language, region_hint, "timestamp", engagement_count,
    reply_to_external_id, mention_refs, ingested_at
FROM social_posts
WHERE id = $1;

-- name: ListPostsByTopic :many
SELECT id, topic_id, platform, external_id, author_reference, author_display_name,
    content, language, region_hint, "timestamp", engagement_count,
    reply_to_external_id, mention_refs, ingested_at
FROM social_posts
WHERE topic_id = $1
  AND ($2::timestamptz IS NULL OR "timestamp" >= $2)
  AND ($3::timestamptz IS NULL OR "timestamp" <= $3)
ORDER BY "timestamp" DESC
LIMIT $4 OFFSET $5;

-- name: CountPostsByTopic :one
SELECT COUNT(*)::bigint FROM social_posts WHERE topic_id = $1;

-- name: CountPostsByTopicInRange :one
SELECT COUNT(*)::bigint FROM social_posts
WHERE topic_id = $1 AND "timestamp" >= $2 AND "timestamp" <= $3;

-- name: PostsPerBucket :many
SELECT
    to_timestamp(floor((extract(epoch FROM "timestamp")::double precision) / ($4::double precision * 60)) * ($4::double precision * 60)) AS bucket,
    COUNT(*)::bigint AS post_count
FROM social_posts
WHERE topic_id = $1 AND "timestamp" >= $2 AND "timestamp" <= $3
GROUP BY bucket
ORDER BY bucket ASC;

-- name: TopPostsByEngagement :many
SELECT id, topic_id, platform, external_id, author_reference, author_display_name,
    content, language, region_hint, "timestamp", engagement_count,
    reply_to_external_id, mention_refs, ingested_at
FROM social_posts
WHERE topic_id = $1 AND "timestamp" >= $2 AND "timestamp" <= $3
ORDER BY engagement_count DESC
LIMIT $4;

-- name: RecentPostsBrief :many
SELECT id, content, language, "timestamp", engagement_count
FROM social_posts
WHERE topic_id = $1
ORDER BY "timestamp" DESC
LIMIT $2;

-- name: ListPostsMissingSentiment :many
SELECT p.id, p.topic_id, p.platform, p.external_id, p.author_reference, p.author_display_name,
    p.content, p.language, p.region_hint, p."timestamp", p.engagement_count,
    p.reply_to_external_id, p.mention_refs, p.ingested_at
FROM social_posts p
LEFT JOIN sentiment_results s ON s.post_id = p.id
WHERE p.topic_id = $1 AND s.id IS NULL
ORDER BY p."timestamp" DESC
LIMIT $2;

-- name: ListPostsMissingEmbedding :many
SELECT p.id, p.topic_id, p.platform, p.external_id, p.author_reference, p.author_display_name,
    p.content, p.language, p.region_hint, p."timestamp", p.engagement_count,
    p.reply_to_external_id, p.mention_refs, p.ingested_at
FROM social_posts p
LEFT JOIN post_embeddings e ON e.post_id = p.id
WHERE p.topic_id = $1 AND e.post_id IS NULL
ORDER BY p."timestamp" DESC
LIMIT $2;

-- name: ListPostSignals :many
SELECT p.language, p.region_hint, p.content, p.author_reference,
    p.engagement_count,
    COALESCE(s.sentiment_score, 0)::double precision AS sentiment_score
FROM social_posts p
LEFT JOIN sentiment_results s ON s.post_id = p.id
WHERE p.topic_id = $1
ORDER BY p."timestamp" DESC
LIMIT $2;
