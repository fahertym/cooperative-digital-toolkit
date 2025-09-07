package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type Proposal struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type store struct {
	mu        sync.Mutex
	proposals []Proposal
	nextID    int
}

func main() {
	s := &store{proposals: []Proposal{}, nextID: 1}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
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
			w.Header().Set("Content-Type", "application/json")
			s.mu.Lock()
			defer s.mu.Unlock()
			json.NewEncoder(w).Encode(s.proposals)
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
			s.mu.Lock()
			p := Proposal{
				ID:        s.nextID,
				Title:     in.Title,
				Body:      in.Body,
				CreatedAt: time.Now().UTC(),
			}
			s.nextID++
			s.proposals = append(s.proposals, p)
			s.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(p)
		})
		// tiny helper to fetch one proposal if you want it later:
		r.Get("/proposals/{id}", func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			id, _ := strconv.Atoi(idStr)
			s.mu.Lock()
			defer s.mu.Unlock()
			for _, p := range s.proposals {
				if p.ID == id {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(p)
					return
				}
			}
			http.NotFound(w, r)
		})
	})

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Println("server listening on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
