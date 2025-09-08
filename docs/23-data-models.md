# Data Models

## proposals
- `id SERIAL PRIMARY KEY`
- `title TEXT NOT NULL`
- `body TEXT`
- `status TEXT CHECK (status IN ('open','closed')) NOT NULL DEFAULT 'open'`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`

## votes
- `id SERIAL PRIMARY KEY`
- `proposal_id INT NOT NULL REFERENCES proposals(id) ON DELETE CASCADE`
- `member_id INT NOT NULL`
- `choice TEXT CHECK (choice IN ('for','against','abstain')) NOT NULL`
- `notes TEXT`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- Uniqueness: `UNIQUE (proposal_id, member_id)`
- Indexes: `(proposal_id)`, `(member_id)`

## announcements
- `id SERIAL PRIMARY KEY`
- `title TEXT NOT NULL`
- `body TEXT NOT NULL`
- `priority TEXT CHECK (priority IN ('low','normal','high','urgent')) NOT NULL DEFAULT 'normal'`
- `author_id INT` nullable
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`

### announcement_reads
- `announcement_id INT NOT NULL REFERENCES announcements(id) ON DELETE CASCADE`
- `member_id INT NOT NULL`
- `read_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- Primary key: `(announcement_id, member_id)`
- Indexes: `(member_id)`, `(announcement_id)`

## ledger_entries
- `id SERIAL PRIMARY KEY`
- `member_id INT` nullable (associated via auth header at write time)
- `type TEXT CHECK (type IN ('dues','contribution','expense','income')) NOT NULL`
- `amount NUMERIC(12,2) NOT NULL CHECK (amount != 0)`
- `description TEXT NOT NULL`
- `notes TEXT`
- `idempotency_key TEXT` nullable
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- Partial unique index: `UNIQUE (member_id, idempotency_key) WHERE idempotency_key IS NOT NULL`
- Indexes: `(member_id)`, `(created_at)`, `(type)`

## CSV formats

### proposals
- Header row: `id,title,body,status,created_at`
- Timestamps RFC3339

### ledger_entries
- Columns and order: `Date,Description,Type,Amount,Member ID,Notes,Reference`
- Date = `created_at` formatted `YYYY-MM-DD`
