// Package main is the lesson 17 aggregator CLI.
//
// Usage:
//
//	aggregator -dir=<path>                    # no timeout
//	aggregator -dir=<path> -timeout=5s        # 5-second per-file timeout
//
// Walks the directory, parses each log file in parallel, prints the
// combined level counts to stdout. Files that exceed -timeout are
// printed to stderr.
//
// Exit codes:
//   - 0 on success (including success with timeouts)
//   - 1 on any error (missing -dir, bad dir, malformed file, etc.)
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/17-select-timers/exercises/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point. Parses args, calls aggregator.Walk,
// prints sorted level counts to stdout and any timed-out paths to stderr.
//
// Hint:
//  1. Parse -dir=<path> (required) and -timeout=<dur> (optional, default 0)
//     via time.ParseDuration.
//  2. result, err := aggregator.Walk(dir, timeout)
//  3. if err != nil → return err
//  4. For each result.TimedOut entry: fmt.Fprintf(stderr, "timeout: %s\n", path)
//  5. Sort result.Counts keys alphabetically; print "LEVEL: count" per line to stdout.
func run(args []string, stdout io.Writer, stderr io.Writer) error {
	_ = strings.HasPrefix
	_ = aggregator.Walk
	_ = sort.Strings
	_ = time.ParseDuration
	panic("TODO: parse -dir + -timeout; call Walk; print stdout + stderr")
}
