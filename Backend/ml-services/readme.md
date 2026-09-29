# PulseGraph ML services

Python FastAPI microservices that give PulseGraph its AI capabilities. All
services are self-contained and run fully offline — no model downloads, no
API keys — so the demo stack boots anywhere.

| Service | Port | Endpoint | Purpose |
| --- | --- | --- | --- |
| sentiment-service | 8001 | POST /predict | Emotion, sentiment score and confidence for a text |
| embedding-service | 8002 | POST /embed | Deterministic 384-dim hashing embedding (matches pgvector) |
| summarizer-service | 8003 | POST /generate-summary | Grounded, analytics-only report generation |

Every service exposes `GET /health` and ships its own `Dockerfile` and
`requirements.txt`.

## Grounding rule

The summarizer **only explains metrics it was given** — it never invents
facts. The template renderer copies numbers straight from the payload, and
the Go client (`internal/ai`) additionally runs a number-grounding guardrail
on any ML-generated report before serving it, falling back to the local
template on violations.

## Run

```bash
# via compose (from the repository root)
docker compose -f deployments/docker-compose.yml up --build sentiment-service embedding-service summarizer-service

# or directly
cd sentiment-service && uvicorn main:app --port 8001
```
