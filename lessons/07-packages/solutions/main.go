// Package main is the lesson 07 driver (solutions reference).
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/07-packages/solutions/expense"
	"github.com/ristkari-dev/go-training/lessons/07-packages/solutions/greet"
)

func main() {
	fmt.Println(greet.Greet("Aki"))
	fmt.Println(strings.Repeat("=", 16))

	es := []expense.Expense{
		{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-15", Amount: 12, Category: "lunch"},
		{Date: "2026-05-15", Amount: 75, Category: "rent"},
		{Date: "2026-05-15", Amount: 9.99, Category: "coffee"},
	}
	for _, e := range es {
		line := e.Format()
		if e.IsHigh() {
			line += "  (high)"
		}
		fmt.Println(line)
	}

	fmt.Println(strings.Repeat("-", 16))
	fmt.Print("totals:")

	totals := expense.TotalsByCategory(es)
	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf(" %s=%.2f", k, totals[k])
	}
	fmt.Println()
}
