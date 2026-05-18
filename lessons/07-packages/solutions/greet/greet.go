// Package greet is the reference implementation of lesson 07's warm-up.
package greet

import "fmt"

// Greet returns "Hello, <name>!".
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
