#!/usr/bin/env bash
set -euo pipefail

BASE="${BASE:-http://localhost:8080}"
HDR=( -H "Content-Type: application/json" -H "X-User-Id: 1" )

echo "== Seeding demo data =="
curl -fsS -X POST "$BASE/api/proposals" "${HDR[@]}" -d '{"title":"Demo","body":"Welcome"}' >/dev/null
curl -fsS -X POST "$BASE/api/announcements" "${HDR[@]}" -d '{"title":"Kickoff","body":"Pilot starts"}' >/dev/null
curl -fsS -X POST "$BASE/api/ledger" -H "X-Idempotency-Key: seed-1" "${HDR[@]}" -d '{"type":"dues","amount":10,"description":"Seed"}' >/dev/null
echo "Seeded demo data."

