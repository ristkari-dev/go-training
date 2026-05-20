// Package expense holds the Expense type for the lesson 10 capstone.
//
// Carried forward verbatim from lesson 09 (4 methods: Format, IsHigh,
// ApplyDiscount, Bump; JSON tags intact). No test in this lesson —
// already tested in L09. Lesson 10's store and cmd/expenses packages
// import this.
package expense

import "fmt"

// Expense is one row of the expense tracker.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" with "%s  €%-7.2f %s".
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

// IsHigh reports whether the expense is over €50.
func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

// ApplyDiscount reduces e.Amount by the given rate. Pointer receiver.
func (e *Expense) ApplyDiscount(rate float64) {
	e.Amount *= (1 - rate)
}

// Bump adds amount to e.Amount. Pointer receiver.
func (e *Expense) Bump(amount float64) {
	e.Amount += amount
}
