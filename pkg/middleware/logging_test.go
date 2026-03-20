package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Example handler for testing
func testHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

func TestLoggingMiddleware_Success(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := LoggingMiddleware(testHandler(http.StatusOK), logger)

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	output := buf.String()
	if !strings.Contains(output, "Started GET /test-path") || !strings.Contains(output, "Completed GET /test-path 200") {
		t.Errorf("log output did not contain expected messages: %s", output)
	}
}

func TestLoggingMiddleware_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := LoggingMiddleware(testHandler(http.StatusInternalServerError), logger)

	req := httptest.NewRequest(http.MethodPost, "/fail-path", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	output := buf.String()
	if !strings.Contains(output, "Started POST /fail-path") || !strings.Contains(output, "Completed POST /fail-path 500") {
		t.Errorf("log output did not contain expected messages: %s", output)
	}
}
