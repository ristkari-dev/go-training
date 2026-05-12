// Package exercises is the starter code for lesson 03: Control flow.
//
// This file holds the WARM-UP exercise. Two small functions to build muscle
// memory for "branch on a value with if" and "loop a fixed number of times
// with for." Make the failing tests in warmup_test.go pass.
package exercises

import "strconv"

// WarmupClassify returns "positive", "negative", or "zero" for the given int.
//
// Examples:
//
//	WarmupClassify(7)   → "positive"
//	WarmupClassify(-3)  → "negative"
//	WarmupClassify(0)   → "zero"
//
// Hint: an if / else if / else chain is the most natural shape. Compare with
// 0 using >, <, and ==.
func WarmupClassify(n int) string {
	panic("TODO: return \"positive\" / \"negative\" / \"zero\" using if/else if/else")
}

// WarmupFizzBuzz returns the classic FizzBuzz sequence for the integers 1..n.
//
// For each integer i in 1..n (inclusive):
//   - if i is divisible by 15: "FizzBuzz"
//   - else if i is divisible by 3: "Fizz"
//   - else if i is divisible by 5: "Buzz"
//   - else: the integer formatted with strconv.Itoa(i)
//
// If n < 1, return an empty (non-nil) slice.
//
// Examples:
//
//	WarmupFizzBuzz(5)   → ["1", "2", "Fizz", "4", "Buzz"]
//	WarmupFizzBuzz(15)  → [..., "Fizz", "13", "14", "FizzBuzz"]
//	WarmupFizzBuzz(0)   → []
//
// Hint: a C-style for loop (`for i := 1; i <= n; i++`) plus an if/else if
// chain inside is the simplest implementation. Use strconv.Itoa to convert
// the int to its decimal string. Append to the result slice as you go.
func WarmupFizzBuzz(n int) []string {
	_ = strconv.Itoa // keep the import compiling until you use it
	panic("TODO: build the FizzBuzz slice with a for loop and an if/else if chain")
}
