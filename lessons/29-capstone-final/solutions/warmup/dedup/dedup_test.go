package dedup

import (
	"fmt"
	"sync"
	"testing"
)

func TestSeenFirstThenRepeat(t *testing.T) {
	d := New(10)
	if d.Seen("a") {
		t.Error("first sight should be false")
	}
	if !d.Seen("a") {
		t.Error("second sight should be true")
	}
}

func TestBoundedEviction(t *testing.T) {
	d := New(2)
	d.Seen("a")
	d.Seen("b")
	d.Seen("c") // over cap 2 → evicts oldest "a"; holds {b, c}
	// "a" was evicted, so it reads as new. (This probe re-adds "a" and
	// evicts "b" — Seen mutates, so probe only the one key under test.)
	if d.Seen("a") {
		t.Error("a should have been evicted (read as new)")
	}
}

func TestConcurrent(t *testing.T) {
	d := New(1000)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(n int) { defer wg.Done(); d.Seen(fmt.Sprintf("k%d", n)) }(i)
	}
	wg.Wait()
}
