// Package auth implements the Auth/RBAC module (architecture §2.2):
// users, sessions, and — in later tickets — OIDC/magic-link login and
// (user, role, resource_type, resource_id) grants.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

var (
	// ErrUserNotFound is returned when no user matches the given lookup.
	ErrUserNotFound = errors.New("auth: user not found")
	// ErrEmailTaken is returned by CreateUser when the email is already
	// registered — callers (AUTH-3/AUTH-4's login flows) need to
	// distinguish this from a generic failure to decide whether to look
	// the existing user up instead of creating a new one.
	ErrEmailTaken = errors.New("auth: email already registered")
)

type User struct {
	ID    string
	Email string
}

// CreateUser inserts a new user row. It does not decide identity-linking
// policy (OIDC vs. magic-link, anti-merge rules) — that's AUTH-5's job;
// this is just the storage primitive those tickets build on.
func CreateUser(ctx context.Context, conn *sql.DB, email string) (User, error) {
	var u User
	err := conn.QueryRowContext(ctx,
		`INSERT INTO users (email) VALUES ($1) RETURNING id, email`,
		email,
	).Scan(&u.ID, &u.Email)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("auth: create user: %w", err)
	}
	return u, nil
}

func GetUserByEmail(ctx context.Context, conn *sql.DB, email string) (User, error) {
	var u User
	err := conn.QueryRowContext(ctx,
		`SELECT id, email FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("auth: get user by email: %w", err)
	}
	return u, nil
}
