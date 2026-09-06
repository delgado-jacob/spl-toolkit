package main

import (
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func TestMapperRegistryAddGetRemove(t *testing.T) {
	registry := newMapperRegistry()
	want := mapper.New()

	id, ok := registry.add(want)
	if !ok || id <= 0 {
		t.Fatalf("add() = (%d, %v), want a positive handle", id, ok)
	}
	got, ok := registry.get(id)
	if !ok || got != want {
		t.Fatalf("get(%d) = (%p, %v), want (%p, true)", id, got, ok, want)
	}

	registry.remove(id)
	registry.remove(id)
	if _, ok := registry.get(id); ok {
		t.Fatalf("removed handle %d still resolves", id)
	}

	if _, err := got.MapQuery("search src_ip=192.0.2.1"); err != nil {
		t.Fatalf("mapper reference obtained before removal is no longer usable: %v", err)
	}
}

func TestMapperRegistryConcurrentUniqueHandles(t *testing.T) {
	registry := newMapperRegistry()
	const goroutines = 32
	const perGoroutine = 100

	ids := make(chan int, goroutines*perGoroutine)
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range perGoroutine {
				id, ok := registry.add(mapper.New())
				if !ok {
					t.Errorf("registry exhausted unexpectedly")
					return
				}
				ids <- id
				registry.remove(id)
			}
		}()
	}
	wg.Wait()
	close(ids)

	seen := make(map[int]struct{}, goroutines*perGoroutine)
	for id := range ids {
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("duplicate handle %d", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != goroutines*perGoroutine {
		t.Fatalf("created %d handles, want %d", len(seen), goroutines*perGoroutine)
	}
}

func TestMapperRegistryExhaustionDoesNotWrap(t *testing.T) {
	registry := newMapperRegistry()
	registry.nextID = maxMapperID

	id, ok := registry.add(mapper.New())
	if !ok || id != maxMapperID {
		t.Fatalf("last add() = (%d, %v), want (%d, true)", id, ok, maxMapperID)
	}
	if id, ok := registry.add(mapper.New()); ok || id != 0 {
		t.Fatalf("add after exhaustion = (%d, %v), want (0, false)", id, ok)
	}
	if registry.nextID <= maxMapperID {
		t.Fatalf("counter wrapped or failed to advance: %d", registry.nextID)
	}
}
