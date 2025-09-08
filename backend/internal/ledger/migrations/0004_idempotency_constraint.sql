-- Add non-partial UNIQUE constraint to support ON CONFLICT inference
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ledger_idem_member_unique'
  ) THEN
    ALTER TABLE ledger_entries
      ADD CONSTRAINT ledger_idem_member_unique
      UNIQUE (member_id, idempotency_key);
  END IF;
END$$;


