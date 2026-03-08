package main

import (
	"net/http"
	"time"
)

// healthData is the response body for the health endpoint.
type healthData struct {
	Moon      string `json:"moon"`
	Timestamp string `json:"timestamp"`
}

// healthResponse is the top-level response envelope for the health endpoint.
type healthResponse struct {
	Data healthData `json:"data"`
}

// handleHealth returns the service version and current UTC timestamp.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Data: healthData{
			Moon:      MoonVersion,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	WriteJSON(w, http.StatusOK, resp)
}

// newAuthSessionHandler creates the AuthSessionHandler with its dependencies.
// logger and rl may be nil; rate limiting is skipped when rl is nil.
func newAuthSessionHandler(db DatabaseAdapter, cfg *AppConfig, logger *Logger, rl *RateLimiter) *AuthSessionHandler {
	return &AuthSessionHandler{db: db, cfg: cfg, logger: logger, rateLimiter: rl}
}

// handleCollectionsQuery is a stub for GET /collections:query.
func handleCollectionsQuery(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented")
}

// handleCollectionsMutate is a stub for POST /collections:mutate.
func handleCollectionsMutate(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented")
}

// handleResourceQuery is a stub for GET /data/{resource}:query.
func handleResourceQuery(w http.ResponseWriter, r *http.Request) {
	resource := extractResource(r.URL.Path)
	if resource == "" {
		WriteError(w, http.StatusBadRequest, "Missing resource name")
		return
	}
	WriteError(w, http.StatusNotImplemented, "Not implemented")
}

// handleResourceMutate is a stub for POST /data/{resource}:mutate.
func handleResourceMutate(w http.ResponseWriter, r *http.Request) {
	resource := extractResource(r.URL.Path)
	if resource == "" {
		WriteError(w, http.StatusBadRequest, "Missing resource name")
		return
	}
	WriteError(w, http.StatusNotImplemented, "Not implemented")
}

// handleResourceSchema is a stub for GET /data/{resource}:schema.
func handleResourceSchema(w http.ResponseWriter, r *http.Request) {
	resource := extractResource(r.URL.Path)
	if resource == "" {
		WriteError(w, http.StatusBadRequest, "Missing resource name")
		return
	}
	WriteError(w, http.StatusNotImplemented, "Not implemented")
}
