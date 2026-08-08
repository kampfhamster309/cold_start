package auth

import (
	"context"
	"database/sql"
	"net/http"
)

type contextKey int

const userIDContextKey contextKey = iota

// UserIDFromContext returns the user ID set by RequireSession, if any.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
	return id, ok
}

// RequireSession resolves the request's session cookie into a user ID and
// stores it in the request context for downstream handlers/middleware
// (e.g. RequireGlobalRole). Rejects with 401 if the cookie is missing or
// the session is invalid/expired.
func RequireSession(conn *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := SessionTokenFromRequest(r)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := ValidateSession(r.Context(), conn, token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireGlobalRole rejects a request with 403 unless the user resolved
// by RequireSession (which must run earlier in the chain) holds a global
// grant for role. Phase 0's UI only exposes global roles (AUTH-6), so
// this is the only grant-checking middleware wired up for now; resource-
// scoped checks reuse HasGrant directly once Phase 2 needs them.
func RequireGlobalRole(conn *sql.DB, role Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			granted, err := HasGrant(r.Context(), conn, userID, role, ResourceGlobal, nil)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !granted {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
