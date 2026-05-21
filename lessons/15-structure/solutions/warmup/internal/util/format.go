// Package util is the lesson 15 warm-up reference implementation.
package util

import "fmt"

// Money formats a euro amount as "€<2-decimal>".
func Money(amount float64) string {
	return fmt.Sprintf("€%.2f", amount)
}
