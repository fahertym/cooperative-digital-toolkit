-- backend/internal/announcements/migrations/0002_constraints.sql
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'announcements_priority_chk'
  ) THEN
    ALTER TABLE announcements
      ADD CONSTRAINT announcements_priority_chk
      CHECK (priority IN ('low', 'normal', 'high', 'urgent'));
  END IF;
END$$;

-- Ensure title and body are not empty
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'announcements_title_not_empty_chk'
  ) THEN
    ALTER TABLE announcements
      ADD CONSTRAINT announcements_title_not_empty_chk
      CHECK (length(trim(title)) > 0);
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'announcements_body_not_empty_chk'
  ) THEN
    ALTER TABLE announcements
      ADD CONSTRAINT announcements_body_not_empty_chk
      CHECK (length(trim(body)) > 0);
  END IF;
END$$;

-- Ensure updated_at is >= created_at
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'announcements_updated_at_chk'
  ) THEN
    ALTER TABLE announcements
      ADD CONSTRAINT announcements_updated_at_chk
      CHECK (updated_at >= created_at);
  END IF;
END$$;

-- Add foreign key constraint for announcement_reads -> announcements
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'announcement_reads_announcement_id_fk'
  ) THEN
    ALTER TABLE announcement_reads
      ADD CONSTRAINT announcement_reads_announcement_id_fk
      FOREIGN KEY (announcement_id) REFERENCES announcements(id) ON DELETE CASCADE;
  END IF;
END$$;

