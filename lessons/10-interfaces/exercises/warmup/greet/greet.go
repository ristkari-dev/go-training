// Package greet is the lesson 10 warm-up: a tiny demo of interface
// satisfaction. The Greeter interface has one method; two struct types
// (English, Finnish) each implement it — and that's all it takes for them
// to be "Greeters." No `implements` keyword. No declared relationship.
package greet

// Greeter is a one-method interface: anything with `Greet(name string) string`
// is a Greeter.
type Greeter interface {
	Greet(name string) string
}

// English greets in English.
type English struct{}

// Greet returns "Hello, <name>!".
//
// Examples:
//
//	English{}.Greet("World")  → "Hello, World!"
//	English{}.Greet("Aki")    → "Hello, Aki!"
//
// Hint: fmt.Sprintf("Hello, %s!", name).
func (English) Greet(name string) string {
	panic("TODO: return \"Hello, <name>!\"")
}

// Finnish greets in Finnish.
type Finnish struct{}

// Greet returns "Hei, <name>!".
//
// Examples:
//
//	Finnish{}.Greet("World")  → "Hei, World!"
//	Finnish{}.Greet("Aki")    → "Hei, Aki!"
//
// Hint: fmt.Sprintf("Hei, %s!", name).
func (Finnish) Greet(name string) string {
	panic("TODO: return \"Hei, <name>!\"")
}
