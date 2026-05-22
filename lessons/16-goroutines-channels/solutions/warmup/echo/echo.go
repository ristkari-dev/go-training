// Package echo is the lesson 16 warm-up reference implementation.
package echo

// Echo returns a new channel that receives everything sent on in.
func Echo(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for v := range in {
			out <- v
		}
		close(out)
	}()
	return out
}
