// Package main is the lesson 16 aggregator CLI.
//
// Usage:
//
//	aggregator -dir=<path>
//
// Walks the directory, parses each log file in parallel, and prints
// the combined level counts to stdout.
//
// Exit codes:
//   - 0 on success
//   - 1 on any error (missing -dir, bad dir, malformed file, etc.).
//     Error message goes to stderr.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/16-goroutines-channels/exercises/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point. Parses args, calls aggregator.Walk,
// prints results to stdout in alphabetical order.
//
// Hint:
//  1. Parse -dir=<path> from args (required; error if missing).
//  2. counts, err := aggregator.Walk(dir)
//  3. if err != nil → return err
//  4. Sort keys alphabetically.
//  5. For each key: fmt.Fprintf(stdout, "%s: %d\n", key, counts[key])
func run(args []string, stdout io.Writer) error {
	_ = strings.HasPrefix
	_ = aggregator.Walk
	_ = sort.Strings
	panic("TODO: parse -dir, call aggregator.Walk, print sorted results")
}
