PR 1: Members domain

Database changes:
- Create `members` table with columns: `id BIGSERIAL PRIMARY KEY`, `email TEXT UNIQUE NOT NULL`, `display_name TEXT NOT NULL`, `role TEXT NOT NULL DEFAULT 'member'`, timestamps.
- Add CHECK constraint for `role` in ('admin','member').
- Add trigger to automatically set `updated_at` on UPDATE.

Rollback hints:
- `DROP TABLE IF EXISTS members CASCADE;`

