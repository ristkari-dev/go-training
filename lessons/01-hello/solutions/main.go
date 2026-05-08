// Package solutions is the reference implementation for lesson 01: Hello, Go.
package solutions

import "fmt"

// Greet returns a greeting like "Hello, World!".
func Greet(name string) string {
	return "Hello, " + name + "!"
}

// FormatExpense returns a formatted line for one expense.
//
// The format is "date  €amount  category" where amount is printed
// with exactly two decimal places.
//
// Examples:
//
//	FormatExpense("2026-05-07", 4.50, "coffee")     → "2026-05-07  €4.50  coffee"
//	FormatExpense("2026-05-07", 23.5, "groceries")  → "2026-05-07  €23.50  groceries"
//	FormatExpense("2026-04-30", 1234.5, "rent")     → "2026-04-30  €1234.50  rent"
func FormatExpense(date string, amount float64, category string) string {
	return fmt.Sprintf("%s  €%.2f  %s", date, amount, category)
}
