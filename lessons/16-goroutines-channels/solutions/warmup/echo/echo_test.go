package echo

import (
	"reflect"
	"testing"
)

func TestEcho(t *testing.T) {
	cases := []struct {
		name string
		send []string
		want []string
	}{
		{"three", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"one", []string{"only"}, []string{"only"}},
		{"empty", []string{}, []string{}},
		{"unicode", []string{"héllo", "wörld"}, []string{"héllo", "wörld"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := make(chan string)
			out := Echo(in)

			// Producer: send all values, then close.
			go func() {
				for _, v := range tc.send {
					in <- v
				}
				close(in)
			}()

			// Consumer: collect everything Echo emits.
			var got []string
			for v := range out {
				got = append(got, v)
			}

			// Normalise empty slice vs nil for comparison.
			if got == nil {
				got = []string{}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Echo(%v): got %v, want %v", tc.send, got, tc.want)
			}
		})
	}
}

// TestEchoClosesOutput asserts that if the producer never closes the
// input, Echo's goroutine is "leaked" (still running waiting for input)
// — and conversely, if we DO close the input, Echo closes its output.
// This documents the contract: Echo's lifetime is tied to its input.
func TestEchoClosesOutput(t *testing.T) {
	in := make(chan string)
	out := Echo(in)

	close(in)

	// out should close soon after in does. range over out completes
	// without blocking.
	for range out {
		t.Errorf("expected no values from out after closing in")
	}
}
