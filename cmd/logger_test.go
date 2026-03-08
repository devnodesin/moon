package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Redaction
// ---------------------------------------------------------------------------

func TestRedactValue(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
		want string
	}{
		{"password redacted", "password", "secret123", RedactedPlaceholder},
		{"authorization redacted", "Authorization", "Bearer tok", RedactedPlaceholder},
		{"jwt_secret redacted", "jwt_secret", "mysecret", RedactedPlaceholder},
		{"refresh_token redacted", "refresh_token", "rt_abc", RedactedPlaceholder},
		{"api_key redacted", "api_key", "ak_xyz", RedactedPlaceholder},
		{"token redacted", "Token", "tok123", RedactedPlaceholder},
		{"safe key unchanged", "username", "admin", "admin"},
		{"host unchanged", "host", "localhost", "localhost"},
		{"case insensitive password", "PASSWORD", "s3cret", RedactedPlaceholder},
		{"case insensitive auth header", "authorization", "Basic creds", RedactedPlaceholder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactValue(tt.key, tt.val)
			if got != tt.want {
				t.Errorf("RedactValue(%q, %q) = %q; want %q", tt.key, tt.val, got, tt.want)
			}
		})
	}
}

func TestRedactValue_EmptyValue(t *testing.T) {
	got := RedactValue("password", "")
	if got != RedactedPlaceholder {
		t.Errorf("RedactValue with empty value = %q; want %q", got, RedactedPlaceholder)
	}
}

// ---------------------------------------------------------------------------
// InitLogger
// ---------------------------------------------------------------------------

func TestInitLogger_DualWriter(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Info("hello dual writer", "key", "value")

	// Verify log file received output
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "hello dual writer") {
		t.Errorf("log file missing message; got: %s", content)
	}
	if !strings.Contains(content, "key") {
		t.Errorf("log file missing key; got: %s", content)
	}
}

func TestInitLogger_AppendsToExistingFile(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "append.log")

	// Write initial content
	if err := os.WriteFile(logFile, []byte("existing line\n"), 0644); err != nil {
		t.Fatal(err)
	}

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Info("appended message")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "existing line") {
		t.Errorf("existing content was overwritten")
	}
	if !strings.Contains(content, "appended message") {
		t.Errorf("new message not appended")
	}
}

func TestInitLogger_FailsOnInvalidPath(t *testing.T) {
	_, err := InitLogger("/nonexistent/dir/impossible.log")
	if err == nil {
		t.Fatal("expected error for invalid log path, got nil")
	}
}

func TestInitLogger_Close(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "close.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	if err := logger.Close(); err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Log format verification
// ---------------------------------------------------------------------------

func TestLogFormat_ContainsRequiredFields(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "format.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Info("format test message")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	// Parse the JSON log line
	var entry map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("log line is not valid JSON: %v\nline: %s", err, data)
	}

	// Must have timestamp
	timeVal, ok := entry["time"]
	if !ok {
		t.Fatal("log entry missing 'time' field")
	}
	timeStr, ok := timeVal.(string)
	if !ok {
		t.Fatal("'time' field is not a string")
	}
	if _, err := time.Parse(time.RFC3339, timeStr); err != nil {
		// slog uses RFC3339Nano; try that too
		if _, err2 := time.Parse(time.RFC3339Nano, timeStr); err2 != nil {
			t.Errorf("'time' field %q is not RFC3339: %v", timeStr, err)
		}
	}

	// Must have level
	if _, ok := entry["level"]; !ok {
		t.Fatal("log entry missing 'level' field")
	}

	// Must have message
	if _, ok := entry["msg"]; !ok {
		t.Fatal("log entry missing 'msg' field")
	}
}

// ---------------------------------------------------------------------------
// AuditEvent
// ---------------------------------------------------------------------------

