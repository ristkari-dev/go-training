// Package greet provides a simple greeting helper for lesson 07's
// "split code into a subpackage" warm-up.
//
// We're putting this function in its own package — separate from main —
// to demonstrate Go's subpackage mechanics. The main.go in the parent
// directory imports this with:
//
//	import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
//
// and calls greet.Greet("Aki").
package greet

// Greet returns "Hello, <name>!".
//
// The function name is exported (capital G) so callers from other packages
// can use it. If we wrote `func greet(name string)` (lower-case g), the
// import would compile but `greet.greet("Aki")` would not — the symbol
// would be invisible from outside this package.
//
// Examples:
//
//	Greet("Aki")    → "Hello, Aki!"
//	Greet("World")  → "Hello, World!"
//	Greet("")       → "Hello, !"
//
// Hint: fmt.Sprintf("Hello, %s!", name) is the one-line implementation.
func Greet(name string) string {
	panic("TODO: return the greeting per the doc comment")
}
