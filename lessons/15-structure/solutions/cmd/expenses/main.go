// Package main is the lesson 15 expense tracker CLI — the v2 polished
// version. The summary subcommand now uses internal/summary for
// totals/biggest/bar; the rest matches L11/L13.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/store"
	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/summary"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		var pe *store.ParseError
		if errors.As(err, &pe) {
			fmt.Fprintf(os.Stderr, "error: could not parse %s at line %d: %v\n",
				pe.Path, pe.Line, pe.Cause)
		} else {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

// run is the testable entry point. Takes args + an io.Writer for stdout
// output (so tests can capture into a bytes.Buffer).
//
// The signature changed from L11/L13 (which used os.Stdout implicitly)
// — passing the Writer explicitly is required for golden file tests.
func run(args []string, stdout io.Writer) error {
	storeKind, args, err := parseStoreOption(args)
	if err != nil {
		return err
	}
	path, args, err := parseFileOption(args)
	if err != nil {
		return err
	}

	s, err := buildStore(storeKind, path)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("missing subcommand (try: add, list, summary)")
	}
	switch args[0] {
	case "add":
		return cmdAdd(s, args[1:], stdout)
	case "list":
		return cmdList(s, stdout)
	case "summary":
		return cmdSummary(s, stdout)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func parseStoreOption(args []string) (string, []string, error) {
	if len(args) > 0 && strings.HasPrefix(args[0], "-store=") {
		v := args[0][len("-store="):]
		if v != "mem" && v != "json" {
			return "", nil, fmt.Errorf("invalid -store value %q (must be mem or json)", v)
		}
		return v, args[1:], nil
	}
	return "json", args, nil
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

func buildStore(kind, path string) (store.Store, error) {
	switch kind {
	case "mem":
		return store.NewMemoryStore(), nil
	case "json":
		return store.NewJSONStore(path), nil
	default:
		return nil, fmt.Errorf("unsupported store kind %q", kind)
	}
}

func loadOrEmpty(s store.Store) ([]expense.Expense, error) {
	es, err := s.Load()
	if errors.Is(err, store.ErrNotFound) {
		return []expense.Expense{}, nil
	}
	return es, err
}

func cmdAdd(s store.Store, args []string, w io.Writer) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	}
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", args[1], err)
	}
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}
	e := expense.Expense{Date: args[0], Amount: amount, Category: args[2]}
	es = append(es, e)
	if err := s.Save(es); err != nil {
		return err
	}
	fmt.Fprintln(w, "added:", e.Format())
	return nil
}

func cmdList(s store.Store, w io.Writer) error {
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}
	for _, e := range es {
		fmt.Fprintln(w, e.Format())
	}
	return nil
}

// cmdSummary uses internal/summary for TotalsByCategory + BiggestCategory + Bar.
// Output: totals per category (alphabetical), then biggest line, then bar chart.
func cmdSummary(s store.Store, w io.Writer) error {
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}

	var total float64
	for _, e := range es {
		total += e.Amount
	}
	fmt.Fprintf(w, "%d expenses, total €%.2f\n", len(es), total)
	if len(es) == 0 {
		return nil
	}

	totals := summary.TotalsByCategory(es)

	// Sorted by category name for stable output.
	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	// Sort via the strings convention.
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	fmt.Fprintln(w, "by category:")
	for _, k := range keys {
		fmt.Fprintf(w, "  %-10s €%-7.2f\n", k, totals[k])
	}

	// Biggest line + bar chart.
	name, biggest := summary.BiggestCategory(totals)
	if name != "" {
		fmt.Fprintf(w, "biggest: %s €%.2f\n", name, biggest)
		fmt.Fprintf(w, "  %s\n", summary.Bar(biggest, biggest, 20))
	}

	return nil
}
