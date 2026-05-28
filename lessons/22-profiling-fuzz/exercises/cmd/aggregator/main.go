// Package main is the lesson 22 aggregator CLI exercises tree.
//
// Extends L19's main.go with -mode=pool + -workers=N flags. Walk +
// WalkLocked carry forward unchanged; WalkPool wiring is included so
// that once students implement WalkPool in internal/aggregator the CLI
// just works.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/exercises/internal/aggregator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "cancelled by user")
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	dir, timeout, mode, workers, err := parseFlags(args)
	if err != nil {
		return err
	}

	var result aggregator.WalkResult
	var walkErr error
	switch mode {
	case "channel":
		result, walkErr = aggregator.Walk(ctx, dir, timeout)
	case "mutex":
		result, walkErr = aggregator.WalkLocked(ctx, dir, timeout)
	case "pool":
		result, walkErr = aggregator.WalkPool(ctx, dir, timeout, workers)
	default:
		return fmt.Errorf("invalid -mode %q (want channel|mutex|pool)", mode)
	}

	// Print partial output even on error/cancellation — caller may want
	// what we DID collect.
	for _, path := range result.TimedOut {
		fmt.Fprintf(stderr, "timeout: %s\n", path)
	}

	keys := make([]string, 0, len(result.Counts))
	for k := range result.Counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(stdout, "%s: %d\n", k, result.Counts[k])
	}

	return walkErr
}

func parseFlags(args []string) (string, time.Duration, string, int, error) {
	var (
		dir     string
		timeout time.Duration
		mode    = "channel"
		workers int
	)
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "-dir="):
			dir = a[len("-dir="):]
		case strings.HasPrefix(a, "-timeout="):
			t, err := time.ParseDuration(a[len("-timeout="):])
			if err != nil {
				return "", 0, "", 0, fmt.Errorf("invalid -timeout: %w", err)
			}
			timeout = t
		case strings.HasPrefix(a, "-mode="):
			mode = a[len("-mode="):]
		case strings.HasPrefix(a, "-workers="):
			w, err := strconv.Atoi(a[len("-workers="):])
			if err != nil || w < 0 {
				return "", 0, "", 0, fmt.Errorf("invalid -workers: %s", a)
			}
			workers = w
		default:
			return "", 0, "", 0, fmt.Errorf("unknown flag %q", a)
		}
	}
	if dir == "" {
		return "", 0, "", 0, fmt.Errorf("-dir=<path> is required")
	}
	return dir, timeout, mode, workers, nil
}
