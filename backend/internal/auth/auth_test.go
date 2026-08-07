package auth_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"cold_start/backend/internal/auth"
	"cold_start/backend/internal/db"
)

// testDB connects to TEST_DATABASE_URL and runs migrations against it.
// Skips (not fails) when unset, so `go test ./...` works without a
// Postgres running locally; CI sets it against a service container.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping DB-backed test")
	}

	conn, err := db.Connect(url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return conn
}

// uniqueEmail avoids collisions between test runs against a persistent
// (not per-test-truncated) database.
func uniqueEmail(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%s-%d@example.test", t.Name(), time.Now().UnixNano())
}

func TestCreateAndGetUser(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	email := uniqueEmail(t)

	created, err := auth.CreateUser(ctx, conn, email)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.Email != email {
		t.Fatalf("expected email %q, got %q", email, created.Email)
	}

	fetched, err := auth.GetUserByEmail(ctx, conn, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if fetched.ID != created.ID {
		t.Fatalf("expected id %q, got %q", created.ID, fetched.ID)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()
	email := uniqueEmail(t)

	if _, err := auth.CreateUser(ctx, conn, email); err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}
	if _, err := auth.CreateUser(ctx, conn, email); err != auth.ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	_, err := auth.GetUserByEmail(ctx, conn, uniqueEmail(t))
	if err != auth.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	user, err := auth.CreateUser(ctx, conn, uniqueEmail(t))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	token, err := auth.CreateSession(ctx, conn, user.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty session token")
	}

	gotUserID, err := auth.ValidateSession(ctx, conn, token)
	if err != nil {
		t.Fatalf("ValidateSession: %v", err)
	}
	if gotUserID != user.ID {
		t.Fatalf("expected user id %q, got %q", user.ID, gotUserID)
	}

	if err := auth.InvalidateSession(ctx, conn, token); err != nil {
		t.Fatalf("InvalidateSession: %v", err)
	}
	if _, err := auth.ValidateSession(ctx, conn, token); err != auth.ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid after invalidation, got %v", err)
	}
}

func TestValidateSession_Expired(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	user, err := auth.CreateUser(ctx, conn, uniqueEmail(t))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// Negative TTL: already expired the moment it's created.
	token, err := auth.CreateSession(ctx, conn, user.ID, -time.Minute)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if _, err := auth.ValidateSession(ctx, conn, token); err != auth.ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid for an expired session, got %v", err)
	}
}

func TestValidateSession_UnknownToken(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	if _, err := auth.ValidateSession(ctx, conn, "not-a-real-token"); err != auth.ErrSessionInvalid {
		t.Fatalf("expected ErrSessionInvalid for an unknown token, got %v", err)
	}
}

func TestInvalidateAllSessionsForUser(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	user, err := auth.CreateUser(ctx, conn, uniqueEmail(t))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	tokenA, err := auth.CreateSession(ctx, conn, user.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession A: %v", err)
	}
	tokenB, err := auth.CreateSession(ctx, conn, user.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession B: %v", err)
	}

	if err := auth.InvalidateAllSessionsForUser(ctx, conn, user.ID); err != nil {
		t.Fatalf("InvalidateAllSessionsForUser: %v", err)
	}

	for _, tok := range []string{tokenA, tokenB} {
		if _, err := auth.ValidateSession(ctx, conn, tok); err != auth.ErrSessionInvalid {
			t.Fatalf("expected session invalidated, got %v", err)
		}
	}
}
