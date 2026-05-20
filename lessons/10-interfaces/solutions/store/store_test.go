package store

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/10-interfaces/solutions/expense"
)

// runStoreContract verifies behaviours every Store implementation must support.
func runStoreContract(t *testing.T, s Store) {
	t.Helper()

	// 1. Load on an empty store returns a non-nil empty slice and nil error.
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if got == nil {
		t.Errorf("Load empty: got nil slice, want non-nil empty")
	}
	if len(got) != 0 {
		t.Errorf("Load empty: got %d items, want 0", len(got))
	}

	// 2. Save then Load returns the same expenses in order.
	want := []expense.Expense{
		{Date: "2026-05-20", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-20", Amount: 12, Category: "lunch"},
		{Date: "2026-05-20", Amount: 75, Category: "rent"},
	}
	if err := s.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err = s.Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load after Save: got %v, want %v", got, want)
	}

	// 3. A second Save REPLACES (doesn't merge).
	replacement := []expense.Expense{
		{Date: "2026-05-20", Amount: 1.50, Category: "snack"},
	}
	if err := s.Save(replacement); err != nil {
		t.Fatalf("Save replacement: %v", err)
	}
	got, err = s.Load()
	if err != nil {
		t.Fatalf("Load after replacement: %v", err)
	}
	if !reflect.DeepEqual(got, replacement) {
		t.Errorf("Load after replacement: got %v, want %v", got, replacement)
	}
}

func TestJSONStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "expenses.json")
	runStoreContract(t, NewJSONStore(path))
}

func TestMemoryStore(t *testing.T) {
	runStoreContract(t, NewMemoryStore())
}
