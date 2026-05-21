package countlines

import (
	"io"
	"strings"
	"testing"
)

// TestCountLines is a SKELETON. At least cover:
//   - Empty input → 0
//   - Single line, no trailing newline → 1
//   - Multiple lines with trailing newline → N
//   - Trailing newline doesn't add a phantom line
//
// Use strings.NewReader(s) to build an io.Reader from a string literal.
// This is THE idiom for testing functions that accept io.Reader.
func TestCountLines(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		// TODO: at least 4 cases.
		// {"empty", "", 0},
		// {"single-line", "foo", 1},
		// {"two-lines", "foo\nbar", 2},
		// {"trailing-newline", "foo\nbar\n", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got, err := CountLines(strings.NewReader(tc.in))
			//   if err != nil → t.Fatalf("unexpected error: %v", err)
			//   if got != tc.want → t.Errorf("...")
			_ = tc
			_ = strings.NewReader
		})
	}
}

// TestCountLinesError is a SKELETON. Demonstrate that a reader returning
// an error mid-stream surfaces it via the Err() check at the end of the
// scan loop. Use io.MultiReader to glue a good reader + an errReader.
func TestCountLinesError(t *testing.T) {
	// TODO:
	//   r := io.MultiReader(strings.NewReader("a\nb\n"), errReader{})
	//   _, err := CountLines(r)
	//   if err == nil → t.Fatal("expected error from failing reader")
	_ = io.MultiReader
}

// errReader is a helper for testing: every Read returns an error.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
