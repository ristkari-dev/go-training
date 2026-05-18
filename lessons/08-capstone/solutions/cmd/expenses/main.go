// Package main is the lesson 08 capstone CLI (solutions reference).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/expense"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/storage"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/summary"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	path, args, err := parseFileOption(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("missing subcommand (try: add, list, summary)")
	}
	switch args[0] {
	case "add":
		return cmdAdd(path, args[1:])
	case "list":
		return cmdList(path)
	case "summary":
		return cmdSummary(path)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func parseFileOption(args []string) (string, []string, error) {
	if len(args) > 0 && strings.HasPrefix(args[0], "-file=") {
		return args[0][len("-file="):], args[1:], nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("could not resolve home directory: %w", err)
	}
	return filepath.Join(home, ".expenses.json"), args, nil
}

func cmdAdd(path string, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	}
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", args[1], err)
	}
	es, err := storage.LoadExpenses(path)
	if err != nil {
		return err
	}
	e := expense.Expense{Date: args[0], Amount: amount, Category: args[2]}
	es = append(es, e)
	if err := storage.SaveExpenses(path, es); err != nil {
		return err
	}
	fmt.Println("added:", e.Format())
	return nil
}

func cmdList(path string) error {
	es, err := storage.LoadExpenses(path)
	if err != nil {
		return err
	}
	for _, e := range es {
		fmt.Println(e.Format())
	}
	return nil
}

func cmdSummary(path string) error {
	es, err := storage.LoadExpenses(path)
	if err != nil {
		return err
	}

	var total float64
	for _, e := range es {
		total += e.Amount
	}
	fmt.Printf("%d expenses, total €%.2f\n", len(es), total)
	if len(es) == 0 {
		return nil
	}

	fmt.Println("by category:")
	totals := summary.TotalsByCategory(es)

	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var maxTotal float64
	for _, t := range totals {
		if t > maxTotal {
			maxTotal = t
		}
	}

	for _, k := range keys {
		fmt.Printf("  %-10s €%-7.2f %s\n", k, totals[k], summary.Bar(totals[k], maxTotal, 8))
	}

	name, biggestTotal := summary.BiggestCategory(totals)
	fmt.Printf("biggest category: %s (€%.2f)\n", name, biggestTotal)
	return nil
}
