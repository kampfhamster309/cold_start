package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cold_start/backend/internal/auth"
)

func TestSessionCookieRoundTrip(t *testing.T) {
	rec := httptest.NewRecorder()
	auth.SetSessionCookie(rec, "the-token", true, time.Hour)

	result := rec.Result()
	var cookie *http.Cookie
	for _, c := range result.Cookies() {
		if c.Name == "cold_start_session" {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("expected a cold_start_session cookie to be set")
	}
	if !cookie.HttpOnly {
		t.Error("expected HttpOnly")
	}
	if !cookie.Secure {
		t.Error("expected Secure when secure=true was passed")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite=Lax, got %v", cookie.SameSite)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)

	token, err := auth.SessionTokenFromRequest(req)
	if err != nil {
		t.Fatalf("SessionTokenFromRequest: %v", err)
	}
	if token != "the-token" {
		t.Fatalf("expected token %q, got %q", "the-token", token)
	}
}

func TestSessionTokenFromRequest_NoCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := auth.SessionTokenFromRequest(req)
	if err != auth.ErrNoSessionCookie {
		t.Fatalf("expected ErrNoSessionCookie, got %v", err)
	}
}

func TestClearSessionCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	auth.ClearSessionCookie(rec, true)

	result := rec.Result()
	var cookie *http.Cookie
	for _, c := range result.Cookies() {
		if c.Name == "cold_start_session" {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("expected a cold_start_session cookie to be set")
	}
	if cookie.MaxAge >= 0 {
		t.Errorf("expected a negative MaxAge to clear the cookie, got %d", cookie.MaxAge)
	}
}
