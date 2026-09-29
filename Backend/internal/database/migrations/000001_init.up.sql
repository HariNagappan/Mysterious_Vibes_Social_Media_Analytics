-- Core PulseGraph schema: identity, sources, monitored topics, collected
-- posts, sentiment results, derived analytics and the influence graph.

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'analyst'
                  CHECK (role IN ('admin', 'analyst', 'viewer')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS social_sources (
    id            BIGSERIAL PRIMARY KEY,
    platform      TEXT        NOT NULL,
    source_name   TEXT        NOT NULL,
    configuration JSONB       NOT NULL DEFAULT '{}'::jsonb,
    status        TEXT        NOT NULL DEFAULT 'inactive'
                  CHECK (status IN ('active', 'inactive', 'error')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (platform, source_name)
);

CREATE TABLE IF NOT EXISTS topics (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    keywords    TEXT[]      NOT NULL DEFAULT '{}',
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS topics_keywords_idx ON topics USING GIN (keywords);

CREATE TABLE IF NOT EXISTS social_posts (
    id                   BIGSERIAL PRIMARY KEY,
    topic_id             BIGINT      NOT NULL REFERENCES topics (id) ON DELETE CASCADE,
    platform             TEXT        NOT NULL,
    external_id          TEXT        NOT NULL,
    author_reference     TEXT        NOT NULL,
    author_display_name  TEXT        NOT NULL DEFAULT '',
    content              TEXT        NOT NULL,
    language             TEXT        NOT NULL DEFAULT 'und',
    region_hint          TEXT        NOT NULL DEFAULT '',
    "timestamp"          TIMESTAMPTZ NOT NULL,
    engagement_count     INTEGER     NOT NULL DEFAULT 0,
    reply_to_external_id TEXT        NOT NULL DEFAULT '',
    mention_refs         TEXT[]      NOT NULL DEFAULT '{}',
    ingested_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (platform, external_id)
);
CREATE INDEX IF NOT EXISTS social_posts_topic_ts_idx ON social_posts (topic_id, "timestamp" DESC);
CREATE INDEX IF NOT EXISTS social_posts_author_idx ON social_posts (author_reference);
CREATE INDEX IF NOT EXISTS social_posts_language_idx ON social_posts (language);
CREATE INDEX IF NOT EXISTS social_posts_engagement_idx ON social_posts (topic_id, engagement_count DESC);

CREATE TABLE IF NOT EXISTS sentiment_results (
    id              BIGSERIAL PRIMARY KEY,
    post_id         BIGINT      NOT NULL UNIQUE REFERENCES social_posts (id) ON DELETE CASCADE,
    emotion         TEXT        NOT NULL,
    sentiment_score DOUBLE PRECISION NOT NULL,
    confidence      DOUBLE PRECISION NOT NULL,
    model_version   TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS sentiment_results_emotion_idx ON sentiment_results (emotion);

CREATE TABLE IF NOT EXISTS demographic_analytics (
    id                    BIGSERIAL PRIMARY KEY,
    topic_id              BIGINT      NOT NULL REFERENCES topics (id) ON DELETE CASCADE,
    snapshot_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    language_distribution JSONB       NOT NULL DEFAULT '{}'::jsonb,
    region_distribution   JSONB       NOT NULL DEFAULT '{}'::jsonb,
    interest_distribution JSONB       NOT NULL DEFAULT '{}'::jsonb,
    audience_groups       JSONB       NOT NULL DEFAULT '[]'::jsonb
);
CREATE INDEX IF NOT EXISTS demographic_analytics_topic_idx
    ON demographic_analytics (topic_id, snapshot_at DESC);

CREATE TABLE IF NOT EXISTS trend_analytics (
    id          BIGSERIAL PRIMARY KEY,
    topic_id    BIGINT      NOT NULL REFERENCES topics (id) ON DELETE CASCADE,
    keyword     TEXT        NOT NULL DEFAULT '',
    trend_score DOUBLE PRECISION NOT NULL,
    growth_rate DOUBLE PRECISION NOT NULL,
    velocity    DOUBLE PRECISION NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS trend_analytics_topic_idx ON trend_analytics (topic_id, detected_at DESC);

CREATE TABLE IF NOT EXISTS graph_nodes (
    id                      BIGSERIAL PRIMARY KEY,
    external_user_reference TEXT        NOT NULL UNIQUE,
    display_name            TEXT        NOT NULL DEFAULT '',
    platform                TEXT        NOT NULL DEFAULT '',
    influence_score         DOUBLE PRECISION NOT NULL DEFAULT 0,
    community_id            INTEGER     NOT NULL DEFAULT 0,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS graph_edges (
    id               BIGSERIAL PRIMARY KEY,
    topic_id         BIGINT  NOT NULL REFERENCES topics (id) ON DELETE CASCADE,
    source_node      BIGINT  NOT NULL REFERENCES graph_nodes (id) ON DELETE CASCADE,
    target_node      BIGINT  NOT NULL REFERENCES graph_nodes (id) ON DELETE CASCADE,
    interaction_type TEXT    NOT NULL
                     CHECK (interaction_type IN ('reply', 'mention', 'repost', 'quote')),
    weight           INTEGER NOT NULL DEFAULT 1,
    UNIQUE (topic_id, source_node, target_node, interaction_type)
);
CREATE INDEX IF NOT EXISTS graph_edges_topic_idx ON graph_edges (topic_id);
CREATE INDEX IF NOT EXISTS graph_edges_source_idx ON graph_edges (source_node);
CREATE INDEX IF NOT EXISTS graph_edges_target_idx ON graph_edges (target_node);
