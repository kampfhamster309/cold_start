package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

// unreachableDB opens a handle against a port nothing listens on. sql.Open
// doesn't connect eagerly, so this succeeds; the point is to exercise the
// health endpoint's unhealthy path without needing a real Postgres in CI.
func unreachableDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sql.Open("postgres", "postgres://user:pass@127.0.0.1:1/db?sslmode=disable&connect_timeout=1")
	if err != nil {
		t.Fatalf("open test db handle: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestHealthEndpoint_DatabaseUnreachable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	NewRouter(unreachableDB(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 when the database is unreachable, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
}
