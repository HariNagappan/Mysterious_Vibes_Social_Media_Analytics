# PulseGraph - API

Base URL: `http://localhost:8080/api/v1` — interactive spec at `/swagger`,
machine-readable at `/swagger/openapi.yaml`.

## Conventions

- **Auth**: `Authorization: Bearer <JWT>` on every endpoint except
  `/auth/*` and the operational routes.
- **Errors** — one envelope everywhere:

  ```json
  { "error": { "code": "topic_not_found", "message": "topic was not found" } }
  ```

- **Pagination**: `?limit=&offset=` (bounded server-side).
- **Time windows**: RFC 3339 `?from=&to=` (defaults to the last 7 days where
  a window is accepted).
- Correlation: every response carries `X-Request-ID` (echoed from the
  request when provided).

## Authentication

### POST /auth/register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Demo Analyst","email":"demo@pulsegraph.dev","password":"PulseGraph@2026","role":"analyst"}'
```

```json
{
  "token": "<jwt>",
  "expires_at": "2026-10-01T12:00:00Z",
  "user": { "id": 4, "name": "Demo Analyst", "email": "demo@pulsegraph.dev", "role": "analyst" }
}
```

### POST /auth/login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"analyst@pulsegraph.dev","password":"PulseGraph@2026"}'
```

Errors: `401 invalid_credentials`.

## Topics

### POST /topics

```json
{
  "name": "Flood misinformation",
  "keywords": ["flood", "relief", "evacuation"],
  "description": "Monitors rumour spread during flooding."
}
```

### GET /topics

`{ "topics": [ ... ], "total": 3 }`

## Intelligence endpoints

All intelligence endpoints are topic-scoped and return 404 with
`topic_not_found` for unknown ids.

### GET /topics/:id/timeline

Chronological posts, discussion growth buckets and important events
(top-engagement moments).

```json
{
  "topic_id": 1,
  "window": { "from": "2026-09-22T00:00:00Z", "to": "2026-09-29T00:00:00Z" },
  "posts": [ { "id": 1001, "platform": "twitter", "content": "...", "engagement_count": 42 } ],
  "growth": [ { "bucket": "2026-09-22T06:00:00Z", "post_count": 25 } ],
  "important_events": [ { "post_id": 1024, "author": "x:user:floodwatch_assam", "engagement_count": 3120 } ]
}
```

Query params: `from`, `to` (RFC 3339), `limit` (<=500), `offset`.

### GET /topics/:id/sentiment

```json
{
  "topic_id": 1,
  "overall_score": -0.31,
  "analyzed_posts": 380,
  "total_posts": 380,
  "distribution": [ { "emotion": "fear", "count": 151, "share": 0.397 } ],
  "timeline": [ { "bucket": "2026-09-22T00:00:00Z", "avg_score": -0.2, "post_count": 18 } ]
}
```

### GET /topics/:id/demographics

```json
{
  "topic_id": 1,
  "language_distribution": { "en": 0.45, "hi": 0.35, "bn": 0.2 },
  "region_distribution": { "Assam": 0.55, "West Bengal": 0.2 },
  "interest_distribution": { "Disaster response": 0.62 },
  "audience_groups": [
    { "label": "Broadcasters", "share": 0.08, "description": "High-reach accounts ..." }
  ]
}
```

### GET /topics/:id/trends

```json
{
  "topic_id": 1,
  "overall": { "keyword": "overall", "trend_score": 61.2, "growth_rate": 0.44, "velocity": 1.8 },
  "rising_keywords": [
    { "keyword": "relief", "trend_score": 58.9, "growth_rate": 0.62, "velocity": 0.9 }
  ]
}
```

### GET /topics/:id/network

```json
{
  "topic_id": 1,
  "nodes": [ { "id": 12, "external_user_reference": "x:user:floodwatch_assam", "influence_score": 0.88, "community_id": 0 } ],
  "edges": [ { "source_node": 12, "target_node": 44, "interaction_type": "mention", "weight": 7 } ],
  "top_influencers": [ { "reference": "x:user:floodwatch_assam", "influence_score": 0.88 } ],
  "community_count": 3
}
```

### GET /topics/:id/dashboard

The complete intelligence payload used by the frontend:

```json
{
  "topic": { "id": 1, "name": "Flood misinformation - Assam" },
  "generated_at": "2026-09-29T18:00:00Z",
  "statistics": { "overview": { "total_posts": 380, "total_engagement": 51230 } },
  "timeline": { "...": "..." },
  "sentiment": { "...": "..." },
  "demographics": { "...": "..." },
  "trends": { "...": "..." },
  "network": { "...": "..." },
  "summary": {
    "title": "PulseGraph intelligence report - Flood misinformation - Assam",
    "executive_summary": "PulseGraph analysed ...",
    "key_findings": [ "Analysed 380 posts from ..." ],
    "mode": "ml-service"
  }
}
```

Responses are cached in Redis (60s default) and invalidated by the scheduler
after each analytics cycle.

## Search

### GET /search/posts?q=flood

Requires `SEARCH_ENABLED=true` with a reachable Elasticsearch/OpenSearch.
When disabled the endpoint answers `{ "search_enabled": false, "hits": [] }`
so the frontend can hide the feature.

## Operational endpoints

| Route | Purpose |
| --- | --- |
| `GET /healthz` | liveness |
| `GET /readyz` | readiness (checks PostgreSQL + Redis) |
| `GET /metrics` | Prometheus |
| `GET /swagger` | human index page |
| `GET /swagger/openapi.yaml` | OpenAPI 3.0 specification |

## Error codes

| code | HTTP | meaning |
| --- | --- | --- |
| `invalid_request` | 400 | payload/param validation failed |
| `invalid_id` | 400 | path id is not a positive integer |
| `invalid_window` | 400 | `from` >= `to` |
| `missing_token` / `invalid_token` | 401 | absent or bad JWT |
| `invalid_credentials` | 401 | wrong email/password |
| `forbidden` | 403 | role not permitted |
| `topic_not_found` | 404 | unknown topic id |
| `email_taken` | 409 | duplicate registration |
| `internal_error` | 500 | unexpected server error (logged with request id) |
