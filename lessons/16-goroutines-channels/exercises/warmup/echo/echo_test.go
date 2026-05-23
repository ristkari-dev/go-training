package echo

import (
	"testing"
)

// TestEcho is a SKELETON. Cover at least:
//   - Sending 3 values produces 3 outputs in order
//   - Closing input → closing output (the test should range over the
//     output and finish without deadlocking)
//   - Empty input (closed immediately) → empty output (closed immediately)
func TestEcho(t *testing.T) {
	cases := []struct {
		name string
		send []string
		want []string
	}{
		// TODO: at least 3 cases.
		// {"three", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		// {"empty", []string{}, []string{}},
		// {"one", []string{"only"}, []string{"only"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   in := make(chan string)
			//   out := Echo(in)
			//   go func() {
			//       for _, v := range tc.send { in <- v }
			//       close(in)
			//   }()
			//   var got []string
			//   for v := range out { got = append(got, v) }
			//   compare got to tc.want
			_ = tc
		})
	}
}