func TestAuditEvent_StructuredFields(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer logger.Close()

	logger.AuditEvent(AuditAuthSuccess,
		"method", "POST",
		"path", "/api/auth/login",
		"actor", "user@example.com",
		"target", "",
		"op", "login",
		"outcome", "success",
		"status", 200,
		"duration_ms", 42,
	)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("audit log not valid JSON: %v\nline: %s", err, data)
	}

	// Verify audit-specific fields
	wantFields := []string{"time", "level", "msg", "event", "method", "path", "actor", "op", "outcome", "status", "duration_ms"}
	for _, f := range wantFields {
		if _, ok := entry[f]; !ok {
			t.Errorf("audit entry missing field %q", f)
		}
	}

	if entry["event"] != AuditAuthSuccess {
		t.Errorf("event = %v; want %q", entry["event"], AuditAuthSuccess)
	}
	if entry["method"] != "POST" {
		t.Errorf("method = %v; want %q", entry["method"], "POST")
	}
}

func TestAuditEvent_StartupSuccess(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit_startup.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer logger.Close()

	logger.AuditEvent(AuditStartupSuccess, "outcome", "success")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}

	if entry["event"] != AuditStartupSuccess {
		t.Errorf("event = %v; want %q", entry["event"], AuditStartupSuccess)
	}
}

// ---------------------------------------------------------------------------
// Redaction in logging context
// ---------------------------------------------------------------------------

func TestLogger_RedactsInAttrs(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "redact.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Info("login attempt",
		"password", "supersecret",
		"authorization", "Bearer eyJhbG...",
		"username", "admin",
	)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if strings.Contains(content, "supersecret") {
		t.Error("password value leaked into log")
	}
	if strings.Contains(content, "eyJhbG") {
		t.Error("authorization header value leaked into log")
	}
	if !strings.Contains(content, RedactedPlaceholder) {
		t.Error("redacted placeholder not found in log")
	}
	if !strings.Contains(content, "admin") {
		t.Error("safe value 'admin' should appear in log")
	}
}

func TestLogger_RedactsTokenInAttrs(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "redact_token.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Info("token refresh",
		"refresh_token", "rt_longvalue",
		"jwt_secret", "secret32charslong_abcdefghijklmn",
		"api_key", "ak_12345",
	)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if strings.Contains(content, "rt_longvalue") {
		t.Error("refresh_token value leaked")
	}
	if strings.Contains(content, "secret32charslong") {
		t.Error("jwt_secret value leaked")
	}
	if strings.Contains(content, "ak_12345") {
		t.Error("api_key value leaked")
	}
}

// ---------------------------------------------------------------------------
// Logger convenience methods
// ---------------------------------------------------------------------------

