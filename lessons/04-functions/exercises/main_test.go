package exercises

import "testing"

// TestCategorise is a SKELETON. The struct shape and the for-loop scaffold
// are here for you; your job is to fill in the cases slice (think about
// the boundaries: <10, exactly 10, between, exactly 50, >50) and the body
// of the t.Run block.
//
// When you're done, this test should fail (because Categorise panics with
// a TODO) — exactly like the warm-up tests. Once your Categorise
// implementation in main.go is correct, all sub-tests pass.
func TestCategorise(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   string
	}{
		// TODO: add at least 5 cases here. Cover the snack/regular/splurge
		// branches AND the boundary values (9.99, 10.00, 50.00, 50.01).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call Categorise(tc.amount), compare with tc.want,
			// use t.Errorf with a clear message if they differ.
			_ = tc
		})
	}
}

// TestFormatExpense is a SKELETON. Same shape as TestCategorise — fill in
// the cases and the t.Run body.
//
// Look at the doc comment on FormatExpense for the expected output strings
// (note the spacing: two spaces between fields, the €-prefixed amount
// padded to 7 chars, etc.). The "Hint" line in the doc comment gives away
// the exact fmt.Sprintf format string; building those expected outputs by
// hand is part of the exercise.
func TestFormatExpense(t *testing.T) {
	cases := []struct {
		name   string
		date   string
		amount float64
		cat    string
		want   string
	}{
		// TODO: add at least 3 cases here. Use the examples in the
		// FormatExpense doc comment as a starting point.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call FormatExpense, compare with tc.want, t.Errorf if differ.
			_ = tc
		})
	}
}
