// Package redact provides utilities for redacting sensitive data from logs
// and error responses to prevent credential leakage (PRD-049).
package redact

import (
	"strings"
)

// RedactedPlaceholder is the value used to replace sensitive data.
const RedactedPlaceholder = "***REDACTED***"

// DefaultSensitiveFields lists the default field names considered sensitive.
// Field names are matched case-insensitively.
var DefaultSensitiveFields = []string{
	"password",
	"token",
	"secret",
	"api_key",
	"apikey",
	"authorization",
	"jwt",
	"refresh_token",
	"access_token",
	"client_secret",
	"private_key",
	"credential",
	"auth",
	"bearer",
}

// Redactor handles sensitive data redaction.
type Redactor struct {
	sensitiveFields map[string]bool
}

// New creates a new Redactor with the default sensitive fields.
func New() *Redactor {
	r := &Redactor{
		sensitiveFields: make(map[string]bool),
	}
	for _, field := range DefaultSensitiveFields {
		r.sensitiveFields[strings.ToLower(field)] = true
	}
	return r
}

// NewWithFields creates a new Redactor with additional sensitive fields.
func NewWithFields(additionalFields []string) *Redactor {
	r := New()
	for _, field := range additionalFields {
		r.sensitiveFields[strings.ToLower(field)] = true
	}
	return r
}

// IsSensitiveField checks if a field name is sensitive (case-insensitive).
func (r *Redactor) IsSensitiveField(fieldName string) bool {
	return r.sensitiveFields[strings.ToLower(fieldName)]
}

// RedactMap redacts sensitive fields from a map recursively.
// Returns a new map with sensitive values replaced by RedactedPlaceholder.
func (r *Redactor) RedactMap(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}

	result := make(map[string]any)
	for key, value := range data {
		if r.IsSensitiveField(key) {
			result[key] = RedactedPlaceholder
		} else {
			result[key] = r.redactValue(value)
		}
	}
	return result
}

// redactValue recursively redacts sensitive data from a value.
func (r *Redactor) redactValue(value any) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		return r.RedactMap(v)
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = r.redactValue(item)
		}
		return result
	default:
		return value
	}
}

// RedactString redacts a value if the field name is sensitive.
// Returns RedactedPlaceholder if sensitive, otherwise returns the original value.
func (r *Redactor) RedactString(fieldName, value string) string {
	if r.IsSensitiveField(fieldName) {
		return RedactedPlaceholder
	}
	return value
}

// Global default redactor for convenience.
var defaultRedactor = New()

// IsSensitive checks if a field name is sensitive using the default redactor.
func IsSensitive(fieldName string) bool {
	return defaultRedactor.IsSensitiveField(fieldName)
}

// Map redacts sensitive fields from a map using the default redactor.
func Map(data map[string]any) map[string]any {
	return defaultRedactor.RedactMap(data)
}

// String redacts a value if sensitive using the default redactor.
func String(fieldName, value string) string {
	return defaultRedactor.RedactString(fieldName, value)
}
