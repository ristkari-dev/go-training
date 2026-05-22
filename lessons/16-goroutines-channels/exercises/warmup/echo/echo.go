// Package echo is the lesson 16 warm-up: a tiny demonstration of
// goroutines + channels working together.
//
// Echo takes an input channel of strings and returns a new output
// channel of strings. The output gets every value the input received,
// in order. When the input is closed (no more values), the output is
// closed too.
//
// The pedagogical point: Echo returns the output channel IMMEDIATELY.
// The reading-and-writing happens inside a spawned goroutine. This is
// the canonical "function that does work concurrently" shape in Go —
// not a Future, not a Promise, just a channel you can read from while
// the work happens in the background.
package echo

// Echo returns a new channel that receives everything sent on in,
// in order. When in is closed, the returned channel is closed too.
//
// Hint:
//  1. out := make(chan string)
//  2. go func() {
//     for v := range in {
//     out <- v
//     }
//     close(out)
//     }()
//  3. return out
//
// Three things to internalise:
//   - out is created BEFORE the goroutine starts.
//   - The goroutine closes out when in closes.
//   - Echo returns immediately; the work happens later.
func Echo(in <-chan string) <-chan string {
	panic("TODO: make output channel; spawn goroutine that range-copies and closes; return output")
}
