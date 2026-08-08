package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cold_start/backend/internal/auth"
)

// chain builds RequireSession -> RequireGlobalRole(role) -> a 200 OK leaf
// handler, matching how a real protected route would be wired.
func chain(conn *sql.DB, role auth.Role) http.Handler {
	leaf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return auth.RequireSession(conn)(auth.RequireGlobalRole(conn, role)(leaf))
}

func TestRequireGlobalRole_NoCookie(t *testing.T) {
	conn := testDB(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	chain(conn, auth.RoleAdmin).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with no session cookie, got %d", rec.Code)
	}
}

func TestRequireGlobalRole_InvalidSessionToken(t *testing.T) {
	conn := testDB(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "cold_start_session", Value: "not-a-real-token"})
	chain(conn, auth.RoleAdmin).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with an invalid session token, got %d", rec.Code)
	}
}

func TestRequireGlobalRole_ValidSessionNoGrant(t *testing.T) {
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

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "cold_start_session", Value: token})
	chain(conn, auth.RoleAdmin).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a valid session without the required grant, got %d", rec.Code)
	}
}

func TestRequireGlobalRole_ValidSessionWithGrant(t *testing.T) {
	conn := testDB(t)
	ctx := context.Background()

	user, err := auth.CreateUser(ctx, conn, uniqueEmail(t))
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if err := auth.GrantRole(ctx, conn, user.ID, auth.RoleAdmin, auth.ResourceGlobal, nil); err != nil {
		t.Fatalf("GrantRole: %v", err)
	}
	token, err := auth.CreateSession(ctx, conn, user.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "cold_start_session", Value: token})
	chain(conn, auth.RoleAdmin).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for a valid session with the required grant, got %d", rec.Code)
	}
}
