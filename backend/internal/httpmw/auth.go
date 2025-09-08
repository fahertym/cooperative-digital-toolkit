package httpmw

import (
	"context"
	"net/http"
	"strconv"
)

type ctxUserKey struct{}

// WithUser enforces presence of X-User-Id and injects it into context.
func WithUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := r.Header.Get("X-User-Id")
		if s == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id, err := strconv.Atoi(s)
		if err != nil || id <= 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey{}, int32(id))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CurrentUserID returns (id, true) if a user id is present in the context.
func CurrentUserID(ctx context.Context) (int32, bool) {
	v := ctx.Value(ctxUserKey{})
	if v == nil {
		return 0, false
	}
	id, ok := v.(int32)
	return id, ok
}
