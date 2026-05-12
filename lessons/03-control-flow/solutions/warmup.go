// This file holds the warm-up reference solution for lesson 03: Control flow.
package main

import "strconv"

// WarmupClassify returns "positive", "negative", or "zero" for the given int.
func WarmupClassify(n int) string {
	if n > 0 {
		return "positive"
	} else if n < 0 {
		return "negative"
	}
	return "zero"
}

// WarmupFizzBuzz returns the FizzBuzz sequence for the integers 1..n.
func WarmupFizzBuzz(n int) []string {
	out := []string{}
	for i := 1; i <= n; i++ {
		switch {
		case i%15 == 0:
			out = append(out, "FizzBuzz")
		case i%3 == 0:
			out = append(out, "Fizz")
		case i%5 == 0:
			out = append(out, "Buzz")
		default:
			out = append(out, strconv.Itoa(i))
		}
	}
	return out
}
