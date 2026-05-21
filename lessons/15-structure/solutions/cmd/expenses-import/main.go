// Package main is the lesson 15 CSV expense importer (verbatim L13
// carry-forward with import-path rewrites to the new internal/ layout).
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/csvimport"
	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/store"
)

func main() {
	if err := run(os.Args[1:], os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, inputReader io.Reader) error {
	filePath, inputPath, err := parseFlags(args)
	if err != nil {
		return err
	}

	r := inputReader
	if inputPath != "" {
		f, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("open input %s: %w", inputPath, err)
		}
		defer f.Close()
		r = f
	}

	es, err := csvimport.Parse(r)
	if err != nil {
		return err
	}

	s := store.NewJSONStore(filePath)
	if err := s.Save(es); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Printf("imported %d expenses to %s\n", len(es), filePath)
	return nil
}

func parseFlags(args []string) (filePath, inputPath string, err error) {
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "-file="):
			filePath = a[len("-file="):]
		case strings.HasPrefix(a, "-input="):
			inputPath = a[len("-input="):]
		default:
			return "", "", fmt.Errorf("unknown flag %q (try -file=<path> or -input=<path>)", a)
		}
	}
	if filePath == "" {
		return "", "", fmt.Errorf("-file=<path> is required")
	}
	return filePath, inputPath, nil
}
