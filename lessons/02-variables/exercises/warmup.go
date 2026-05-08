// Package exercises is the starter code for lesson 02: Variables, types, operators.
//
// This file holds the WARM-UP exercise. Two small functions to build muscle
// memory for "declare a variable" and "convert between types." Make the
// failing tests in warmup_test.go pass.
package exercises

// WarmupZeroValues returns the zero values of int, float64, string, and bool
// in that order.
//
// In Go, every variable has a value. If you declare it without assigning, the
// variable holds the type's "zero value." This function demonstrates the four
// most common ones.
//
// Examples:
//
//	WarmupZeroValues() → (0, 0.0, "", false)
//
// Hint: declare four variables with `var name type` and return them. You don't
// need to assign anything — the zero values are already there.
func WarmupZeroValues() (int, float64, string, bool) {
	panic("TODO: declare an int, a float64, a string, and a bool with var, then return them")
}

// WarmupConvert performs two simple numeric type conversions and returns both.
//
// Given an int and a float64:
// - return the int as a float64 using float64(intVal)
// - return the float64 as an int using int(floatVal); the fractional part is truncated toward zero
//
// Examples:
//
//	WarmupConvert(5, 3.7)   → (5.0, 3)
//	WarmupConvert(0, -2.9)  → (0.0, -2)
//
// Hint: Go is strict about types. To "promote" an int to float64 you write
// float64(x); to truncate a float64 to int you write int(y).
func WarmupConvert(intVal int, floatVal float64) (asFloat float64, asInt int) {
	panic("TODO: return float64(intVal) and int(floatVal)")
}
