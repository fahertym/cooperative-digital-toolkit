#!/usr/bin/env bash
set -euo pipefail

BASE=${BASE:-http://localhost:8080}

echo "== Smoke: Health =="
curl -fsS "$BASE/healthz" && echo "ok" || (echo "health failed" && exit 1)
echo

echo "== Smoke: List (JSON array) =="
curl -fsS "$BASE/api/proposals" | jq 'length' || (echo "list failed" && exit 1)
echo

echo "== Smoke: Create =="
ID=$(curl -fsS -X POST "$BASE/api/proposals" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Smoke run","body":"from script"}' | jq -r '.id')
echo "Created id=$ID"
echo

echo "== Smoke: Close =="
curl -fsS -X POST "$BASE/api/proposals/$ID/close" | jq -r '.status'
echo

echo "== Smoke: Get =="
curl -fsS "$BASE/api/proposals/$ID" | jq '{id,title,status}'
echo

echo "== Smoke: CSV header =="
curl -fsS "$BASE/api/proposals/.csv" | head -n 1
echo

echo "Smoke OK ✅"


