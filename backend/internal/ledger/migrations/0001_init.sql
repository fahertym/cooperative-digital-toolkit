-- backend/internal/ledger/migrations/0001_init.sql
CREATE TABLE IF NOT EXISTS ledger_entries (
  id SERIAL PRIMARY KEY,
  type TEXT NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  description TEXT NOT NULL,
  member_id INTEGER,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ledger_entries_created_at_idx ON ledger_entries (created_at DESC);
CREATE INDEX IF NOT EXISTS ledger_entries_type_idx ON ledger_entries (type);
CREATE INDEX IF NOT EXISTS ledger_entries_member_id_idx ON ledger_entries (member_id);


