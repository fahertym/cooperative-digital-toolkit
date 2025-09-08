package announcements

import (
	"coop.tools/backend/internal/httpmw"
	"github.com/go-chi/chi/v5"
)

func Mount(r chi.Router, h Handlers) {
	route := func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.Get)
		r.With(httpmw.WithUser).Post("/{id}/read", h.MarkRead)
	}
	r.Route("/announcements", route)
	r.Get("/announcements/unread", h.UnreadCount)
}
