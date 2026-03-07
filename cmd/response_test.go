package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	body := map[string]string{"key": "value"}
	WriteJSON(w, http.StatusOK, body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("expected JSON content type, got %q", ct)
	}

	var got map[string]string
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got["key"] != "value" {
		t.Fatalf("expected key=value, got %v", got)
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		message string
	}{
		{"bad request", http.StatusBadRequest, "Bad request"},
		{"not found", http.StatusNotFound, "Not found"},
		{"internal error", http.StatusInternalServerError, "Internal server error"},
		{"method not allowed", http.StatusMethodNotAllowed, "Method not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, tt.status, tt.message)

			if w.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, w.Code)
			}

			var got ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}
			if got.Message != tt.message {
				t.Fatalf("expected message %q, got %q", tt.message, got.Message)
			}
		})
	}
}

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	data := []any{map[string]string{"id": "abc"}}
	WriteSuccess(w, http.StatusOK, "Resources retrieved successfully", data)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var got SuccessResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got.Message != "Resources retrieved successfully" {
		t.Fatalf("unexpected message: %q", got.Message)
	}
	if len(got.Data) != 1 {
		t.Fatalf("expected 1 data item, got %d", len(got.Data))
	}
}

func TestWriteSuccessFull(t *testing.T) {
	w := httptest.NewRecorder()
	data := []any{map[string]string{"id": "abc"}}
	meta := map[string]any{"total": 42}
	links := map[string]any{"next": "/page/2"}
	WriteSuccessFull(w, http.StatusOK, "OK", data, meta, links)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got SuccessResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got.Meta["total"] == nil {
		t.Fatal("expected meta.total to be present")
	}
	if got.Links["next"] == nil {
		t.Fatal("expected links.next to be present")
	}
}

func TestWriteSuccess_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	WriteSuccess(w, http.StatusOK, "OK", nil)

	var got map[string]any
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if _, ok := got["data"]; ok {
		t.Fatal("expected data to be omitted when nil")
	}
}

func TestWriteMessage(t *testing.T) {
	w := httptest.NewRecorder()
	WriteMessage(w, http.StatusOK, "Logged out successfully")

	var got ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got.Message != "Logged out successfully" {
		t.Fatalf("unexpected message: %q", got.Message)
	}
}

func TestWriteSuccess_Created(t *testing.T) {
	w := httptest.NewRecorder()
	data := []any{map[string]string{"id": "new-id"}}
	WriteSuccess(w, http.StatusCreated, "Created", data)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
