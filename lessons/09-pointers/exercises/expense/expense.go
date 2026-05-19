// Package expense is the lesson 09 main exercise — adds pointer-receiver
// mutation methods to the Expense type from lesson 08.
//
// Existing value-receiver methods (Format, IsHigh) stay — they read
// e but don't change it. The two new methods (ApplyDiscount, Bump) use
// pointer receivers because they mutate e.Amount.
package expense

// Expense is one row of the expense tracker — date, amount, category.
//
// Same fields and JSON tags as lesson 08. Mutation methods (ApplyDiscount,
// Bump) operate on a pointer to the value; read-only methods (Format,
// IsHigh) take a value.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" via fmt.Sprintf with
// "%s  €%-7.2f %s". Same byte-for-byte output as lesson 04/06/07/08.
//
// Value receiver — reads e, doesn't mutate.
//
// Hint: same as lesson 08.
func (e Expense) Format() string {
	panic("TODO: return fmt.Sprintf(\"%s  €%-7.2f %s\", e.Date, e.Amount, e.Category)")
}

// IsHigh reports whether the expense is over €50. Value receiver.
//
// Boundary: amount of exactly 50 is NOT high; 50.01 IS high.
func (e Expense) IsHigh() bool {
	panic("TODO: return e.Amount > 50")
}

// ApplyDiscount reduces e.Amount by the given rate (0.0 to 1.0).
//
// Pointer receiver — the call mutates the caller's Expense. After
// `e.ApplyDiscount(0.10)`, e.Amount is 90% of what it was.
//
// rate < 0 or rate > 1 is undefined behaviour for this exercise (no
// validation; trust the caller). Lesson 11 covers error handling for
// these "trust boundary" cases properly.
//
// Examples:
//
//	e := Expense{Amount: 100}
//	e.ApplyDiscount(0.10)
//	e.Amount  → 90.0
//
//	e := Expense{Amount: 50}
//	e.ApplyDiscount(0)
//	e.Amount  → 50.0      (no change for zero rate)
//
//	e := Expense{Amount: 100}
//	e.ApplyDiscount(1)
//	e.Amount  → 0.0       (100% discount → free)
//
// Hint: e.Amount *= (1 - rate)
func (e *Expense) ApplyDiscount(rate float64) {
	panic("TODO: multiply e.Amount by (1 - rate)")
}

// Bump adds amount to e.Amount.
//
// Pointer receiver — mutates the caller. `amount` can be positive
// (increase) or negative (refund / correction). Use Bump for adjustments
// that aren't a proportional discount.
//
// Examples:
//
//	e := Expense{Amount: 10}
//	e.Bump(5)
//	e.Amount  → 15.0
//
//	e := Expense{Amount: 10}
//	e.Bump(-3)
//	e.Amount  → 7.0
//
//	e := Expense{Amount: 10}
//	e.Bump(0)
//	e.Amount  → 10.0
//
// Hint: e.Amount += amount
func (e *Expense) Bump(amount float64) {
	panic("TODO: add amount to e.Amount")
}
