// Package fanin is the lesson 20 warm-up reference implementation.
package fanin

import "sync"

// FanIn multiplexes N input channels into one. Output closes when
// all inputs close.
func FanIn[T any](chans ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	for _, ch := range chans {
		wg.Add(1)
		go func(input <-chan T) {
			defer wg.Done()
			for v := range input {
				out <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
