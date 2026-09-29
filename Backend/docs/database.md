# PulseGraph - Database

PostgreSQL 16 with the **pgvector** extension. Schema changes are managed by
versioned migrations embedded into the API binary.

## Entity map

```
users                social_sources            topics
  │                                             │
  │ (dashboard users)                           │ one topic ── many posts
  ▼                                             ▼
              social_posts ──────────► sentiment_results
                  │   │
                  │   └──► post_embeddings (vector(384))
                  │
                  └── author_reference ──► graph_nodes ──► graph_edges

topics ──► demographic_analytics   (aggregated only)
topics ──► trend_analytics         (time series of trend measurements)
```

## Tables

### users
Dashboard users. `role` is one of `admin`, `analyst`, `viewer`.

| column | type | notes |
| --- | --- | --- |
| id | bigserial PK | |
| name | text | |
| email | text | unique |
| password_hash | text | bcrypt |
| role | text | check-constrained |
| created_at / updated_at | timestamptz | |

### social_sources
Connected platforms. `configuration` is JSONB (credentials never stored in
plaintext in production; demo mode is key-less).

### topics
A monitored conversation with its keyword set (`text[]`, GIN-indexed).

### social_posts
Collected content. `external_id` + `platform` are unique together, making
ingestion idempotent. `mention_refs` (text[]) and `reply_to_external_id` feed
the graph builder. `"timestamp"` is the post time; `ingested_at` is the
collection time.

Indexes: `(topic_id, timestamp DESC)`, `(topic_id, engagement_count DESC)`,
`author_reference`, `language`.

### sentiment_results
One verdict per post (`UNIQUE (post_id)`), upserted by the worker.
`model_version` records which model produced the score
(`sentiment-service` or the `lexicon-v1` fallback).

### demographic_analytics
**Aggregated only** — never per-user rows. Snapshots carry JSONB
distributions for languages, regions and interests plus behavioural audience
groups (broadcasters / amplifiers / participants / advocates / critics).

### trend_analytics
Time series of trend measurements per topic: an `overall` sentinel row plus
one row per tracked keyword, each with `trend_score`, `growth_rate` and
`velocity`. The API reads the freshest batch (10-minute window).

### graph_nodes / graph_edges
The influence graph. Nodes are external accounts (`external_user_reference`
unique) with PageRank `influence_score` and a `community_id` from label
propagation. Edges are per-topic, typed (`reply`, `mention`, `repost`,
`quote`) and weighted by interaction count (upserts accumulate weight).

### post_embeddings
pgvector storage (`vector(384)`) with an IVFFlat cosine index for semantic
similarity search (`internal/database/queries/embeddings.sql`).

## Migrations

- Versioned as `000001_name.up.sql` / `.down.sql` (golang-migrate compatible
  naming) under `internal/database/migrations/`.
- Applied automatically by the API at boot when `RUN_MIGRATIONS=true`
  (compose default), or explicitly:

  ```bash
  go run ./cmd/api -migrate-only
  # or
  ./scripts/run_migrations.sh
  ```

- Applied versions are tracked in `schema_migrations`; re-running is a no-op.
- To use the golang-migrate CLI later, point it at the same folder — naming
  is already compatible.

## Regenerating the query layer

Queries live in `internal/database/queries/*.sql` (sqlc annotations). The
compiled counterparts in `internal/database/generated/` are currently
hand-maintained mirrors because the `sqlc` CLI is not available in every
build environment:

```bash
sqlc generate        # refreshes internal/database/generated from queries + migrations
```

## Natural next steps (production)

- Row-level retention policies (posts older than N days → cold storage).
- Table partitioning for `social_posts` / `sentiment_results` by month.
- `demographic_analytics` / `trend_analytics` rollups for long-range charts.
- Read replicas for analytics queries; PgBouncer for connection fan-in.
