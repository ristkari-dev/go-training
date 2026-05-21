package summary

import (
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/internal/expense"
)

// BenchmarkTotalsByCategory is a SKELETON. The full benchmark exists
// in the solutions tree (10,000 random expenses, fixed seed, b.N loop).
// Here we keep the body trivial so the file compiles.
//
// To run benchmarks:
//
//	go test -bench=. ./lessons/15-structure/solutions/internal/summary/
//
// (Run from the solutions tree where the implementation is real.)
func BenchmarkTotalsByCategory(b *testing.B) {
	// TODO: generate 10,000 random expenses (rand with fixed seed);
	// b.ResetTimer(); b.ReportAllocs(); loop calling TotalsByCategory.
	_ = expense.Expense{}
	b.Skip("benchmark not implemented in exercises tree; see solutions")
}
