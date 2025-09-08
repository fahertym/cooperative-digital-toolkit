package announcements

import "github.com/go-chi/chi/v5"

func Mount(r chi.Router, h Handlers) {
	route := func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/unread", h.GetUnreadCount)
		r.Get("/{id}", h.Get)
		r.Post("/{id}/read", h.MarkAsRead)
	}
	r.Route("/announcements", route)
}
