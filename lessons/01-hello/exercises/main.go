// Package exercises is the starter code for lesson 01: Hello, Go.
//
// This file holds the MAIN exercise. Greet is already implemented as a
// tiny worked example; FormatExpense is yours to write. Make the failing
// tests in main_test.go pass.
package exercises

import "fmt"

// Greet returns a greeting like "Hello, World!".
//
// Provided as a worked example of a function that takes a string,
// concatenates with + operators, and returns a new string.
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
//
// Hint: import "fmt" and use fmt.Sprintf with the format verb %.2f
// to print the amount with two decimal places.
func FormatExpense(date string, amount float64, category string) string {
	return (fmt.Sprintf("TODO: implement FormatExpense (see hint in the doc comment)"))
}
