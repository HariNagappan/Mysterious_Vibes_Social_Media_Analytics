# PulseGraph - AI Social Media Analytics Backend

Backend of an AI-powered social media analytics platform built for
**Smart India Hackathon 2026 - Problem Statement 26152 (Social Media Analytics)**.

PulseGraph ingests public conversations from **X (Twitter)** and **Telegram**
and answers the questions an intelligence analyst actually asks:

| Question | Delivered by |
| --- | --- |
| What are people discussing? | timeline + keyword extraction |
| How are people reacting emotionally? | sentiment engine + emotion timeline |
| Who is participating? | aggregated audience analytics |
| Which narratives are growing? | trend detection with growth + velocity scoring |
| Who is influencing the spread? | influence graph (PageRank + communities) |
| How is information moving? | interaction edges (reply/mention) per topic |

Every endpoint is wired to a real processing pipeline: connectors -> normalizer
-> Redis Streams -> language detection -> sentiment -> embeddings -> trends ->
graph -> aggregation -> grounded AI report.

## Highlights

- **Clean architecture** - Handler -> Service -> Repository -> Database, with
  manual dependency injection in `cmd/*`; no business logic in handlers.
- **Real event pipeline** - Redis Streams with consumer groups, retries and
  dead-lettering; transport hidden behind interfaces so Kafka is a drop-in.
- **AI that cannot lie** - the summarizer only explains computed metrics, and
  a number-grounding guardrail rejects any report that invents a figure.
- **Graceful degradation** - the sentiment engine falls back to an in-process
  lexicon when the ML service is down; search degrades to a no-op; the demo
  stack runs fully offline with zero API keys.
- **Graph intelligence** - PageRank influence scores and label-propagation
  communities over reply/mention interactions.
- **Production mechanics** - embedded migrations, structured JSON logs,
  Prometheus metrics on every component, graceful shutdown, rate-bounded
  pagination, request correlation IDs, OpenAPI spec.
- **Deployment ready** - Docker Compose for the full stack (plus
  Elasticsearch and Prometheus profiles) and Kubernetes manifests.

## Architecture

```
connectors (twitter/telegram, plugin registry)
      │
      ▼
normalizer ─► Redis Streams ─► language detection ─► sentiment (ML + fallback)
      │                                                    │
      │                                                    ▼
      │                                             embeddings (pgvector)
      ▼                                                    │
scheduler: trends ─► graph (PageRank + communities) ─► analytics aggregation
      │                                                    │
      ▼                                                    ▼
  REST API (Gin + JWT + Swagger) ◄── Redis cache ◄── grounded AI report
```

See `docs/architecture.md` for the full picture and design rationale.

## Tech stack

| Concern | Choice |
| --- | --- |
| Language / framework | Go 1.22+ / Gin |
| Database | PostgreSQL 16 + pgvector |
| Query layer | SQLC-style typed queries (`sqlc generate`-compatible) |
| Cache + queue | Redis (cache + Redis Streams) |
| Search | Elasticsearch/OpenSearch abstraction (optional, no-op fallback) |
| AI services | Python FastAPI microservices (sentiment, embedding, summarizer) |
| Auth | JWT (HS256, bcrypt password hashing) |
| Docs | Hand-maintained OpenAPI 3.0 served at `/swagger` |
| Observability | structured JSON logs (zap) + Prometheus metrics |
| Deployment | Docker Compose + Kubernetes manifests |

## Repository layout

```
cmd/            api, worker, scheduler binaries
internal/       domain code (api, auth, ingestion, pipeline, sentiment, trends,
                graph, ai, analytics, database, cache, config, logger, search,
                topics, timeline, demographics, middleware)
ml-services/    FastAPI microservices + Dockerfiles
deployments/    docker-compose.yml, Dockerfiles, Kubernetes, Prometheus
docs/           architecture, database and API documentation
scripts/        deterministic demo seeder + helper scripts
tests/          black-box smoke assets
```

## Quickstart (Docker)

```bash
# 1. Boot the stack (Postgres, Redis, API, worker, scheduler, 3 ML services)
docker compose -f deployments/docker-compose.yml up --build

# API:      http://localhost:8080        (Swagger at /swagger)
# ML:       :8001 :8002 :8003            (health at /health)
```

Apply migrations happen automatically on API boot (`RUN_MIGRATIONS=true`).

```bash
# 2. Seed the deterministic demo data (1000+ posts, 3 scenarios)
python -m pip install -r scripts/requirements.txt
python scripts/seed_db.py --dsn "postgres://pulsegraph:pulsegraph@localhost:5432/pulsegraph?sslmode=disable"

# 3. Walk through the API
./scripts/demo.sh
```

The compose scheduler ships with `DEMO_INGEST_ENABLED=true`: it continuously
generates fresh posts every 30 seconds so sentiment, trends and the network
keep evolving during a live demo.

### Demo credentials

