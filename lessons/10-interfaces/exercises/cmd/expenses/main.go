// Package main is the lesson 10 expense tracker CLI — same shape as L08's
// CLI but refactored to use the store.Store interface, with a new
// -store=mem|json flag to pick the implementation at runtime.
//
// Usage:
//
//	expenses [-store=mem|json] [-file=path] <subcommand> [args...]
//
// Subcommands:
//
//	add DATE AMOUNT CATEGORY    Append an expense.
//	list                         Print all expenses.
//	summary                      Print per-category totals.
//
// The -store flag selects the backing store: "json" (default) uses
// JSONStore at the -file path; "mem" uses MemoryStore (data is lost
// when the process exits — useful for tests).
//
// The -file flag (only meaningful for -store=json) defaults to
// $HOME/.expenses.json. Both flags must appear BEFORE the subcommand.
//
// IMPORTANT: This file ships fully working in lesson 10 — your job is
// to implement store.Store + JSONStore + MemoryStore in store/store.go.
// Once those are done, this binary works.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/10-interfaces/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/10-interfaces/exercises/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
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
		return cmdAdd(s, args[1:])
	case "list":
		return cmdList(s)
	case "summary":
		return cmdSummary(s)
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
		// parseStoreOption guards this; defensive only.
		return nil, fmt.Errorf("unsupported store kind %q", kind)
	}
}

func cmdAdd(s store.Store, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	}
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", args[1], err)
	}
	es, err := s.Load()
	if err != nil {
		return err
	}
	e := expense.Expense{Date: args[0], Amount: amount, Category: args[2]}
	es = append(es, e)
	if err := s.Save(es); err != nil {
		return err
	}
	fmt.Println("added:", e.Format())
	return nil
}

func cmdList(s store.Store) error {
	es, err := s.Load()
	if err != nil {
		return err
	}
	for _, e := range es {
		fmt.Println(e.Format())
	}
	return nil
}

func cmdSummary(s store.Store) error {
	es, err := s.Load()
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
	totals := map[string]float64{}
	for _, e := range es {
		totals[e.Category] += e.Amount
	}

	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("  %-10s €%-7.2f\n", k, totals[k])
	}
	return nil
}
