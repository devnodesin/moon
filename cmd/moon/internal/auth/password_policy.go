package auth

import (
	"fmt"
	"strings"
	"unicode"
)

// PasswordPolicy defines the password validation rules.
type PasswordPolicy struct {
	MinLength          int
	RequireUppercase   bool
	RequireLowercase   bool
	RequireDigit       bool
	RequireSpecialChar bool
	SpecialChars       string
}

// DefaultPasswordPolicy returns the default password policy.
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:          8,
		RequireUppercase:   true,
		RequireLowercase:   true,
		RequireDigit:       true,
		RequireSpecialChar: false,
		SpecialChars:       "!@#$%^&*()_+-=[]{}|;':\",./<>?",
	}
}

// Validate checks if the password meets the policy requirements.
func (p *PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters", p.MinLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(p.SpecialChars, r):
			hasSpecial = true
		}
	}

	if p.RequireUppercase && !hasUpper {
		return fmt.Errorf("password must include at least one uppercase letter")
	}

	if p.RequireLowercase && !hasLower {
		return fmt.Errorf("password must include at least one lowercase letter")
	}

	if p.RequireDigit && !hasDigit {
		return fmt.Errorf("password must include at least one number")
	}

	if p.RequireSpecialChar && !hasSpecial {
		return fmt.Errorf("password must include at least one special character")
	}

	return nil
}

// ValidationErrors returns a list of all validation failures.
func (p *PasswordPolicy) ValidationErrors(password string) []string {
	var errors []string

	if len(password) < p.MinLength {
		errors = append(errors, fmt.Sprintf("minimum %d characters required", p.MinLength))
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(p.SpecialChars, r):
			hasSpecial = true
		}
	}

	if p.RequireUppercase && !hasUpper {
		errors = append(errors, "uppercase letter required")
	}

	if p.RequireLowercase && !hasLower {
		errors = append(errors, "lowercase letter required")
	}

	if p.RequireDigit && !hasDigit {
		errors = append(errors, "number required")
	}

	if p.RequireSpecialChar && !hasSpecial {
		errors = append(errors, "special character required")
	}

	return errors
}
