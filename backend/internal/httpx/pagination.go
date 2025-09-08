package httpx

import (
    "net/http"
    "strconv"
)

// ParseLimitOffset parses limit and offset with sane caps.
// If missing, returns (0,0) which callers can treat as defaults.
func ParseLimitOffset(r *http.Request, max int) (limit, offset int, err error) {
    if ls := r.URL.Query().Get("limit"); ls != "" {
        v, e := strconv.Atoi(ls)
        if e != nil || v <= 0 {
            return 0, 0, eIf(e, ErrInvalidParam)
        }
        if v > max { v = max }
        limit = v
    }
    if os := r.URL.Query().Get("offset"); os != "" {
        v, e := strconv.Atoi(os)
        if e != nil || v < 0 {
            return 0, 0, eIf(e, ErrInvalidParam)
        }
        offset = v
    }
    return
}

// ErrInvalidParam is a sentinel for invalid query parameters.
var ErrInvalidParam = errInvalidParam{}

type errInvalidParam struct{}
func (errInvalidParam) Error() string { return "invalid parameter" }

func eIf(e error, def error) error { if e != nil { return e }; return def }

