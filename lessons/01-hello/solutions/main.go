// Package solutions is the reference implementation for lesson 01: Hello, Go.
package solutions

import "fmt"

// Greet returns a greeting like "Hello, World!".
func Greet(name string) string {
	return "Hello, " + name + "!"
}

// FormatExpense returns a formatted line for one expense.
func FormatExpense(date string, amount float64, category string) string {
	return fmt.Sprintf("%s  €%.2f  %s", date, amount, category)
}
