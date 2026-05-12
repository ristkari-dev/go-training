// Package solutions is the reference implementation for lesson 06: Structs and methods.
package solutions

import "fmt"

// Expense is one row of the expense tracker.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using "%s  €%-7.2f %s".
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

// IsHigh reports whether the expense is over €50.
func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

// TotalsByCategory returns a map of category → summed amount.
// Empty/nil input returns an empty (non-nil) map.
func TotalsByCategory(es []Expense) map[string]float64 {
	totals := map[string]float64{}
	for _, e := range es {
		totals[e.Category] += e.Amount
	}
	return totals
}
