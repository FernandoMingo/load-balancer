package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"FernandoMingo/load-balancer/pkg/middleware"
)

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "OK")
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	middleware.RecoveryMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "OK" {
		t.Errorf("unexpected response body: %q", body)
	}
}

func TestRecoveryMiddleware_WithPanic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	middleware.RecoveryMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Internal Server Error") {
		t.Errorf("expected response body to include 'Internal Server Error', got %q", rr.Body.String())
	}
}
