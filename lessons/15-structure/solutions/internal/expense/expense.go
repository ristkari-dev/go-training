// Package expense holds the Expense type for the lesson 15 capstone.
//
// Carried forward verbatim from lesson 10 (4 methods: Format, IsHigh,
// ApplyDiscount, Bump; JSON tags intact). The store, csvimport, summary,
// and cmd/expenses packages in lesson 15 all import this.
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
