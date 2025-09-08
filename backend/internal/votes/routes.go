package votes

import (
	"coop.tools/backend/internal/httpmw"
	"github.com/go-chi/chi/v5"
)

func Mount(r chi.Router, h Handlers) {
	route := func(r chi.Router) {
		r.Get("/", h.List)
		r.With(httpmw.WithUser).Post("/", h.Create)
		r.With(httpmw.WithUser).Put("/", h.Update)
		r.Get("/tally", h.GetTally)
	}
	r.Route("/proposals/{proposal_id}/votes", route)
}
