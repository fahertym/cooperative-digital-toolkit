-- backend/internal/announcements/migrations/0003_member_fks.sql
-- Normalize member references to BIGINT and enforce FKs to members(id).

-- Ensure announcements.author_id is BIGINT
ALTER TABLE announcements
  ALTER COLUMN author_id TYPE BIGINT USING author_id::bigint;

-- Backfill null author_id by creating a system admin member and assigning it
DO $$
DECLARE sys_id BIGINT;
BEGIN
  IF EXISTS (SELECT 1 FROM announcements WHERE author_id IS NULL) THEN
    -- Create system member if not exists
    INSERT INTO members (email, display_name, role)
    VALUES ('system@local', 'System', 'admin')
    ON CONFLICT (email) DO NOTHING;

    SELECT id INTO sys_id FROM members WHERE email='system@local';
    UPDATE announcements SET author_id = sys_id WHERE author_id IS NULL;
  END IF;
END$$;

-- Enforce NOT NULL and FK (idempotent guards)
DO $$
BEGIN
  BEGIN
    ALTER TABLE announcements ALTER COLUMN author_id SET NOT NULL;
  EXCEPTION WHEN others THEN
    -- ignore if already set or constrained by other rules
  END;
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname='announcements_author_fk'
  ) THEN
    ALTER TABLE announcements
      ADD CONSTRAINT announcements_author_fk
      FOREIGN KEY (author_id) REFERENCES members(id) ON DELETE RESTRICT;
  END IF;
END$$;

-- Ensure announcement_reads.member_id is BIGINT and has FK to members(id)
ALTER TABLE announcement_reads
  ALTER COLUMN member_id TYPE BIGINT USING member_id::bigint;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname='announcement_reads_member_fk'
  ) THEN
    ALTER TABLE announcement_reads
      ADD CONSTRAINT announcement_reads_member_fk
      FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE;
  END IF;
END$$;

