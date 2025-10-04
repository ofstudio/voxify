package randtoken

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewLength(t *testing.T) {
	for _, n := range []int{0, 1, 5, 16, 64} {
		tok := New(n)
		if len(tok) != n {
			t.Errorf("New(%d) returned token of length %d, want %d", n, len(tok), n)
		}
	}
}

func TestNewDiverse(t *testing.T) {
	// Generate many tokens and check that not all are identical
	count := 1000
	m := make(map[string]struct{})
	for i := 0; i < count; i++ {
		tok := New(16)
		m[tok] = struct{}{}
	}
	if len(m) < count/2 {
		t.Errorf("too many collisions: got %d unique tokens out of %d", len(m), count)
	}
}

func TestNewParallel(t *testing.T) {
	const goroutines = 50
	const perG = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make(chan error, goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				tok := New(32)
				if len(tok) != 32 {
					errs <- fmt.Errorf("got len %d, want 32", len(tok))
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}
