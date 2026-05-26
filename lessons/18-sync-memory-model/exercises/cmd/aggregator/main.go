// Package main is the lesson 18 aggregator CLI.
//
// Usage:
//
//	aggregator -dir=<path>                          # mode=channel (default)
//	aggregator -dir=<path> -mode=channel            # explicit channel
//	aggregator -dir=<path> -mode=mutex              # mutex-based
//	aggregator -dir=<path> -mode=mutex -timeout=5s  # combine
//
// Walks the directory, parses each log file in parallel, prints the
// combined level counts to stdout. Files that exceed -timeout are
// printed to stderr. -mode selects which Walk variant to use:
// channel-based (L17) or mutex-based (L18).
//
// Exit codes: 0 on success; 1 on any error.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/18-sync-memory-model/exercises/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point.
//
// Hint:
//  1. Parse -dir, -timeout, -mode flags. -mode defaults to "channel".
//  2. Pick walk function: channel → aggregator.Walk; mutex → aggregator.WalkLocked.
//     Invalid -mode value → return error.
//  3. result, err := walkFn(dir, timeout)
//  4. Print TimedOut to stderr, sorted Counts to stdout (per L17).
func run(args []string, stdout io.Writer, stderr io.Writer) error {
	_ = strings.HasPrefix
	_ = aggregator.Walk
	_ = aggregator.WalkLocked
	_ = sort.Strings
	_ = time.ParseDuration
	panic("TODO: parse -dir + -timeout + -mode; pick Walk or WalkLocked; print results")
}
