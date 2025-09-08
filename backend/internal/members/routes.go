package members

import "github.com/go-chi/chi/v5"

func Mount(r chi.Router, h Handlers) {
    r.Route("/members", func(r chi.Router) {
        r.Get("/", h.FindByEmail) // expects ?email=
        r.Post("/", h.Create)
        r.Get("/{id}", h.GetByID)
    })
}

