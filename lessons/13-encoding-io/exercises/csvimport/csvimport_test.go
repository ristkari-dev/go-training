package csvimport

import (
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/exercises/expense"
)

// TestParseGoldenPath is a SKELETON. Cover at least:
//   - Empty input → empty slice + nil
//   - One row
//   - Multiple rows
//   - Blank lines interspersed and skipped
//   - Leading/trailing whitespace tolerated
func TestParseGoldenPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []expense.Expense
	}{
		// TODO: 4+ cases.
		// {"empty", "", []expense.Expense{}},
		// {"single", "2026-05-21,4.50,coffee", []expense.Expense{{Date:"2026-05-21", Amount:4.50, Category:"coffee"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got, err := Parse(strings.NewReader(tc.in))
			//   if err != nil → t.Fatalf("unexpected error: %v", err)
			//   if !reflect.DeepEqual(got, tc.want) → t.Errorf("...")
			_ = tc
			_ = strings.NewReader
		})
	}
}

// TestParseMalformed is a SKELETON. Cover at least:
//   - Wrong field count (1 or 2 commas) → non-nil error
//   - Amount that doesn't parse → non-nil error mentioning the line number
func TestParseMalformed(t *testing.T) {
	// TODO:
	//   _, err := Parse(strings.NewReader("a,b"))  // missing third field
	//   if err == nil → t.Fatal("expected error for wrong field count")
	//   _, err = Parse(strings.NewReader("2026-05-21,oops,coffee"))
	//   if err == nil → t.Fatal("expected error for bad amount")
	//   if !strings.Contains(err.Error(), "line 1") → t.Errorf("error should mention line number, got %v", err)
}
