package loadbalancer

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const defaultHealthCheckInterval = 20 * time.Second

const (
	Attempts int = iota
	Retry
)

type LoadBalancer struct {
	Pool                *ServerPool
	healthCheckInterval time.Duration
}

// GetRetryFromContext returns the retries for request
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

// GetAttemptsFromContext returns the attempts for request
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func New(rawBackendURLs []string, healthCheckInterval time.Duration) (*LoadBalancer, error) {
	var lb *LoadBalancer
	backends := make([]*Backend, 0, len(rawBackendURLs))
	for _, rawURL := range rawBackendURLs {
		trimmed := strings.TrimSpace(rawURL)
		if trimmed == "" {
			continue
		}

		serverURL, err := url.Parse(trimmed)
		if err != nil {
			return nil, err
		}

		backend := &Backend{
			URL:   serverURL,
			Alive: true,
		}

		proxy := httputil.NewSingleHostReverseProxy(serverURL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[%s] %s\n", serverURL.Host, err.Error())
			retries := GetRetryFromContext(r)
			if retries < 3 {
				time.Sleep(10 * time.Millisecond)
				ctx := context.WithValue(r.Context(), Retry, retries+1)
				proxy.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Mark backend as down
			backend.SetAlive(false)

			// Retry the same request on the next available backend.
			attempts := GetAttemptsFromContext(r)
			if attempts >= len(backends) {
				http.Error(w, "service unavailable", http.StatusServiceUnavailable)
				return
			}

			log.Printf("%s(%s) attempting failover %d\n", serverURL.Host, r.URL.Path, attempts)
			ctx := context.WithValue(r.Context(), Attempts, attempts+1)
			lb.ServeHTTP(w, r.WithContext(ctx))
		}
		backend.ReverseProxy = proxy

		backends = append(backends, backend)
	}

	if len(backends) == 0 {
		return nil, errors.New("please provide one or more backends to load balance")
	}

	if healthCheckInterval <= 0 {
		healthCheckInterval = defaultHealthCheckInterval
	}

	lb = &LoadBalancer{
		Pool:                NewServerPool(backends),
		healthCheckInterval: healthCheckInterval,
	}

	return lb, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := lb.Pool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "service unavailable", http.StatusServiceUnavailable)
}

func IsBackendAlive(targetURL *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", targetURL.Host, timeout)
	if err != nil {
		log.Println("Site unreachable error:", err)
		return false
	}

	_ = conn.Close()
	return true
}

func (lb *LoadBalancer) HealthCheck() {
	for _, backend := range lb.Pool.Backends() {
		status := "up"
		alive := IsBackendAlive(backend.URL)
		backend.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", backend.URL, status)
	}
}

func (lb *LoadBalancer) StartHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(lb.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("Starting health check")
			lb.HealthCheck()
			log.Println("Health check finished")
		}
	}
}
