package store

import (
	"path/filepath"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/10-interfaces/exercises/expense"
)

// runStoreContract is a helper that verifies the Store CONTRACT — i.e.,
// behaviours every implementation must support. Both TestJSONStore and
// TestMemoryStore call it.
//
// SKELETON. Fill in the body to:
//  1. Load() on an empty store returns ([]expense.Expense{}, nil) — non-nil
//     empty slice, nil error.
//  2. After Save([e1, e2, e3]), Load() returns those three in order.
//  3. After a second Save([e4]), Load() returns just [e4] (Save REPLACES).
//
// Hint: define a fixture once at the top of the function:
//
//	e1 := expense.Expense{Date: "2026-05-20", Amount: 4.50, Category: "coffee"}
//	...
func runStoreContract(t *testing.T, s Store) {
	t.Helper()
	// TODO: implement the three contract checks described above.
	_ = s
	_ = expense.Expense{}
}

// TestJSONStore is a SKELETON. Construct a JSONStore at a temp path, then
// hand it to runStoreContract. The temp path uses t.TempDir() so the test
// is hermetic and self-cleaning.
func TestJSONStore(t *testing.T) {
	// TODO:
	//   path := filepath.Join(t.TempDir(), "expenses.json")
	//   runStoreContract(t, NewJSONStore(path))
	_ = filepath.Join
	_ = t
}

// TestMemoryStore is a SKELETON. Construct a fresh MemoryStore and hand
// it to runStoreContract. No temp dir needed — memory is process-local.
func TestMemoryStore(t *testing.T) {
	// TODO: runStoreContract(t, NewMemoryStore())
	_ = t
}
