// Package exercises is the starter code for lesson 02: Variables, types, operators.
//
// This file holds the MAIN exercise. Implement Total, Average, and Summary
// so the failing tests in main_test.go pass.
package exercises

// TipRate is the tip percentage applied by Summary.
//
// Constants in Go declare values that never change. Compare with `var` for
// mutable variables.
const TipRate = 0.14

// Total returns the sum of three amounts.
//
// Examples:
//
//	Total(4.50, 12.00, 23.50) → 40.0
//	Total(0, 0, 0)            → 0.0
func Total(a, b, c float64) float64 {
	panic("TODO: return a + b + c")
}

// Average returns the average of three amounts.
//
// Examples:
//
//	Average(10, 20, 30)         → 20.0
//	Average(4.50, 12.00, 23.50) → 13.333… (exact value depends on float arithmetic)
//
// Hint: divide by 3.0 (a float64), not 3 (an int), to keep float64 arithmetic.
func Average(a, b, c float64) float64 {
	panic("TODO: return (a + b + c) / 3.0")
}

// Summary returns a multi-line summary of three amounts including a 14% tip.
//
// The format is:
//
//	3 expenses
//	Total: €T.TT
//	Average: €A.AA
//	Tip (14%): €P.PP
//	Total with tip: €X.XX
//
// where each amount is printed with two decimal places.
//
// Examples:
//
//	Summary(4.50, 12.00, 23.50) →
//	  "3 expenses\nTotal: €40.00\nAverage: €13.33\nTip (14%): €5.60\nTotal with tip: €45.60"
//
// Hint: use Total and Average above, plus the TipRate constant. Format with
// fmt.Sprintf using %.2f for each amount; separate lines with \n.
func Summary(a, b, c float64) string {
	panic("TODO: implement Summary using Total, Average, and TipRate")
}
