-- backend/internal/announcements/migrations/0001_init.sql
CREATE TABLE IF NOT EXISTS announcements (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  author_id INTEGER,
  priority TEXT NOT NULL DEFAULT 'normal',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS announcement_reads (
  announcement_id INTEGER NOT NULL,
  member_id INTEGER NOT NULL,
  read_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (announcement_id, member_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS announcements_created_at_idx ON announcements (created_at DESC);
CREATE INDEX IF NOT EXISTS announcements_priority_idx ON announcements (priority);
CREATE INDEX IF NOT EXISTS announcements_author_id_idx ON announcements (author_id);

CREATE INDEX IF NOT EXISTS announcement_reads_member_id_idx ON announcement_reads (member_id);
CREATE INDEX IF NOT EXISTS announcement_reads_announcement_id_idx ON announcement_reads (announcement_id);

-- Foreign key relationships (will be enforced when members table is created)
-- For now, we use soft references via integer IDs

