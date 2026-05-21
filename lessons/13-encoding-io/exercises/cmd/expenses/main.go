// Package main is the lesson 13 expense tracker CLI.
//
// Carried forward from lesson 11: same CLI surface, same error handling.
// The only change is which store package it imports — lesson 13's store
// uses streaming Encoder/Decoder internally, but the public API is
// identical, so this CLI doesn't need to change.
//
// Uses errors.Is(err, store.ErrNotFound) to make first-time use friendly.
// Uses errors.As to extract *store.ParseError for richer error messages
// when the JSON file is malformed.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/exercises/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
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
		return nil, fmt.Errorf("unsupported store kind %q", kind)
	}
}

// loadOrEmpty calls s.Load and treats ErrNotFound as "empty + nil" so
// callers don't need to repeat the errors.Is check in every subcommand.
// Other errors pass through.
func loadOrEmpty(s store.Store) ([]expense.Expense, error) {
	es, err := s.Load()
	if errors.Is(err, store.ErrNotFound) {
		return []expense.Expense{}, nil
	}
	return es, err
}

func cmdAdd(s store.Store, args []string) error {
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
	fmt.Println("added:", e.Format())
	return nil
}

func cmdList(s store.Store) error {
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}
	for _, e := range es {
		fmt.Println(e.Format())
	}
	return nil
}

func cmdSummary(s store.Store) error {
	es, err := loadOrEmpty(s)
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
