package members

import (
    "context"
    "embed"

    "github.com/jackc/pgx/v5/pgxpool"

    "coop.tools/backend/internal/migrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// ApplyMigrations applies the members domain SQL migrations.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
    return migrate.Apply(ctx, pool, migrationsFS, "migrations", "members")
}

