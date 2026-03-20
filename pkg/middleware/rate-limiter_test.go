package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitMiddleware_AllowRequest(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:12345"

	rr := httptest.NewRecorder()
	RateLimitMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}
	if !handlerCalled {
		t.Errorf("expected handler to be called")
	}
}

func TestRateLimitMiddleware_BlockRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.2:12345"

	// Exhaust the limiter
	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		RateLimitMiddleware(handler).ServeHTTP(rr, req)
	}

	rr := httptest.NewRecorder()
	RateLimitMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %v", rr.Code)
	}

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %v", rr.Code)
	}
}

func TestGetClient_NewClient(t *testing.T) {
	ip := "192.0.2.3"
	limiter := getClient(ip)
	if limiter == nil {
		t.Fatal("expected a new limiter")
	}
}

func TestGetClient_ExistingClient(t *testing.T) {
	ip := "192.0.2.4"
	limiter1 := getClient(ip)
	time.Sleep(time.Millisecond * 10)
	limiter2 := getClient(ip)
	if limiter1 != limiter2 {
		t.Error("expected the same limiter for the same IP")
	}
}
