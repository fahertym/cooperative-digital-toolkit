PR 1: Members domain

Database changes:
- Create `members` table with columns: `id BIGSERIAL PRIMARY KEY`, `email TEXT UNIQUE NOT NULL`, `display_name TEXT NOT NULL`, `role TEXT NOT NULL DEFAULT 'member'`, timestamps.
- Add CHECK constraint for `role` in ('admin','member').
- Add trigger to automatically set `updated_at` on UPDATE.

Rollback hints:
- `DROP TABLE IF EXISTS members CASCADE;`

---

PR 2: Announcements FKs and consistency

Database changes:
- Convert `announcements.author_id` to BIGINT, backfill nulls to a `system@local` admin if present, enforce `NOT NULL`, and add FK to `members(id)` with `ON DELETE RESTRICT`.
- Convert `announcement_reads.member_id` to BIGINT and add FK to `members(id)` with `ON DELETE CASCADE`.

Rollback hints:
- `ALTER TABLE announcement_reads DROP CONSTRAINT IF EXISTS announcement_reads_member_fk;`
- `ALTER TABLE announcements DROP CONSTRAINT IF EXISTS announcements_author_fk;`
- `ALTER TABLE announcements ALTER COLUMN author_id DROP NOT NULL;`
