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

echo "== Ledger Smoke: List (JSON array) =="
curl -fsS "$BASE/api/ledger" | jq 'length' || (echo "ledger list failed" && exit 1)
echo

echo "== Ledger Smoke: Create =="
LEDGER_ID=$(curl -fsS -X POST "$BASE/api/ledger" \
  -H 'Content-Type: application/json' \
  -d '{"type":"dues","amount":50.00,"description":"Smoke test dues","member_id":1}' | jq -r '.id')
echo "Created ledger entry id=$LEDGER_ID"
echo

echo "== Ledger Smoke: Get =="
curl -fsS "$BASE/api/ledger/$LEDGER_ID" | jq '{id,type,amount,description}'
echo

echo "== Ledger Smoke: CSV header =="
curl -fsS "$BASE/api/ledger/.csv" | head -n 1
echo

echo "== Announcements Smoke: List (JSON array) =="
curl -fsS "$BASE/api/announcements" | jq 'length' || (echo "announcements list failed" && exit 1)
echo

echo "== Announcements Smoke: Create =="
ANNOUNCEMENT_ID=$(curl -fsS -X POST "$BASE/api/announcements" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Smoke test announcement","body":"This is a smoke test announcement","priority":"normal"}' | jq -r '.id')
echo "Created announcement id=$ANNOUNCEMENT_ID"
echo

echo "== Announcements Smoke: Get =="
curl -fsS "$BASE/api/announcements/$ANNOUNCEMENT_ID" | jq '{id,title,priority,is_read}'
echo

echo "== Announcements Smoke: Mark as Read =="
curl -fsS -X POST "$BASE/api/announcements/$ANNOUNCEMENT_ID/read" \
  -H 'Content-Type: application/json' \
  -d '{"member_id":1}' | jq '{id,is_read}'
echo

echo "== Announcements Smoke: Unread Count =="
curl -fsS "$BASE/api/announcements/unread?member_id=1" | jq '{member_id,unread_count}'
echo

echo "Smoke OK ✅"


