// Package solutions is the reference implementation for lesson 04: Functions & first tests.
package solutions

import "fmt"

// Categorise returns "snack" for amounts under €10, "regular" for €10..€50
// (inclusive on both ends), "splurge" above €50.
func Categorise(amount float64) string {
	switch {
	case amount < 10:
		return "snack"
	case amount <= 50:
		return "regular"
	default:
		return "splurge"
	}
}

// FormatExpense returns "YYYY-MM-DD  €AMOUNT  category" with the amount
// left-padded to 7 characters using the "%-7.2f" verb.
func FormatExpense(date string, amount float64, cat string) string {
	return fmt.Sprintf("%s  €%-7.2f %s", date, amount, cat)
}
