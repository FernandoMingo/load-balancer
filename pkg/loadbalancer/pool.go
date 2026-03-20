package loadbalancer

import "sync/atomic"

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func NewServerPool(backends []*Backend) *ServerPool {
	return &ServerPool{backends: backends}
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) Backends() []*Backend {
	return s.backends
}

// NextIndex atomically selects the next backend index.
func (s *ServerPool) NextIndex() int {
	if len(s.backends) == 0 {
		return -1
	}

	return int(atomic.AddUint64(&s.current, 1) % uint64(len(s.backends)))
}

// GetNextPeer returns the next alive backend.
func (s *ServerPool) GetNextPeer() *Backend {
	if len(s.backends) == 0 {
		return nil
	}

	next := s.NextIndex()
	for i := 0; i < len(s.backends); i++ {
		idx := (next + i) % len(s.backends)
		if s.backends[idx].IsAlive() {
			atomic.StoreUint64(&s.current, uint64(idx))
			return s.backends[idx]
		}
	}

	return nil
}
