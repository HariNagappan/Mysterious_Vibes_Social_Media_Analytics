# PulseGraph - Architecture

PulseGraph is an AI-powered social media analytics platform built for
**Smart India Hackathon 2026, Problem Statement 26152**. It ingests public
conversations from X (Twitter) and Telegram and extracts four kinds of
intelligence:

1. **Sentiment and emotion** — what people feel, how it evolves
2. **Audience characteristics** — language, region and interest patterns
3. **Emerging trends** — which narratives are growing, and how fast
4. **Influence and propagation** — who shapes the conversation and how it moves

## High-level view

```
                         ┌────────────────────────────────────────────────────┐
   X (Twitter) ──┐       │                     Go backend                     │
   Telegram  ────┼──►    │                                                    │
   (future:      │       │  cmd/api        REST API (Gin) + JWT + Swagger     │
    Instagram,   │       │  cmd/worker     stream consumers (AI stages)       │
    Reddit, ...) │       │  cmd/scheduler  analytics cycles + demo ingestion  │
                 │       │                                                    │
   ingestion/connectors  │   Handler ─► Service ─► Repository ─► generated    │
   (plugin registry)     │        (clean architecture layering)               │
                         └───────────────┬────────────────────────────────────┘
                                         │
        ┌────────────────────────────────┼──────────────────────────────────┐
        │                                │                                  │
        ▼                                ▼                                  ▼
  ┌───────────┐                   ┌────────────┐                    ┌──────────────┐
  │  Redis    │                   │ PostgreSQL │                    │ Python ML    │
  │  cache +  │                   │ + pgvector │                    │ services     │
  │  Streams  │                   │            │                    │ (FastAPI)    │
  └───────────┘                   └────────────┘                    └──────────────┘
                                         ▲                                  │
        Elasticsearch/OpenSearch ────────┘ (abstraction, optional)          │
        abstraction (no-op fallback)                                       │
                                                                           ▼
                                                            sentiment /embedding /
                                                            summarizer services
```

## Processing pipeline

The async pipeline is an event chain over **Redis Streams**. Each stage
publishes the next stage's event; failures are retried and dead-lettered.

```
connector ─► normalizer ─► queue (stream:posts)
                               │
                               ▼
                     language detection (heuristic, in-Go)
                               │
                               ▼
   worker stage 1: sentiment scoring ──► search indexing
                               │
                               ▼ (stream:sentiment)
   worker stage 2: text embedding ──► pgvector storage
                               │
                               ▼ (stream:analytics)
   scheduler: trend calculation ─► graph construction ─► analytics
   aggregation ─► AI report generation (dashboard, on demand)
```

Event types (`internal/pipeline/events.go`):

| Event | Producer | Consumer |
| --- | --- | --- |
| `post.ingested` | scheduler (demo ingestion) / future live connectors | worker stage 1 |
| `sentiment.completed` | worker stage 1 | worker stage 2 |
| `embedding.completed` | worker stage 2 | analytics trail |
| `trend.updated`, `graph.updated` | scheduler | audit trail |

### The Kafka seam

`internal/pipeline` defines `Producer` and `Consumer` interfaces with a Redis
Streams implementation. A Kafka producer/consumer can be swapped in without
touching a single caller — the transport is an implementation detail of the
pipeline package.

## Clean architecture layering

```
Handler (internal/api/handlers, internal/auth/handler.go)
   │  HTTP binding, validation, error envelopes — no business logic
Service (internal/timeline, internal/sentiment, internal/trends, ...)
   │  use-cases, scoring, aggregation, caching decisions
Repository (internal/*/repository.go)
   │  generated query layer calls, row → domain mapping
Database (internal/database: migrations + sqlc-style generated package)
```

Cross-cutting packages: `config` (env-first configuration), `logger`
(structured JSON via zap), `cache` (Redis wrapper), `middleware`
(auth/logging/recovery/metrics/CORS/request-id).

## Why these choices

- **Redis Streams for the prototype** — zero extra infrastructure, consumer
  groups, ack + pending semantics. The interfaces keep Kafka a drop-in later.
- **SQLC-style generated layer** — typed SQL in, typed Go out. The package in
  `internal/database/generated` is a hand-maintained mirror of `sqlc generate`
  output (the CLI may not exist in every build environment; run
  `sqlc generate` to refresh it).
- **Self-contained Python services** — the demo must run anywhere, offline.
  The sentiment lexicon, hashing embeddings and template summarizer all work
  with zero downloads; heavier models are a configuration change.
- **Grounding rule for AI reports** — the summarizer only explains computed
  metrics, and the Go client additionally validates every number in a report
  against the input payload (`internal/ai/report.go`). A report that invents
  a number is rejected and replaced by the deterministic template.
- **Graceful degradation** — every AI call has an in-process fallback
  (sentiment → lexicon analyzer), or retry + dead-letter (embeddings), so a
  single failed dependency never takes the platform down.
- **Aggregate-only demographics** — per-user records are never stored; the
  estimator works on aggregated distributions.

## Components

| Component | Binary | Responsibility |
| --- | --- | --- |
| API | `cmd/api` | REST, auth, dashboard assembly, metrics, OpenAPI, migrations |
| Worker | `cmd/worker` | stream consumers: sentiment, search indexing, embeddings |
| Scheduler | `cmd/scheduler` | periodic trends/demographics/graph recompute; demo ingestion |

## Connector plugin model

Supporting a new platform means implementing one interface and registering it:

```go
type Connector interface {
    Name() string
    Platform() string
    Fetch(ctx context.Context, opts FetchOptions) ([]RawPost, error)
}
```

The scheduler resolves connectors through the registry, so Instagram,
Facebook, Reddit or YouTube connectors are additive changes only. Live
connectors call the real APIs when credentials are configured and fall back
to deterministic demo generators otherwise, which keeps the pipeline fully
demonstrable without keys.

## Deployment topology

- **docker-compose** (`deployments/docker-compose.yml`) — postgres (pgvector),
  redis, api, worker, scheduler, three ML services; optional profiles for
  Elasticsearch (`--profile search`) and Prometheus (`--profile monitoring`).
- **Kubernetes** (`deployments/kubernetes/`) — namespace, configmap, secret
  example, postgres StatefulSet, redis, api/worker/scheduler and ML services.

## Observability

- Structured JSON logs on every component (zap, shared `logger` package).
- Prometheus metrics: HTTP (api) plus business/pipeline metrics
  (`pulsegraph_posts_ingested_total`, `pulsegraph_sentiment_processed_total`,
  `pulsegraph_pipeline_failures_total`, ...) exposed by api (:8080/metrics),
  worker (:9091/metrics) and scheduler (:9092/metrics).
- Correlation via `X-Request-ID` echoed through every access log line.
