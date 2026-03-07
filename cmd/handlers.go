package main

import (
	"net/http"
)

// handleHealth returns a minimal 200 response for health checks.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}

// handleAuthSession is a stub for POST /auth:session (login, refresh, logout).
func handleAuthSession(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented")
}

// handleAuthMeGet is a stub for GET /auth:me.
func handleAuthMeGet(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented")
}

// handleAuthMePost is a stub for POST /auth:me.
func handleAuthMePost(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented")
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
