package middleware

import (
	"log"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs incoming requests
func LoggingMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: 200}

		logger.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(recorder, r)

		logger.Printf("Completed %s %s %d in %v", r.Method, r.URL.Path, recorder.status, time.Since(start))
	})
}
