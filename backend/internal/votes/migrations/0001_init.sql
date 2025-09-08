CREATE TABLE IF NOT EXISTS votes (
  id SERIAL PRIMARY KEY,
  proposal_id INTEGER NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
  member_id INTEGER NOT NULL,
  choice TEXT NOT NULL CHECK (choice IN ('for', 'against', 'abstain')),
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(proposal_id, member_id)
);

CREATE INDEX IF NOT EXISTS idx_votes_proposal_id ON votes(proposal_id);
CREATE INDEX IF NOT EXISTS idx_votes_member_id ON votes(member_id);
