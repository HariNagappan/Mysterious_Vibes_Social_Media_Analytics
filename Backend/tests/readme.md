# Tests

Unit tests live next to their code (Go convention) and run with:

```bash
go test ./...
```

Covered modules include JWT handling, language detection, trend detection,
PageRank/community detection, the sentiment lexicon fallback, the AI report
grounding guardrail, demographics estimation, request validation and the
ingestion normalizer. Tests do not require PostgreSQL or Redis.

This folder holds cross-cutting black-box assets:

| File | Purpose |
| --- | --- |
| `smoke.sh` | Hits the operational endpoints of a locally running stack. |

```bash
# stack running on localhost:8080
./tests/smoke.sh
```
