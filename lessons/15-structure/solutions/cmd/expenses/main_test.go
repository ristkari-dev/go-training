package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/store"
	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/testutil"
)

// TestSummaryGolden runs the summary subcommand against a fixed dataset,
// captures stdout into a bytes.Buffer, and compares to the golden file
// at testdata/summary.golden.
//
// To regenerate the golden file after intentional output changes:
//
//	go test -update ./lessons/15-structure/solutions/cmd/expenses/
//
// The test uses an in-memory MemoryStore to avoid filesystem I/O for the
// test fixture. The CLI itself (run via os.Args) uses JSONStore; this
// test bypasses parseStoreOption / buildStore to seed the store directly.
func TestSummaryGolden(t *testing.T) {
	s := store.NewMemoryStore()
	if err := s.Save([]expense.Expense{
		{Date: "2026-05-21", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-21", Amount: 12.00, Category: "lunch"},
		{Date: "2026-05-22", Amount: 80.00, Category: "groceries"},
		{Date: "2026-05-22", Amount: 3.50, Category: "coffee"},
	}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	var buf bytes.Buffer
	if err := cmdSummary(s, &buf); err != nil {
		t.Fatalf("cmdSummary: %v", err)
	}

	testutil.AssertGolden(t, buf.Bytes(), filepath.Join("testdata", "summary.golden"))
}
