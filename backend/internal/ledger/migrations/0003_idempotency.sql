-- Add idempotency support for ledger entries
ALTER TABLE ledger_entries
  ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

-- Unique per member to support replay safety; only when key is provided
CREATE UNIQUE INDEX IF NOT EXISTS ux_ledger_idem_member
  ON ledger_entries (member_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;