| Role | Email | Password |
| --- | --- | --- |
| admin | admin@pulsegraph.dev | PulseGraph@2026 |
| analyst | analyst@pulsegraph.dev | PulseGraph@2026 |
| viewer | viewer@pulsegraph.dev | PulseGraph@2026 |

### Optional profiles

```bash
docker compose -f deployments/docker-compose.yml --profile search up -d      # Elasticsearch
# then set SEARCH_ENABLED=true for the api service
docker compose -f deployments/docker-compose.yml --profile monitoring up -d  # Prometheus :9090
```

## Local development (no Docker for the Go code)

```bash
# infrastructure only
docker compose -f deployments/docker-compose.yml up -d postgres redis

# migrations + API
go run ./cmd/api            # applies migrations, serves :8080

# pipeline processes (separate terminals)
go run ./cmd/worker
go run ./cmd/scheduler

# ML services (separate terminals)
cd ml-services/sentiment-service  && uvicorn main:app --port 8001
cd ml-services/embedding-service  && uvicorn main:app --port 8002
cd ml-services/summarizer-service && uvicorn main:app --port 8003
```

## API overview

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/api/v1/auth/register` | create a dashboard user (JWT issued) |
| POST | `/api/v1/auth/login` | obtain a JWT |
| POST | `/api/v1/topics` | register a monitored conversation |
| GET | `/api/v1/topics` | list monitored topics |
| GET | `/api/v1/topics/:id/timeline` | posts + growth + important events |
| GET | `/api/v1/topics/:id/sentiment` | emotion distribution + sentiment timeline |
| GET | `/api/v1/topics/:id/demographics` | audience groups + language/region patterns |
| GET | `/api/v1/topics/:id/trends` | rising keywords + growth scores |
| GET | `/api/v1/topics/:id/network` | graph nodes/edges + influence + communities |
| GET | `/api/v1/topics/:id/dashboard` | complete intelligence incl. AI summary |
| GET | `/api/v1/search/posts?q=` | full-text search (when search is enabled) |

Full reference with request/response samples: `docs/api.md` and `/swagger`.

## AI services

| Service | Port | Endpoint | Notes |
| --- | --- | --- | --- |
| sentiment-service | 8001 | `POST /predict` | English + Hinglish lexicon, deterministic |
| embedding-service | 8002 | `POST /embed` | 384-dim hashing embedding (pgvector-ready) |
| summarizer-service | 8003 | `POST /generate-summary` | grounded reports only |

## Demo scenarios

1. **Flood misinformation (Assam)** - fear/anger arc peaks mid-window,
   recovers as relief coordination takes over.
2. **City Marathon 2026** - hype building to race day, route/traffic debate.
3. **Nimbus X1 Launch** - launch spike followed by mixed reviews.

Each scenario has its own language and region mix, interaction graph and
viral spikes ("important events"), so every dashboard view shows real
structure.

## Configuration

Environment-first (see `.env.example` for the full list). Key variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_PORT` | 8080 | API port |
| `DATABASE_URL` | postgres://pulsegraph:pulsegraph@localhost:5432/pulsegraph | PostgreSQL DSN |
| `REDIS_ADDR` | localhost:6379 | cache + streams |
| `JWT_SECRET` | dev fallback | **set in production** |
| `RUN_MIGRATIONS` | true | apply migrations on boot |
| `ML_SENTIMENT_URL` / `ML_EMBEDDING_URL` / `ML_SUMMARIZER_URL` | localhost:8001/8002/8003 | ML services |
| `SEARCH_ENABLED` | false | enable Elasticsearch/OpenSearch |
| `DEMO_INGEST_ENABLED` | false | scheduler generates demo posts continuously |
| `SCHEDULE_INTERVAL` | 2m | analytics recompute cadence |

## Testing

```bash
go test ./...        # unit tests (no external services required)
go vet ./...

# black-box smoke (stack running)
./tests/smoke.sh
```

## Monitoring

Prometheus metrics are served by every long-running component:

| Component | Endpoint |
| --- | --- |
| api | `:8080/metrics` |
| worker | `:9091/metrics` |
| scheduler | `:9092/metrics` |

Business metrics include `pulsegraph_posts_ingested_total`,
`pulsegraph_sentiment_processed_total`, `pulsegraph_embeddings_generated_total`,
`pulsegraph_pipeline_failures_total{stage=...}` and HTTP latency histograms.

## Documentation

- `docs/architecture.md` - components, pipeline, Kafka seam, design decisions
- `docs/database.md` - schema, indexes, migrations, pgvector usage
- `docs/api.md` - endpoint reference with samples and error codes
- `docs/` plus inline Swagger annotations on handlers

## Roadmap (production)

- Live connectors with real API credentials + backfill jobs
- Kafka transport swap (interfaces already in place)
- Model upgrades: transformer sentiment / sentence-transformers embeddings
- Anomaly alerts (spike detection -> webhook/Slack) and report scheduling
- Multi-tenant workspaces and per-topic RBAC
