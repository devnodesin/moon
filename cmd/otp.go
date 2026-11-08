package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// GenerateOTP returns a numeric OTP code and its SHA-256 hash.
func GenerateOTP(length int) (string, string, error) {
	if length <= 0 {
		length = OTPCodeLength
	}
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	h := sha256.Sum256([]byte(code))
	return code, hex.EncodeToString(h[:]), nil
}

// StoreOTP inserts an OTP row.
func StoreOTP(ctx context.Context, db DatabaseAdapter, userID, codeHash, purpose string, expiryMinutes int) error {
	if expiryMinutes <= 0 {
		expiryMinutes = OTPTTLMinutes
	}
	now := time.Now().UTC()
	expires := now.Add(time.Duration(expiryMinutes) * time.Minute)
	return db.InsertRow(ctx, "moon_auth_otp_codes", map[string]any{
		"id":          GenerateULID(),
		"user_id":     userID,
		"code_hash":   codeHash,
		"purpose":     purpose,
		"expires_at":  expires.Format(time.RFC3339),
		"created_at":  now.Format(time.RFC3339),
		"consumed_at": nil,
	})
}

var otpVerifyMu sync.Mutex

// VerifyOTP checks the code against the latest unconsumed OTP for the user.
// Returns true on success and marks the OTP consumed.
func VerifyOTP(ctx context.Context, db DatabaseAdapter, userID, code, purpose string) (bool, error) {
	otpVerifyMu.Lock()
	defer otpVerifyMu.Unlock()
	codeHash := sha256.Sum256([]byte(code))
	hashHex := hex.EncodeToString(codeHash[:])

	rows, _, err := db.QueryRows(ctx, "moon_auth_otp_codes", QueryOptions{
		Filters: []Filter{
			{Field: "user_id", Op: "eq", Value: userID},
			{Field: "purpose", Op: "eq", Value: purpose},
		},
		Sort:    []SortField{{Field: "created_at", Desc: true}},
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		return false, err
	}

	now := time.Now().UTC()
	var candidate map[string]any
	for _, r := range rows {
		if r["consumed_at"] != nil {
			continue
		}
		exp, err := time.Parse(time.RFC3339, stringVal(r, "expires_at"))
		if err != nil || now.After(exp) {
			continue
		}
		if stringVal(r, "code_hash") == hashHex {
			candidate = r
			break
		}
	}
	if candidate == nil {
		return false, nil
	}

	id, _ := candidate["id"].(string)
	nowStr := now.Format(time.RFC3339)
	if err := db.UpdateRow(ctx, "moon_auth_otp_codes", id, map[string]any{
		"consumed_at": nowStr,
	}); err != nil {
		return false, err
	}
	// Consume any remaining active OTPs for this user/purpose to prevent
	// reuse of stale codes.
	for _, r := range rows {
		rid, _ := r["id"].(string)
		if rid == "" || rid == id {
			continue
		}
		if r["consumed_at"] != nil {
			continue
		}
		_ = db.UpdateRow(ctx, "moon_auth_otp_codes", rid, map[string]any{
			"consumed_at": nowStr,
		})
	}
	return true, nil
}

// CleanupExpiredOTPs deletes expired or consumed OTP rows.
func CleanupExpiredOTPs(ctx context.Context, db DatabaseAdapter) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if strings.ContainsAny(now, "'\\") {
		return fmt.Errorf("invalid timestamp")
	}
	// Bulk delete expired or consumed OTP rows in a single statement.
	// ExecDDL is used here for raw SQL execution (both DDL and DML).
	return db.ExecDDL(ctx, fmt.Sprintf(
		"DELETE FROM moon_auth_otp_codes WHERE expires_at < '%s' OR consumed_at IS NOT NULL", now))
}
