package csvimport

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
)

func TestParseGoldenPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []expense.Expense
	}{
		{
			"empty",
			"",
			[]expense.Expense{},
		},
		{
			"single",
			"2026-05-21,4.50,coffee",
			[]expense.Expense{{Date: "2026-05-21", Amount: 4.50, Category: "coffee"}},
		},
		{
			"multiple",
			"2026-05-21,4.50,coffee\n2026-05-21,12.00,lunch",
			[]expense.Expense{
				{Date: "2026-05-21", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-21", Amount: 12.00, Category: "lunch"},
			},
		},
		{
			"blank-lines-skipped",
			"2026-05-21,4.50,coffee\n\n\n2026-05-21,12.00,lunch\n",
			[]expense.Expense{
				{Date: "2026-05-21", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-21", Amount: 12.00, Category: "lunch"},
			},
		},
		{
			"whitespace-tolerated",
			"  2026-05-21 , 4.50 , coffee  ",
			[]expense.Expense{{Date: "2026-05-21", Amount: 4.50, Category: "coffee"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tc.in))
			if err != nil {
				t.Fatalf("Parse: unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Parse = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestParseMalformed(t *testing.T) {
	t.Run("wrong-field-count", func(t *testing.T) {
		_, err := Parse(strings.NewReader("a,b"))
		if err == nil {
			t.Fatal("expected error for wrong field count, got nil")
		}
	})
	t.Run("bad-amount-mentions-line-number", func(t *testing.T) {
		_, err := Parse(strings.NewReader("2026-05-21,oops,coffee"))
		if err == nil {
			t.Fatal("expected error for bad amount, got nil")
		}
		if !strings.Contains(err.Error(), "line 1") {
			t.Errorf("error should mention line number, got %v", err)
		}
	})
	t.Run("bad-amount-line-2", func(t *testing.T) {
		input := "2026-05-21,4.50,coffee\n2026-05-21,oops,lunch"
		out, err := Parse(strings.NewReader(input))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "line 2") {
			t.Errorf("error should mention line 2, got %v", err)
		}
		if len(out) != 1 {
			t.Errorf("expected partial output of 1 entry from line 1, got %d", len(out))
		}
	})
}
