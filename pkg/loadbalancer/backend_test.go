package loadbalancer

import "testing"

func TestBackendAliveState(t *testing.T) {
	backend := &Backend{}

	if backend.IsAlive() {
		t.Fatal("expected backend to be dead by default")
	}

	backend.SetAlive(true)
	if !backend.IsAlive() {
		t.Fatal("expected backend to be alive after SetAlive(true)")
	}

	backend.SetAlive(false)
	if backend.IsAlive() {
		t.Fatal("expected backend to be dead after SetAlive(false)")
	}
}
