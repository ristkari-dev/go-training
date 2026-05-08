// Package solutions is the reference implementation for lesson 01: Hello, Go.
//
// This file holds the warm-up reference solution.
package solutions

// WarmupHello returns the greeting "Hello, Go!".
func WarmupHello() string {
	return "Hello, Go!"
}

// WarmupGreet returns "Hello, <name>!" for the given name.
func WarmupGreet(name string) string {
	return "Hello, " + name + "!"
}
