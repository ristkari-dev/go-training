// Package main is the lesson 16 aggregator CLI reference implementation.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/16-goroutines-channels/solutions/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	dir, err := parseDir(args)
	if err != nil {
		return err
	}

	counts, err := aggregator.Walk(dir)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(stdout, "%s: %d\n", k, counts[k])
	}
	return nil
}

func parseDir(args []string) (string, error) {
	for _, a := range args {
		if strings.HasPrefix(a, "-dir=") {
			return a[len("-dir="):], nil
		}
	}
	return "", fmt.Errorf("-dir=<path> is required")
}
