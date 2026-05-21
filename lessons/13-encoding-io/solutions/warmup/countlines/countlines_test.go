package countlines

import (
	"io"
	"strings"
	"testing"
)

func TestCountLines(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"single-line-no-newline", "foo", 1},
		{"single-line-trailing-newline", "foo\n", 1},
		{"two-lines-no-trailing", "foo\nbar", 2},
		{"two-lines-trailing", "foo\nbar\n", 2},
		{"many-lines", "a\nb\nc\nd\ne\n", 5},
		{"blank-lines-counted", "\n\n\n", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CountLines(strings.NewReader(tc.in))
			if err != nil {
				t.Fatalf("CountLines(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("CountLines(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestCountLinesError(t *testing.T) {
	// Read two clean lines first, then a reader that errors.
	r := io.MultiReader(strings.NewReader("a\nb\n"), errReader{})
	n, err := CountLines(r)
	if err == nil {
		t.Fatal("expected error from failing reader, got nil")
	}
	if n != 2 {
		t.Errorf("expected partial count 2 from clean prefix, got %d", n)
	}
}

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
