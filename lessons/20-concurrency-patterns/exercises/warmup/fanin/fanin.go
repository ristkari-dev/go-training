// Package fanin is the lesson 20 warm-up: multiplex N input channels
// into a single output channel.
//
// FanIn is the canonical "many producers, one consumer" multiplexer.
// Each input gets a forwarding goroutine that reads its channel and
// sends to the output. A closer goroutine watches a WaitGroup and
// closes the output once all forwarders are done.
//
// Generic via [T any] (lesson 12) so the same function works for any
// channel element type.
package fanin

import "sync"

// FanIn returns a new channel that receives every value sent on any
// of chans. The output is closed when ALL inputs are closed.
//
// Examples:
//
//	a := make(chan int); b := make(chan int)
//	out := FanIn(a, b)
//	// values sent on a or b appear on out
//	// close(a) AND close(b) → out closes too
//
// Hint:
//  1. out := make(chan T)
//  2. var wg sync.WaitGroup
//  3. for each input channel: wg.Add(1); spawn goroutine that:
//     for v := range input { out <- v }
//     wg.Done()
//  4. go func() { wg.Wait(); close(out) }()
//  5. return out
//
// The closer goroutine pattern (step 4) is the key insight: we can't
// close inline because we don't know when all forwarders finish. The
// goroutine waits for them all then closes once.
func FanIn[T any](chans ...<-chan T) <-chan T {
	_ = sync.WaitGroup{}
	panic("TODO: forward goroutines + closer goroutine; see hint")
}
