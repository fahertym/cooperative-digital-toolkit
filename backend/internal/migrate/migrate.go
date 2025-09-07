package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Apply runs all *.sql files found in dir of fsys, in lexicographic order.
// Each file is applied once per namespace and recorded in schema_migrations(namespace, version).
func Apply(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, dir string, namespace string) error {
	if err := ensureTable(ctx, pool); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("readdir %s: %w", dir, err)
	}

	// Sort by filename to ensure order 0001_..., 0002_...
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		version := e.Name()

		applied, err := alreadyApplied(ctx, pool, namespace, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		b, err := fs.ReadFile(fsys, dir+"/"+version)
		if err != nil {
			return fmt.Errorf("read %s: %w", version, err)
		}

		// One transaction per migration.
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		if _, err := tx.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("exec %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (namespace, version, applied_at) VALUES ($1,$2,$3)`,
			namespace, version, time.Now().UTC(),
		); err != nil {
			return fmt.Errorf("record %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", version, err)
		}
	}

	return nil
}

func ensureTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  namespace TEXT NOT NULL,
  version   TEXT NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY(namespace, version)
);`)
	return err
}

func alreadyApplied(ctx context.Context, pool *pgxpool.Pool, ns, ver string) (bool, error) {
	var exists bool
	if err := pool.QueryRow(ctx,
		`SELECT true FROM schema_migrations WHERE namespace=$1 AND version=$2`,
		ns, ver,
	).Scan(&exists); err == nil {
		return true, nil
	} else {
		// err from Scan likely means no rows; treat as not applied
		return false, nil
	}
}
