package httpmw

import (
    "encoding/json"
    "net/http"
)

// WriteJSONError writes a standard JSON error response with the given status.
// Shape: {"error":"message"}
func WriteJSONError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(struct {
        Error string `json:"error"`
    }{Error: message})
}

