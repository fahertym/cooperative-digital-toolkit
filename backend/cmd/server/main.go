package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"coop.tools/backend/internal/announcements"
	"coop.tools/backend/internal/db"
	"coop.tools/backend/internal/ledger"
	"coop.tools/backend/internal/proposals"
	"coop.tools/backend/internal/votes"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	dsn := db.Env("DATABASE_URL", "postgres://coop:coop@localhost:5432/coopdb?sslmode=disable")
	store, err := db.Connect(ctx, dsn)
	if err != nil {
		log.Fatal("db connect:", err)
	}
	defer store.Close()

	// NEW: domain-owned migrations
	if err := proposals.ApplyMigrations(ctx, store.Pool); err != nil {
		log.Fatal("proposals migrations:", err)
	}
	if err := ledger.ApplyMigrations(ctx, store.Pool); err != nil {
		log.Fatal("ledger migrations:", err)
	}
	if err := announcements.ApplyMigrations(ctx, store.Pool); err != nil {
		log.Fatal("announcements migrations:", err)
	}
	if err := votes.ApplyMigrations(ctx, store.Pool); err != nil {
		log.Fatal("votes migrations:", err)
	}

	corsOrigin := db.Env("CORS_ORIGIN", "http://localhost:5173")

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{corsOrigin},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// API
	r.Route("/api", func(api chi.Router) {
		// Proposals
		propRepo := proposals.NewPgRepo(store.Pool)
		propHandlers := proposals.Handlers{Repo: propRepo}
		proposals.Mount(api, propHandlers)

		// Ledger
		ledgerRepo := ledger.NewPgRepo(store.Pool)
		ledgerHandlers := ledger.Handlers{Repo: ledgerRepo}
		ledger.Mount(api, ledgerHandlers)

		// Announcements
		announcementsRepo := announcements.NewPgRepo(store.Pool)
		announcementsHandlers := announcements.Handlers{Repo: announcementsRepo}
		announcements.Mount(api, announcementsHandlers)

		// Votes
		votesRepo := votes.NewPgRepo(store.Pool)
		votesHandlers := votes.Handlers{Repo: votesRepo}
		votes.Mount(api, votesHandlers)
	})

	addr := ":" + db.Env("PORT", "8080")
	log.Println("server listening on", addr)
	s := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}
