package votes

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"

	"coop.tools/backend/internal/migrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// ApplyMigrations applies this domain's SQL files in order.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	return migrate.Apply(ctx, pool, migrationsFS, "migrations", "votes")
}
