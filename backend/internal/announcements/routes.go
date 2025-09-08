package announcements

import (
    "coop.tools/backend/internal/httpmw"
    "github.com/go-chi/chi/v5"
)

func Mount(r chi.Router, h Handlers) {
    route := func(r chi.Router) {
        r.Get("/", h.List)
        r.With(httpmw.RequireRole("admin")).Post("/", h.Create)
        r.Get("/{id}", h.Get)
        r.With(httpmw.RequireAuth).Post("/{id}/read", h.MarkRead)
    }
    r.Route("/announcements", route)
    r.Get("/announcements/unread", h.UnreadCount)
}
