package votes

import "github.com/go-chi/chi/v5"

func Mount(r chi.Router, h Handlers) {
	route := func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Put("/", h.Update)
		r.Get("/tally", h.GetTally)
	}
	r.Route("/proposals/{proposal_id}/votes", route)
}
