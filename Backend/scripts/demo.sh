#!/usr/bin/env bash
# PulseGraph end-to-end demo walkthrough.
#
# Requires a running stack with seeded data:
#   docker compose -f deployments/docker-compose.yml up -d --build
#   python scripts/seed_db.py --dsn "postgres://pulsegraph:pulsegraph@localhost:5432/pulsegraph?sslmode=disable"
#   ./scripts/demo.sh
set -euo pipefail

API="${API:-http://localhost:8080/api/v1}"

echo "== Login (analyst@pulsegraph.dev) =="
TOKEN=$(curl -sf -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"analyst@pulsegraph.dev","password":"PulseGraph@2026"}' \
  | python -c "import sys,json;print(json.load(sys.stdin)['token'])")
AUTH="Authorization: Bearer $TOKEN"

echo
echo "== Topics =="
curl -sf -H "$AUTH" "$API/topics" | python -m json.tool | head -40

for id in 1 2 3; do
  echo
  echo "== Topic $id: dashboard summary =="
  curl -sf -H "$AUTH" "$API/topics/$id/dashboard" \
    | python -c "import sys,json;d=json.load(sys.stdin);s=d['summary'];print(s['title']);print('mode:',s['mode']);[print('-',f) for f in s['key_findings'][:3]]"
done

echo
echo "== Sentiment (topic 1) =="
curl -sf -H "$AUTH" "$API/topics/1/sentiment" \
  | python -c "import sys,json;d=json.load(sys.stdin);print('overall:',d['overall_score'],'analysed:',d['analyzed_posts'],'of',d['total_posts']);[print(' ',e['emotion'],e['count']) for e in d['distribution']]"

echo
echo "== Trends (topic 1) =="
curl -sf -H "$AUTH" "$API/topics/1/trends" \
  | python -c "import sys,json;d=json.load(sys.stdin);[print(' ',k['keyword'],round(k['trend_score'],2)) for k in d['rising_keywords'][:5]]"

echo
echo "== Network (topic 1) =="
curl -sf -H "$AUTH" "$API/topics/1/network" \
  | python -c "import sys,json;d=json.load(sys.stdin);print('nodes:',len(d['nodes']),'edges:',len(d['edges']),'communities:',d['community_count'])"

echo
echo "Done. Full payloads: GET /api/v1/topics/1/dashboard"
