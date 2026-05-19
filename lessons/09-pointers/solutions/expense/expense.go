// Package expense is the lesson 09 main exercise reference implementation.
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

// ApplyDiscount reduces e.Amount by the given rate (0.0 to 1.0).
// Pointer receiver — mutates the caller.
func (e *Expense) ApplyDiscount(rate float64) {
	e.Amount *= (1 - rate)
}

// Bump adds amount to e.Amount. Pointer receiver — mutates the caller.
func (e *Expense) Bump(amount float64) {
	e.Amount += amount
}
