-- backend/internal/proposals/migrations/0002_status_check.sql
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'proposals_status_chk'
  ) THEN
    ALTER TABLE proposals
      ADD CONSTRAINT proposals_status_chk
      CHECK (status IN ('open','closed','archived'));
  END IF;
END$$;


