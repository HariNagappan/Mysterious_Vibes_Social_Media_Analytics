# Scripts

| Script | Purpose |
| --- | --- |
| `seed_db.py` | Deterministic demo-data generator (1000+ posts, 3 scenarios). `--emit-sql` prints SQL; `--dsn` inserts directly. |
| `run_migrations.ps1` / `run_migrations.sh` | Apply embedded migrations via `api -migrate-only`. |
| `demo.sh` | End-to-end curl walkthrough (login → topics → dashboard → sentiment → trends → network). |

## Typical sequence

```bash
# 1. Start the stack
docker compose -f deployments/docker-compose.yml up -d --build

# 2. Seed demo data (the API applies migrations on boot by default)
python -m pip install -r scripts/requirements.txt
python scripts/seed_db.py --dsn "postgres://pulsegraph:pulsegraph@localhost:5432/pulsegraph?sslmode=disable"

# 3. Walk through the API
./scripts/demo.sh
```

`seed_db.py` is deterministic: same `--seed` (default 42) and same `--anchor`
produce byte-identical data, so every teammate's demo environment matches.
