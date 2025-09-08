package httpx

import (
    "net/http"
    "strconv"
)

// QueryString returns the raw query value for key.
func QueryString(r *http.Request, key string) string {
    return r.URL.Query().Get(key)
}

// QueryInt64 parses an optional int64 from query parameters.
// Returns (nil, nil) when missing.
func QueryInt64(r *http.Request, key string) (*int64, error) {
    s := r.URL.Query().Get(key)
    if s == "" { return nil, nil }
    v, err := strconv.ParseInt(s, 10, 64)
    if err != nil { return nil, err }
    return &v, nil
}

// QueryBoolTrue returns true only if the query param equals "true".
func QueryBoolTrue(r *http.Request, key string) bool {
    return r.URL.Query().Get(key) == "true"
}

