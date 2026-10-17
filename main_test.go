package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want ok", body["status"])
	}
}

func TestRoot(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	want := "Hello from the DevOps week 5 container!"
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestUnknownPathReturns404(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPortFromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	if got := port(); got != "9090" {
		t.Fatalf("port() = %q, want 9090", got)
	}
	t.Setenv("PORT", "")
	if got := port(); got != "8080" {
		t.Fatalf("port() = %q, want 8080", got)
	}
}

func TestHealthcheck(t *testing.T) {
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	if code := healthcheck(srv.URL + "/healthz"); code != 0 {
		t.Fatalf("healthcheck on healthy server = %d, want 0", code)
	}
	if code := healthcheck(srv.URL + "/missing"); code != 1 {
		t.Fatalf("healthcheck on 404 = %d, want 1", code)
	}
}
