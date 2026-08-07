package auth

import (
	"errors"
	"net/http"
	"time"
)

const cookieName = "cold_start_session"

var ErrNoSessionCookie = errors.New("auth: no session cookie present")

// SetSessionCookie writes the session token as an HttpOnly cookie.
// secure should be true whenever the app is served over HTTPS (the
// production case, once OPS-2 configures TLS) — browsers silently drop
// Secure cookies over plain HTTP, which is still how this runs locally
// via Caddy on :80 (INFRA-3), so it's a parameter, not hardcoded true.
func SetSessionCookie(w http.ResponseWriter, token string, secure bool, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	})
}

// ClearSessionCookie expires the cookie immediately — used on logout and
// by the anti-merge rule's session-invalidation path.
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func SessionTokenFromRequest(r *http.Request) (string, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return "", ErrNoSessionCookie
	}
	return c.Value, nil
}
