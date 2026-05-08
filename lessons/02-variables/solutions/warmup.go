// Package solutions is the reference implementation for lesson 02: Variables, types, operators.
//
// This file holds the warm-up reference solution.
package solutions

// WarmupZeroValues returns the zero values of int, float64, string, and bool.
func WarmupZeroValues() (int, float64, string, bool) {
	var i int
	var f float64
	var s string
	var b bool
	return i, f, s, b
}

// WarmupConvert performs simple numeric type conversions.
func WarmupConvert(intVal int, floatVal float64) (asFloat float64, asInt int) {
	return float64(intVal), int(floatVal)
}
