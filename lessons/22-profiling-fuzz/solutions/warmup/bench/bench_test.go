package bench

import (
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/solutions/internal/logparse"
)

func TestGenLines(t *testing.T) {
	for _, n := range []int{0, 1, 10, 100} {
		got := strings.Count(GenLines(n), "\n")
		if got != n {
			t.Errorf("GenLines(%d) has %d newlines, want %d", n, got, n)
		}
	}
}

func TestGenLinesParseable(t *testing.T) {
	entries, err := logparse.Parse(strings.NewReader(GenLines(50)))
	if err != nil {
		t.Fatalf("Parse(GenLines(50)) error: %v", err)
	}
	if len(entries) != 50 {
		t.Errorf("parsed %d entries, want 50", len(entries))
	}
}

// BenchmarkParse measures baseline logparse.Parse throughput.
// Run with: go test -bench=. -benchmem ./.../warmup/bench/
func BenchmarkParse(b *testing.B) {
	data := GenLines(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = logparse.Parse(strings.NewReader(data))
	}
}
