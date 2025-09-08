-- backend/internal/members/migrations/0001_members.sql
CREATE TABLE IF NOT EXISTS members (
  id BIGSERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'member',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Ensure role values are constrained
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname='members_role_chk'
  ) THEN
    ALTER TABLE members
      ADD CONSTRAINT members_role_chk CHECK (role IN ('admin','member'));
  END IF;
END$$;

-- Trigger to set updated_at on UPDATE
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname='members_set_updated_at'
  ) THEN
    CREATE TRIGGER members_set_updated_at
      BEFORE UPDATE ON members
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;

