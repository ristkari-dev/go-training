package counter

import (
	"sync"
	"testing"
)

// TestCounter spawns 100 goroutines, each doing 100 increments
// (10,000 total). Asserts the final value is exactly 10000. Passes
// cleanly under `go test -race` because Counter is mutex-protected.
func TestCounter(t *testing.T) {
	var c Counter
	var wg sync.WaitGroup

	const goroutines = 100
	const incPerG = 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incPerG; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	want := goroutines * incPerG
	if got := c.Value(); got != want {
		t.Errorf("Counter.Value() = %d, want %d", got, want)
	}
}
