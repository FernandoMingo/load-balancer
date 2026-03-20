package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	requestsPerSecond = 1
	burst             = 2
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients         = make(map[string]*client)
	mu              sync.Mutex
	cleanUpInterval = time.Minute * 5
)

// gets a client and adds a rate limiter to it
func getClient(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	c, exists := clients[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
		clients[ip] = &client{limiter, time.Now()}
		return limiter
	}

	c.lastSeen = time.Now()
	return c.limiter
}

// limits requests and forces a retry
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Invalid remote address", http.StatusBadRequest)
			return
		}
		limiter := getClient(ip)

		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
