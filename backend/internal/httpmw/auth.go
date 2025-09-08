package httpmw

import (
    "context"
    "net/http"
    "strconv"
)

// Principal is the authenticated identity attached to a request.
// Role may be "admin", "member", or "guest" (unauthenticated).
type Principal struct {
    MemberID int64
    Role     string
    Email    string
    Name     string
}

// MemberFetcher looks up a member by id and returns a Principal.
// found=false indicates no such member. error indicates backend failure.
type MemberFetcher func(ctx context.Context, id int64) (p Principal, found bool, err error)

type ctxUserKey struct{}

// WithAuth parses X-User-Id and injects Principal into context.
// - Missing header -> guest principal
// - Present but invalid -> 401
// - Present but not found -> 401
func WithAuth(fetch MemberFetcher) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            s := r.Header.Get("X-User-Id")
            if s == "" {
                // guest
                ctx := context.WithValue(r.Context(), ctxUserKey{}, Principal{MemberID: 0, Role: "guest"})
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }
            id64, err := strconv.ParseInt(s, 10, 64)
            if err != nil || id64 <= 0 {
                WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
                return
            }
            p, found, err := fetch(r.Context(), id64)
            if err != nil {
                WriteJSONError(w, http.StatusInternalServerError, "auth lookup failed")
                return
            }
            if !found {
                WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
                return
            }
            ctx := context.WithValue(r.Context(), ctxUserKey{}, p)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// FromContext returns the Principal from context if present.
func FromContext(ctx context.Context) (Principal, bool) {
    v := ctx.Value(ctxUserKey{})
    if v == nil {
        return Principal{}, false
    }
    p, ok := v.(Principal)
    return p, ok
}

// RequireAuth ensures the current principal is not a guest.
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        p, ok := FromContext(r.Context())
        if !ok || p.Role == "guest" || p.MemberID <= 0 {
            WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
            return
        }
        next.ServeHTTP(w, r)
    })
}

// RequireRole ensures the current principal has one of the roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    allowed := make(map[string]struct{}, len(roles))
    for _, r := range roles { allowed[r] = struct{}{} }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            p, ok := FromContext(r.Context())
            if !ok || p.Role == "guest" || p.MemberID <= 0 {
                WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
                return
            }
            if _, ok := allowed[p.Role]; !ok {
                WriteJSONError(w, http.StatusForbidden, "forbidden")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// CurrentUserID returns (id, true) if a user id is present in the context.
// Kept for backward compatibility with existing handlers.
func CurrentUserID(ctx context.Context) (int32, bool) {
    if p, ok := FromContext(ctx); ok && p.MemberID > 0 {
        return int32(p.MemberID), true
    }
    v := ctx.Value(ctxUserKey{}) // legacy types not expected anymore
    if v == nil { return 0, false }
    switch t := v.(type) {
    case int32:
        return t, true
    }
    return 0, false
}
