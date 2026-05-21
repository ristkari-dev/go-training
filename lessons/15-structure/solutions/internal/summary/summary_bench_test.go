package summary

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
)

// BenchmarkTotalsByCategory measures TotalsByCategory throughput on a
// realistic 10,000-entry slice with a few dozen distinct categories.
//
// Run with:
//
//	go test -bench=BenchmarkTotalsByCategory -benchmem \
//	  ./lessons/15-structure/solutions/internal/summary/
//
// Expected output looks like (numbers will vary by machine):
//
//	BenchmarkTotalsByCategory-8     5000    234567 ns/op    8192 B/op    1 allocs/op
//
// The benchmark is informational — no assertions. It's there so
// students can see ns/op and B/op for a realistic workload.
func BenchmarkTotalsByCategory(b *testing.B) {
	// Setup: 10,000 random expenses with a fixed seed for reproducibility.
	r := rand.New(rand.NewSource(42))
	categories := []string{"coffee", "lunch", "groceries", "transit", "entertainment", "books", "household"}
	es := make([]expense.Expense, 10000)
	for i := range es {
		es[i] = expense.Expense{
			Date:     fmt.Sprintf("2026-%02d-%02d", 1+r.Intn(12), 1+r.Intn(28)),
			Amount:   r.Float64() * 100,
			Category: categories[r.Intn(len(categories))],
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = TotalsByCategory(es)
	}
}
