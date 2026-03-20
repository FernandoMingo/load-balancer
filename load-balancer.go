package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	backends []*Backend
	current  uint64
}

var serverPool ServerPool

// Atomically selects the next index in the ServerPool. Does NOT skip dead backends
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)%uint64(len(s.backends))))
}

// Returns a next peer to get an active connection
func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()
	l := len(s.backends) + next // sart from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(s.backends) // take an index by modding backends length
		// Pick an alive Backend. If it's not the original, store it.
		if s.backends[idx].isAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(i))
			}
			return s.backends[idx]
		}
	}
	return nil
}

func lb(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "service unavailable", http.StatusServiceUnavailable)
}

// checks if a Backend is alive by establishing a TCP connection
func isBackendAlive(url *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", url.Host, timeout)
	if err != nil {
		log.Println("Site unreachable error: ", err)
		return false
	}
	_ = conn.Close()
	return true
}

// AddBackend to the server pool
func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.setAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

func healthCheck() {
	t := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-t.C:
			log.Println("Starting Health check")
			serverPool.HealthCheck()
			log.Println("health check finished")
		}
	}
}
