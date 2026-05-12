// Package main is the reference implementation for lesson 03: Control flow.
//
// Both `Categorise` and `Tally` are pure functions covered by main_test.go.
// `main()` demos the interactive Scanln loop using them — it is NOT covered
// by `go test`; verify it interactively with:
//
//	printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
package main

import (
	"fmt"
	"os"
)

// Categorise returns "snack" for amounts under €10, "regular" for €10..€50
// (inclusive on both ends), and "splurge" above €50.
func Categorise(amount float64) string {
	switch {
	case amount < 10:
		return "snack"
	case amount <= 50:
		return "regular"
	default:
		return "splurge"
	}
}

// Tally counts how many entries of amounts fall in each category.
func Tally(amounts []float64) (snack, regular, splurge int) {
	for _, a := range amounts {
		switch Categorise(a) {
		case "snack":
			snack++
		case "regular":
			regular++
		case "splurge":
			splurge++
		}
	}
	return snack, regular, splurge
}

func main() {
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
