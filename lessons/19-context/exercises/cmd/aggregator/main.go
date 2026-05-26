// Package main is the lesson 19 aggregator CLI.
//
// CHANGED from L18:
//   - main() wraps the call in signal.NotifyContext(ctx, os.Interrupt)
//     so Ctrl-C cancels the walk gracefully.
//   - run() takes ctx as its new first parameter.
//   - On context.Canceled: prints partial counts to stdout + "cancelled
//     by user" to stderr, exits 0 (user-requested cancellation isn't
//     a failure).
//
// Usage:
//
//	aggregator -dir=<path> [-mode=channel|mutex] [-timeout=<dur>]
//	Ctrl-C cancels mid-walk gracefully.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/19-context/exercises/internal/aggregator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		// User cancellation isn't an error from the user's perspective —
		// print to stderr but exit 0 with whatever partial output we have.
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "cancelled by user")
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point.
//
// Hint:
//  1. Parse -dir, -timeout, -mode (same as L18).
//  2. Pick walkFn: aggregator.Walk or aggregator.WalkLocked.
//  3. result, err := walkFn(ctx, dir, timeout)
//  4. PRINT result.TimedOut to stderr, sorted Counts to stdout — even
//     if err is context.Canceled, because partial output is valuable.
//  5. Return err to main so it can decide the exit code.
func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	_ = strings.HasPrefix
	_ = aggregator.Walk
	_ = aggregator.WalkLocked
	_ = sort.Strings
	_ = time.ParseDuration
	_ = errors.Is
	panic("TODO: parse flags; call walkFn(ctx, ...); print partial result; return err")
}
