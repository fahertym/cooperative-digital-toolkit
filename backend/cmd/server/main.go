package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"

	"coop.tools/backend/internal/db"
)

type Proposal struct {
	ID        int32     `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	_ = godotenv.Load() // .env optional

	ctx := context.Background()
	dsn := db.Env("DATABASE_URL", "postgres://coop:coop@localhost:5432/coopdb?sslmode=disable")
	store, err := db.Connect(ctx, dsn)
	if err != nil {
		log.Fatal("db connect:", err)
	}
	defer store.Close()

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

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/proposals", func(w http.ResponseWriter, r *http.Request) {
			rows, err := store.Pool.Query(ctx, `SELECT id, title, COALESCE(body,''), created_at FROM proposals ORDER BY id DESC`)
			if err != nil {
				http.Error(w, "query failed", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var list []Proposal
			for rows.Next() {
				var p Proposal
				var createdAt pgtype.Timestamptz
				if err := rows.Scan(&p.ID, &p.Title, &p.Body, &createdAt); err != nil {
					http.Error(w, "scan failed", http.StatusInternalServerError)
					return
				}
				p.CreatedAt = createdAt.Time
				list = append(list, p)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(list)
		})

		r.Post("/proposals", func(w http.ResponseWriter, r *http.Request) {
			var in struct {
				Title string `json:"title"`
				Body  string `json:"body"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if len(in.Title) == 0 {
				http.Error(w, "title required", http.StatusBadRequest)
				return
			}
			var p Proposal
			err := store.Pool.QueryRow(ctx, `
INSERT INTO proposals (title, body) VALUES ($1,$2)
RETURNING id, title, COALESCE(body,''), created_at
`, in.Title, in.Body).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt)
			if err != nil {
				http.Error(w, "insert failed", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(p)
		})

		r.Get("/proposals/{id}", func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			id, _ := strconv.Atoi(idStr)
			var p Proposal
			err := store.Pool.QueryRow(ctx, `
SELECT id, title, COALESCE(body,''), created_at FROM proposals WHERE id=$1
`, id).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt)
			if err != nil {
				if err == pgx.ErrNoRows {
					http.NotFound(w, r)
					return
				}
				http.Error(w, "query failed", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(p)
		})
	})

	addr := ":" + db.Env("PORT", "8080")
	log.Println("server listening on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
