// Package greet is the lesson 10 warm-up reference implementation.
package greet

import "fmt"

// Greeter is a one-method interface.
type Greeter interface {
	Greet(name string) string
}

// English greets in English.
type English struct{}

// Greet returns "Hello, <name>!".
func (English) Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Finnish greets in Finnish.
type Finnish struct{}

// Greet returns "Hei, <name>!".
func (Finnish) Greet(name string) string {
	return fmt.Sprintf("Hei, %s!", name)
}
