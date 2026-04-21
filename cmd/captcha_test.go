package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCaptchaStore_IssueAndValidate(t *testing.T) {
	store := NewCaptchaStore()
	store.now = func() time.Time {
		return time.Date(2026, 4, 21, 7, 0, 0, 0, time.UTC)
	}

	challenge, err := store.Issue()
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if challenge.ID == "" {
		t.Fatal("expected challenge id")
	}
	if challenge.ImageBase64 == "" {
		t.Fatal("expected image data")
	}

	answer := store.challenges[challenge.ID].answer
	if !store.Validate(challenge.ID, answer) {
		t.Fatal("expected challenge to validate")
	}
	if store.Validate(challenge.ID, answer) {
		t.Fatal("expected challenge to be consumed")
	}
}

func TestCaptchaStore_ExpiredChallenge(t *testing.T) {
	store := NewCaptchaStore()
	now := time.Date(2026, 4, 21, 7, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	challenge, err := store.Issue()
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(time.Duration(CaptchaTTLSeconds+1) * time.Second)
	answer := store.challenges[challenge.ID].answer
	if store.Validate(challenge.ID, answer) {
		t.Fatal("expected expired challenge to fail")
	}
}

func TestRandomDigits_InvalidLength(t *testing.T) {
	if _, err := randomDigits(0); err == nil {
		t.Fatal("expected error for invalid length")
	}
}

func TestCaptchaMiddleware(t *testing.T) {
	store := NewCaptchaStore()
	store.now = func() time.Time {
		return time.Date(2026, 4, 21, 7, 0, 0, 0, time.UTC)
	}

	innerCalled := false
	handler := captchaMiddleware(store, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("challenge issued when missing", func(t *testing.T) {
		innerCalled = false
		req := httptest.NewRequest(http.MethodPost, "/data/contact:mutate", bytes.NewBufferString(`{"op":"create","data":[{}]}`))
		req = req.WithContext(SetAuthIdentity(req.Context(), &AuthIdentity{
			CredentialType:  CredentialTypeAPIKey,
			CallerID:        "key-1",
			CaptchaRequired: true,
		}))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
		if innerCalled {
			t.Fatal("inner handler should not be called")
		}
	})

	t.Run("valid captcha passes", func(t *testing.T) {
		challenge, err := store.Issue()
		if err != nil {
			t.Fatalf("Issue: %v", err)
		}
		answer := store.challenges[challenge.ID].answer

		innerCalled = false
		req := httptest.NewRequest(http.MethodPost, "/data/contact:mutate", bytes.NewBufferString(`{"captcha_id":"`+challenge.ID+`","captcha_value":"`+answer+`","op":"create","data":[{}]}`))
		req = req.WithContext(SetAuthIdentity(req.Context(), &AuthIdentity{
			CredentialType:  CredentialTypeAPIKey,
			CallerID:        "key-1",
			CaptchaRequired: true,
		}))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !innerCalled {
			t.Fatal("expected inner handler to be called")
		}
	})

	t.Run("non-api-key request bypasses captcha", func(t *testing.T) {
		innerCalled = false
		req := httptest.NewRequest(http.MethodPost, "/data/contact:mutate", bytes.NewBufferString(`{"op":"create","data":[{}]}`))
		req = req.WithContext(SetAuthIdentity(req.Context(), &AuthIdentity{
			CredentialType: CredentialTypeJWT,
			CallerID:       "user-1",
		}))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !innerCalled {
			t.Fatal("expected inner handler to be called")
		}
	})
}
