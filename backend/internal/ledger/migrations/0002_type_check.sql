-- backend/internal/ledger/migrations/0002_type_check.sql
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ledger_entries_type_chk'
  ) THEN
    ALTER TABLE ledger_entries
      ADD CONSTRAINT ledger_entries_type_chk
      CHECK (type IN ('dues', 'contribution', 'expense', 'income'));
  END IF;
END$$;

-- Add amount constraint to prevent negative amounts for dues/contributions
-- and allow negative amounts for expenses
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ledger_entries_amount_chk'
  ) THEN
    ALTER TABLE ledger_entries
      ADD CONSTRAINT ledger_entries_amount_chk
      CHECK (amount != 0);
  END IF;
END$$;