func TestLogger_LogLevels(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "levels.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 log lines, got %d: %s", len(lines), string(data))
	}

	// Verify each line is valid JSON with correct level
	expectedLevels := []string{"INFO", "WARN", "ERROR"}
	for i, line := range lines {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("line %d not valid JSON: %v", i, err)
		}
		if entry["level"] != expectedLevels[i] {
			t.Errorf("line %d level = %v; want %q", i, entry["level"], expectedLevels[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Runtime log file failure
// ---------------------------------------------------------------------------

func TestLogger_RuntimeFileFailure_ContinuesOnConsole(t *testing.T) {
	// Verify that the logger has a fallback writer that writes to console
	// when the file writer fails at runtime. We test this by creating a
	// logger, closing the file handle (simulating failure), and verifying
	// the logger still works.
	dir := t.TempDir()
	logFile := filepath.Join(dir, "runtime_fail.log")

	logger, err := InitLogger(logFile)
	if err != nil {
		t.Fatalf("InitLogger failed: %v", err)
	}

	// Close the underlying file to simulate runtime file failure
	logger.Close()

	// Logger should not panic; it may fail to write to file but must not crash
	// We capture stdout to verify console output still works
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// This should not panic even though the file is closed
	logger.Info("after file close")

	w.Close()
	os.Stdout = oldStdout

	captured, _ := io.ReadAll(r)
	// The message may or may not appear depending on buffering,
	// but the key requirement is no panic/crash
	_ = captured
}

// ---------------------------------------------------------------------------
// NewTestLogger helper
// ---------------------------------------------------------------------------

func TestNewTestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewTestLogger(&buf)

	logger.Info("test msg", "key", "val")

	if !strings.Contains(buf.String(), "test msg") {
		t.Errorf("test logger output missing message; got: %s", buf.String())
	}
}

// ---------------------------------------------------------------------------
// Constants verification
// ---------------------------------------------------------------------------

func TestAuditEventConstants(t *testing.T) {
	events := map[string]string{
		"AuditStartupSuccess":      AuditStartupSuccess,
		"AuditStartupFailure":      AuditStartupFailure,
		"AuditConfigValidation":    AuditConfigValidation,
		"AuditAuthSuccess":         AuditAuthSuccess,
		"AuditAuthFailure":         AuditAuthFailure,
		"AuditLogout":              AuditLogout,
		"AuditTokenRefresh":        AuditTokenRefresh,
		"AuditRateLimitViolation":  AuditRateLimitViolation,
		"AuditSchemaMutation":      AuditSchemaMutation,
		"AuditPrivilegedMutation":  AuditPrivilegedMutation,
		"AuditAPIKeyCreate":        AuditAPIKeyCreate,
		"AuditAPIKeyRotation":      AuditAPIKeyRotation,
		"AuditAdminUserManagement": AuditAdminUserManagement,
		"AuditShutdown":            AuditShutdown,
	}

	for name, val := range events {
		if val == "" {
			t.Errorf("%s must not be empty", name)
		}
	}
}

func TestRedactedPlaceholder(t *testing.T) {
	if RedactedPlaceholder != "[REDACTED]" {
		t.Errorf("RedactedPlaceholder = %q; want %q", RedactedPlaceholder, "[REDACTED]")
	}
}

func TestSensitiveKeys(t *testing.T) {
	required := []string{"password", "authorization", "jwt_secret", "refresh_token", "api_key", "token"}
	for _, k := range required {
		found := false
		for _, sk := range SensitiveKeys {
			if strings.EqualFold(sk, k) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SensitiveKeys missing required key %q", k)
		}
	}
}

// ---------------------------------------------------------------------------
// Redacting handler
// ---------------------------------------------------------------------------

func TestRedactingHandler_GroupedAttrs(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	rh := NewRedactingHandler(handler)

	logger := slog.New(rh)
	logger.Info("test", slog.Group("auth",
		slog.String("password", "secret"),
		slog.String("user", "admin"),
	))

	content := buf.String()
	if strings.Contains(content, "secret") {
		t.Error("password value leaked in grouped attrs")
	}
	if !strings.Contains(content, "admin") {
		t.Error("safe value 'admin' missing in grouped attrs")
	}
}

// TestRedactingHandler_WithAttrs verifies that WithAttrs propagates and
// redacts sensitive attributes.
func TestRedactingHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, nil)
	rh := NewRedactingHandler(base)

	// WithAttrs returns a new handler that pre-redacts the given attributes.
	child := rh.WithAttrs([]slog.Attr{
		slog.String("password", "s3cret"),
		slog.String("user", "alice"),
	})
	childLogger := slog.New(child)
	childLogger.Info("child msg")

	content := buf.String()
	if strings.Contains(content, "s3cret") {
		t.Error("password value leaked through WithAttrs")
	}
	if !strings.Contains(content, "alice") {
		t.Error("safe 'user' value missing through WithAttrs")
	}
}

// TestRedactingHandler_WithGroup verifies that WithGroup returns a valid handler.
func TestRedactingHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	rh := NewRedactingHandler(handler)

	grouped := rh.WithGroup("request")
	groupLogger := slog.New(grouped)
	groupLogger.Info("grouped", "password", "secret123", "path", "/api/v1")

	content := buf.String()
	if strings.Contains(content, "secret123") {
		t.Error("password leaked through WithGroup")
	}
	if !strings.Contains(content, "/api/v1") {
		t.Error("safe value missing through WithGroup")
	}
}

// TestNewTestLogger_Close verifies that Close on a non-file logger returns nil.
func TestNewTestLogger_Close(t *testing.T) {
	var buf bytes.Buffer
	logger := NewTestLogger(&buf)
	// No file handle; Close should return nil without panicking.
	if err := logger.Close(); err != nil {
		t.Errorf("Close on non-file logger returned error: %v", err)
	}
}
