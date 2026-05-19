package expense

import "testing"

// TestExpenseFormat is a SKELETON. Same shape as lesson 06/07/08 — fill in
// cases for small/medium/large amounts and assert against the expected
// "%s  €%-7.2f %s" output strings.
func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		// TODO: at least 3 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := tc.e.Format(); compare; t.Errorf if differ.
			_ = tc
		})
	}
}

// TestExpenseIsHigh is a SKELETON. Cover boundary 50 explicitly.
func TestExpenseIsHigh(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   bool
	}{
		// TODO: at least 4 cases incl. 49.99, 50 (false), 50.01 (true).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO
			_ = tc
		})
	}
}

// TestExpenseApplyDiscount is a SKELETON. Test that the discount STICKS
// on the caller's Expense — that's the lesson's main point (pointer
// receiver mutates).
//
// Cases to cover: 10% off a round number, 50% off, 0 rate (no-op),
// 1.0 rate (full discount → 0), 100% discount on zero-amount expense.
func TestExpenseApplyDiscount(t *testing.T) {
	cases := []struct {
		name        string
		startAmount float64
		rate        float64
		wantAmount  float64
	}{
		// TODO: at least 4 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   e := Expense{Amount: tc.startAmount}
			//   e.ApplyDiscount(tc.rate)
			//   if e.Amount != tc.wantAmount { t.Errorf(...) }
			_ = tc
		})
	}
}

// TestExpenseBump is a SKELETON. Same shape — verify the bump sticks.
//
// Cases: positive bump, negative bump (refund/correction), zero bump
// (no-op), bump on zero-amount expense.
func TestExpenseBump(t *testing.T) {
	cases := []struct {
		name        string
		startAmount float64
		bump        float64
		wantAmount  float64
	}{
		// TODO: at least 4 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   e := Expense{Amount: tc.startAmount}
			//   e.Bump(tc.bump)
			//   if e.Amount != tc.wantAmount { t.Errorf(...) }
			_ = tc
		})
	}
}
