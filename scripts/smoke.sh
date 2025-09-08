#!/usr/bin/env bash
set -euo pipefail

BASE=${BASE:-http://localhost:8080}
USER=${USER_ID:-1}

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

echo "== Ledger Smoke: Create (idempotent) =="
LEDGER_ID=$(curl -fsS -X POST "$BASE/api/ledger" \
  -H "X-User-Id: $USER" \
  -H "X-Idempotency-Key: abc123" \
  -H 'Content-Type: application/json' \
  -d '{"type":"dues","amount":50.00,"description":"Smoke test dues"}' | jq -r '.id')
echo "Created/returned ledger entry id=$LEDGER_ID"
# Replay idempotent request
curl -fsS -X POST "$BASE/api/ledger" \
  -H "X-User-Id: $USER" \
  -H "X-Idempotency-Key: abc123" \
  -H 'Content-Type: application/json' \
  -d '{"type":"dues","amount":50.00,"description":"Smoke test dues"}' | jq '{id,type,amount,description}'
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
  -H "X-User-Id: $USER" | jq '{id,is_read}'
echo

echo "== Announcements Smoke: Unread Count =="
curl -fsS "$BASE/api/announcements/unread?member_id=$USER" | jq '{member_id,unread_count}'
echo

echo "== Votes Smoke: Cast and Tally =="
PROP_ID=$(curl -fsS -X POST "$BASE/api/proposals" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Smoke vote","body":"test"}' | jq -r '.id')
# cast a vote
curl -fsS -X POST "$BASE/api/proposals/$PROP_ID/votes" \
  -H "X-User-Id: $USER" \
  -H 'Content-Type: application/json' \
  -d '{"choice":"for"}' | jq '{id,choice}'
# read tally
curl -fsS "$BASE/api/proposals/$PROP_ID/votes/tally" | jq '{proposal_id,results,quorum_met}'
echo

echo "Smoke OK ✅"


