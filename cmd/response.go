package main

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse is the standard envelope for successful API responses.
type SuccessResponse struct {
	Message string         `json:"message"`
	Data    []any          `json:"data,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
	Links   map[string]any `json:"links,omitempty"`
}

// ErrorResponse is the standard envelope for error API responses.
type ErrorResponse struct {
	Message string `json:"message"`
}

// CaptchaChallengeResponse is the documented CAPTCHA challenge envelope.
type CaptchaChallengeResponse struct {
	Message string              `json:"message"`
	Captcha CaptchaChallengeDTO `json:"captcha"`
}

// WriteJSON serializes body as JSON and writes it to w with the given status.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

// WriteError writes a standard error response with the given status and message.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Message: message})
}

// WriteCaptchaChallenge writes a CAPTCHA challenge response.
func WriteCaptchaChallenge(w http.ResponseWriter, status int, challenge CaptchaChallengeDTO) {
	WriteJSON(w, status, CaptchaChallengeResponse{
		Message: "Captcha required",
		Captcha: challenge,
	})
}

// WriteSuccess writes a standard success response with data.
func WriteSuccess(w http.ResponseWriter, status int, message string, data []any) {
	resp := SuccessResponse{
		Message: message,
		Data:    data,
	}
	WriteJSON(w, status, resp)
}

// WriteSuccessFull writes a success response with optional meta and links.
func WriteSuccessFull(w http.ResponseWriter, status int, message string, data []any, meta map[string]any, links map[string]any) {
	resp := SuccessResponse{
		Message: message,
		Data:    data,
		Meta:    meta,
		Links:   links,
	}
	WriteJSON(w, status, resp)
}

// WriteMessage writes a message-only success response (no data envelope).
func WriteMessage(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Message: message})
}
