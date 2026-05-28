package bench

import (
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/exercises/internal/logparse"
)

// TestGenLines is a SKELETON. Verify GenLines(n) produces n parseable
// lines.
func TestGenLines(t *testing.T) {
	cases := []int{
		// TODO: e.g. 0, 1, 10, 100
	}
	for _, n := range cases {
		got := strings.Count(GenLines(n), "\n")
		if got != n {
			t.Errorf("GenLines(%d) has %d newlines, want %d", n, got, n)
		}
	}
}

// BenchmarkParse measures the baseline logparse.Parse throughput over
// a fixed synthetic input. Run with: go test -bench=. -benchmem
func BenchmarkParse(b *testing.B) {
	// TODO:
	//   data := GenLines(1000)
	//   b.ReportAllocs()
	//   b.ResetTimer()
	//   for i := 0; i < b.N; i++ {
	//       _, _ = logparse.Parse(strings.NewReader(data))
	//   }
	_ = logparse.Parse
}
