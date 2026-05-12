// Package exercises is the starter code for lesson 03: Control flow.
//
// This file holds the MAIN exercise. Implement Categorise and Tally so the
// failing tests in main_test.go pass. The runMain function below shows the
// interactive Scanln loop — it's *not* tested by `go test`; verify it from
// the terminal once your pure functions pass.
package exercises

import (
	"fmt"
	"os"
)

// Categorise returns "snack" for amounts under €10, "regular" for €10
// (inclusive) up to €50 (inclusive), and "splurge" for amounts above €50.
//
// Examples:
//
//	Categorise(4.50)  → "snack"
//	Categorise(12.00) → "regular"
//	Categorise(50.00) → "regular"   (boundary: 50 is the upper edge of regular)
//	Categorise(75.00) → "splurge"
//	Categorise(0)     → "snack"
//
// Hint: a tagless switch (`switch { case amount < 10: ... }`) reads cleanly
// here. An if/else if/else chain works just as well — try both and pick the
// shape you prefer.
func Categorise(amount float64) string {
	panic("TODO: return \"snack\" / \"regular\" / \"splurge\" by amount")
}

// Tally counts how many entries of amounts fall in each category.
//
// Returns three named ints — the snack count, the regular count, and the
// splurge count — in that order. An empty input returns (0, 0, 0).
//
// Examples:
//
//	Tally([]float64{4.50, 12, 75, 9.99}) → (snack=2, regular=1, splurge=1)
//	Tally([]float64{})                   → (0, 0, 0)
//
// Hint: use `for _, a := range amounts` to walk the slice. Inside, switch
// on Categorise(a) and bump the matching counter.
func Tally(amounts []float64) (snack, regular, splurge int) {
	panic("TODO: walk amounts with for ... range, classify each with Categorise, count")
}

// runMain demos the interactive flow: it reads N from stdin, then reads N
// amounts, classifies and tallies them, and prints a summary. Run it from
// `solutions/main.go`'s main() — exercises don't ship a main() of their own
// because Categorise/Tally must be implemented first.
//
// runMain is *not* covered by `go test`; verify it interactively with:
//
//	printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
//
// Expected:
//
//	Reading 3 amounts...
//	€4.50  → snack
//	€12.00 → regular
//	€75.00 → splurge
//	Tally: 1 snack, 1 regular, 1 splurge
func runMain() {
	var n int
	if _, err := fmt.Scanln(&n); err != nil {
		fmt.Fprintln(os.Stderr, "could not read count:", err)
		return
	}
	if n <= 0 {
		fmt.Println("nothing to do (n must be > 0)")
		return
	}
	fmt.Printf("Reading %d amounts...\n", n)

	amounts := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		var a float64
		if _, err := fmt.Scanln(&a); err != nil {
			fmt.Fprintf(os.Stderr, "could not read amount %d: %v\n", i+1, err)
			return
		}
		amounts = append(amounts, a)
		fmt.Printf("€%-6.2f → %s\n", a, Categorise(a))
	}
	snack, regular, splurge := Tally(amounts)
	fmt.Printf("Tally: %d snack, %d regular, %d splurge\n", snack, regular, splurge)
}
