package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"
)

// CaptchaChallengeDTO is the API-visible CAPTCHA challenge payload.
type CaptchaChallengeDTO struct {
	ID          string `json:"id"`
	ImageBase64 string `json:"image_base64"`
	ExpiresIn   int    `json:"expires_in"`
}

type captchaChallenge struct {
	answer    string
	expiresAt time.Time
}

// CaptchaStore stores short-lived CAPTCHA challenges in memory.
type CaptchaStore struct {
	mu         sync.Mutex
	challenges map[string]captchaChallenge
	now        func() time.Time
}

// NewCaptchaStore creates a new empty CAPTCHA store.
func NewCaptchaStore() *CaptchaStore {
	return &CaptchaStore{
		challenges: make(map[string]captchaChallenge),
		now:        time.Now,
	}
}

// Issue creates and stores a new CAPTCHA challenge.
func (s *CaptchaStore) Issue() (CaptchaChallengeDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now().UTC()
	s.cleanupExpiredLocked(now)

	answer, err := randomDigits(CaptchaCodeLength)
	if err != nil {
		return CaptchaChallengeDTO{}, err
	}

	id := GenerateULID()
	s.challenges[id] = captchaChallenge{
		answer:    answer,
		expiresAt: now.Add(time.Duration(CaptchaTTLSeconds) * time.Second),
	}

	return CaptchaChallengeDTO{
		ID:          id,
		ImageBase64: base64.StdEncoding.EncodeToString([]byte(renderCaptchaSVG(answer))),
		ExpiresIn:   CaptchaTTLSeconds,
	}, nil
}

// Validate verifies and consumes a CAPTCHA challenge.
func (s *CaptchaStore) Validate(id, answer string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now().UTC()
	s.cleanupExpiredLocked(now)

	challenge, ok := s.challenges[id]
	if !ok {
		return false
	}
	delete(s.challenges, id)

	return challenge.expiresAt.After(now) && strings.EqualFold(strings.TrimSpace(answer), challenge.answer)
}

func (s *CaptchaStore) cleanupExpiredLocked(now time.Time) {
	for id, challenge := range s.challenges {
		if !challenge.expiresAt.After(now) {
			delete(s.challenges, id)
		}
	}
}

func randomDigits(length int) (string, error) {
	if length < 1 {
		return "", fmt.Errorf("captcha length must be positive")
	}

	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate captcha digits: %w", err)
	}

	var b strings.Builder
	b.Grow(length)
	for _, value := range buf {
		b.WriteByte('0' + (value % 10))
	}
	return b.String(), nil
}

func renderCaptchaSVG(answer string) string {
	var lineBuilder strings.Builder
	for idx := range 5 {
		x1 := 12 + idx*40
		y1 := 18 + idx*7
		x2 := CaptchaImageWidth - (idx * 32) - 18
		y2 := CaptchaImageHeight - (idx * 9) - 12
		lineBuilder.WriteString(
			fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#cbd5e1" stroke-width="2" />`,
				x1, y1, x2, y2,
			),
		)
	}

	var textBuilder strings.Builder
	for idx, char := range answer {
		x := 26 + idx*34
		y := 52 + (idx%2)*6
		rotation := (idx % 3) - 1
		textBuilder.WriteString(
			fmt.Sprintf(`<text x="%d" y="%d" transform="rotate(%d %d %d)" font-size="34" font-family="monospace" fill="#0f172a">%c</text>`,
				x, y, rotation*9, x, y, char,
			),
		)
	}

	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d"><rect width="100%%" height="100%%" fill="#f8fafc" rx="8" ry="8" />%s%s</svg>`,
		CaptchaImageWidth,
		CaptchaImageHeight,
		CaptchaImageWidth,
		CaptchaImageHeight,
		lineBuilder.String(),
		textBuilder.String(),
	)
}
