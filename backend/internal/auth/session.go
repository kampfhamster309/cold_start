package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// DefaultSessionTTL is the sliding session lifetime. Not a product
// decision made anywhere in the planning docs (checked 04-open-items.md
// — not listed); a reasonable default, tunable later without a schema
// change since expiry is a per-row column, not derived from a constant.
const DefaultSessionTTL = 30 * 24 * time.Hour

// ErrSessionInvalid covers both "no such session" and "session expired"
// — callers (auth middleware) treat both identically (reject the
// request), so there's no reason to distinguish them at this layer.
var ErrSessionInvalid = errors.New("auth: session invalid or expired")

// CreateSession issues a new session for userID and returns the raw
// token to set as a cookie. Only the token's SHA-256 hash is persisted.
func CreateSession(ctx context.Context, conn *sql.DB, userID string, ttl time.Duration) (token string, err error) {
	token, err = generateToken()
	if err != nil {
		return "", fmt.Errorf("auth: generate session token: %w", err)
	}

	_, err = conn.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		hashToken(token), userID, time.Now().Add(ttl),
	)
	if err != nil {
		return "", fmt.Errorf("auth: create session: %w", err)
	}
	return token, nil
}

// ValidateSession looks up the session by token and returns the owning
// user's ID if it exists and hasn't expired.
func ValidateSession(ctx context.Context, conn *sql.DB, token string) (userID string, err error) {
	var expiresAt time.Time
	err = conn.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM sessions WHERE token_hash = $1`,
		hashToken(token),
	).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrSessionInvalid
	}
	if err != nil {
		return "", fmt.Errorf("auth: validate session: %w", err)
	}
	if time.Now().After(expiresAt) {
		return "", ErrSessionInvalid
	}
	return userID, nil
}

// InvalidateSession deletes a session outright — used by the anti-merge
// rule (AUTH-5) and account-recovery events, which invalidate existing
// sessions rather than letting them expire naturally.
func InvalidateSession(ctx context.Context, conn *sql.DB, token string) error {
	_, err := conn.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, hashToken(token))
	if err != nil {
		return fmt.Errorf("auth: invalidate session: %w", err)
	}
	return nil
}

// InvalidateAllSessionsForUser is the session-invalidation half of the
// anti-merge rule (tech-stack §6): any account-recovery-equivalent event
// invalidates *all* of a user's existing sessions, not just the current
// one.
func InvalidateAllSessionsForUser(ctx context.Context, conn *sql.DB, userID string) error {
	_, err := conn.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("auth: invalidate sessions for user: %w", err)
	}
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
