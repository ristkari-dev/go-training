// Package exercises is the starter code for lesson 01: Hello, Go.
//
// This file holds the WARM-UP exercise. Two small functions to build
// muscle memory for "a Go function that returns a string." Make the
// failing tests in warmup_test.go pass.
package exercises

// WarmupHello returns the greeting "Hello, Go!".
func WarmupHello() string {
	return("TODO: return the string \"Hello, Go!\"")
}

// WarmupGreet returns "Hello, <name>!" for the given name.
//
// Examples:
//
//	WarmupGreet("World") → "Hello, World!"
//	WarmupGreet("Aki")   → "Hello, Aki!"
//
// Hint: Go strings concatenate with the + operator.
func WarmupGreet(name string) string {
	return("TODO: return \"Hello, \" + name + \"!\"")
}
