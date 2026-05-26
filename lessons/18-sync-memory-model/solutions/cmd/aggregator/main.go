// Package main is the lesson 18 aggregator CLI reference implementation.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/18-sync-memory-model/solutions/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type walkFunc func(dir string, timeout time.Duration) (aggregator.WalkResult, error)

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	dir, timeout, mode, err := parseFlags(args)
	if err != nil {
		return err
	}

	var fn walkFunc
	switch mode {
	case "channel":
		fn = aggregator.Walk
	case "mutex":
		fn = aggregator.WalkLocked
	default:
		return fmt.Errorf("invalid -mode %q (want channel|mutex)", mode)
	}

	result, err := fn(dir, timeout)
	if err != nil {
		return err
	}

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
	return nil
}

func parseFlags(args []string) (string, time.Duration, string, error) {
	var (
		dir     string
		timeout time.Duration
		mode    = "channel"
	)
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "-dir="):
			dir = a[len("-dir="):]
		case strings.HasPrefix(a, "-timeout="):
			t, err := time.ParseDuration(a[len("-timeout="):])
			if err != nil {
				return "", 0, "", fmt.Errorf("invalid -timeout: %w", err)
			}
			timeout = t
		case strings.HasPrefix(a, "-mode="):
			mode = a[len("-mode="):]
		default:
			return "", 0, "", fmt.Errorf("unknown flag %q", a)
		}
	}
	if dir == "" {
		return "", 0, "", fmt.Errorf("-dir=<path> is required")
	}
	return dir, timeout, mode, nil
}
