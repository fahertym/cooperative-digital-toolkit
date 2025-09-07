package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"coop.tools/backend/internal/db"
	"coop.tools/backend/internal/proposals"
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

	// Keep using db.AutoMigrate for now to ensure the table exists.
	if err := store.AutoMigrate(ctx); err != nil {
		log.Fatal("db migrate:", err)
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
		propRepo := proposals.NewPgRepo(store.Pool)
		propHandlers := proposals.Handlers{Repo: propRepo}
		proposals.Mount(api, propHandlers)
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
