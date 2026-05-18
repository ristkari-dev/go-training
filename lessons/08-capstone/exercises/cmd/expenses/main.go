// Package main is the lesson 08 capstone — an expense tracker CLI.
//
// Usage:
//
//	expenses [-file=path] <subcommand> [args...]
//
// Subcommands:
//
//	add DATE AMOUNT CATEGORY    Append an expense to the JSON file.
//	list                         Print all expenses, one per line.
//	summary                      Print per-category totals and a bar chart.
//
// The -file=path option (must appear BEFORE the subcommand) sets the JSON
// file path. Default is $HOME/.expenses.json.
//
// Run with:
//
//	go run ./lessons/08-capstone/exercises/cmd/expenses -file=/tmp/x.json add 2026-05-18 4.50 coffee
//	go run ./lessons/08-capstone/exercises/cmd/expenses -file=/tmp/x.json list
//	go run ./lessons/08-capstone/exercises/cmd/expenses -file=/tmp/x.json summary
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/storage"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/summary"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point. It returns an error instead of calling
// os.Exit so integration tests can drive it without process teardown.
func run(args []string) error {
	// Parse -file=path option (must come first if present).
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

// parseFileOption pulls -file=path off the front of args (if present) and
// returns the resolved path plus the remaining args. If -file= isn't given,
// the path defaults to $HOME/.expenses.json.
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

// cmdAdd parses DATE AMOUNT CATEGORY from args, appends an expense to the
// JSON file, and prints a confirmation line.
//
// Expected output:
//
//	added: 2026-05-18  €4.50    coffee
func cmdAdd(path string, args []string) error {
	// Hints:
	// - if len(args) != 3 → return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	// - amount, err := strconv.ParseFloat(args[1], 64); handle err
	// - es, err := storage.LoadExpenses(path); handle err
	// - e := expense.Expense{Date: args[0], Amount: amount, Category: args[2]}
	// - es = append(es, e)
	// - if err := storage.SaveExpenses(path, es); ...
	// - fmt.Println("added:", e.Format())
	_ = strconv.ParseFloat
	_ = expense.Expense{}
	_ = storage.LoadExpenses
	panic("TODO: parse DATE AMOUNT CATEGORY; storage.LoadExpenses; append; storage.SaveExpenses; print confirmation via e.Format()")
}

// cmdList loads the JSON file and prints each expense on its own line via
// Format(). Prints nothing for an empty file (no header).
func cmdList(path string) error {
	panic("TODO: storage.LoadExpenses; for each e: fmt.Println(e.Format())")
}

// cmdSummary loads the JSON file, computes per-category totals, prints
// the totals + bar chart, and the biggest category.
//
// Expected output (for 3 expenses totalling 40 across coffee/lunch/rent):
//
//	3 expenses, total €40.00
//	by category:
//	  coffee     €4.50   ▌
//	  lunch      €12.00  ████
//	  rent       €23.50  ████████
//	biggest category: rent (€23.50)
//
// Empty file output:
//
//	0 expenses, total €0.00
//
// (No "by category" or "biggest" sections when empty.)
func cmdSummary(path string) error {
	// Hints for the body once you've loaded es:
	// - if len(es) == 0 → print "0 expenses, total €0.00" and return nil
	// - var total float64; for _, e := range es { total += e.Amount }
	// - fmt.Printf("%d expenses, total €%.2f\n", len(es), total)
	// - fmt.Println("by category:")
	// - totals := summary.TotalsByCategory(es)
	// - sort the keys for deterministic output (sort.Strings on a keys slice)
	// - find maxTotal for bar scaling: for _, t := range totals { if t > maxTotal { maxTotal = t } }
	// - for each sorted key: fmt.Printf("  %-10s €%-7.2f %s\n", k, totals[k], summary.Bar(totals[k], maxTotal, 8))
	// - name, biggestTotal := summary.BiggestCategory(totals)
	// - fmt.Printf("biggest category: %s (€%.2f)\n", name, biggestTotal)
	_ = sort.Strings
	_ = summary.Bar
	panic("TODO: storage.LoadExpenses; compute total (sum of e.Amount); print header; if non-empty: print per-category totals with summary.Bar; print summary.BiggestCategory")
}
