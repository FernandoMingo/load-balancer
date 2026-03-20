package loadbalancer

import "testing"

func TestGetNextPeerSkipsDeadBackends(t *testing.T) {
	b1 := &Backend{Alive: false}
	b2 := &Backend{Alive: true}
	b3 := &Backend{Alive: false}
	pool := NewServerPool([]*Backend{b1, b2, b3})

	peer := pool.GetNextPeer()
	if peer == nil {
		t.Fatal("expected an alive backend")
	}
	if peer != b2 {
		t.Fatal("expected second backend to be selected")
	}
}

func TestGetNextPeerReturnsNilWhenAllDead(t *testing.T) {
	pool := NewServerPool([]*Backend{
		{Alive: false},
		{Alive: false},
	})

	peer := pool.GetNextPeer()
	if peer != nil {
		t.Fatal("expected nil when all backends are dead")
	}
}
