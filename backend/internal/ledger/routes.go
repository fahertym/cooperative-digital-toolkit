package ledger

import (
	"coop.tools/backend/internal/httpmw"
	"github.com/go-chi/chi/v5"
)

func Mount(r chi.Router, h Handlers) {
    route := func(r chi.Router) {
        r.Get("/", h.List)
        r.Get("/.csv", h.ExportCSV)
        r.With(httpmw.RequireAuth).Post("/", h.Create)
        r.Get("/{id}", h.Get)
    }
	r.Route("/ledger", route)
}
