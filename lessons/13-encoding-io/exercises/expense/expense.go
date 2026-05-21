// Package expense holds the Expense type for the lesson 13 capstone.
//
// Carried forward verbatim from lesson 11 (same 4 methods: Format,
// IsHigh, ApplyDiscount, Bump; JSON tags intact). No test in this
// lesson — already tested in L09.
package expense

import "fmt"

// Expense is one row of the expense tracker.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

func (e *Expense) ApplyDiscount(rate float64) {
	e.Amount *= (1 - rate)
}

func (e *Expense) Bump(amount float64) {
	e.Amount += amount
}
