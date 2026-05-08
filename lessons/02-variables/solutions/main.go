// Package solutions is the reference implementation for lesson 02: Variables, types, operators.
package solutions

import "fmt"

// TipRate is the tip percentage applied by Summary.
const TipRate = 0.14

// Total returns the sum of three amounts.
func Total(a, b, c float64) float64 {
	return a + b + c
}

// Average returns the average of three amounts.
func Average(a, b, c float64) float64 {
	return (a + b + c) / 3.0
}

// Summary returns a multi-line summary of three amounts including a 14% tip.
func Summary(a, b, c float64) string {
	total := Total(a, b, c)
	avg := Average(a, b, c)
	tip := total * TipRate
	totalWithTip := total + tip
	return fmt.Sprintf(
		"3 expenses\nTotal: €%.2f\nAverage: €%.2f\nTip (14%%): €%.2f\nTotal with tip: €%.2f",
		total, avg, tip, totalWithTip,
	)
}
