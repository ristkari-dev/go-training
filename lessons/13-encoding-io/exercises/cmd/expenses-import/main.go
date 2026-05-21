// Package main is the lesson 13 CSV expense importer.
//
// Reads CSV from stdin (or the file named by -input=<path>), parses
// via csvimport.Parse, writes the result to the JSON file at -file=<path>.
//
// Usage:
//
//	expenses-import -file=expenses.json < data.csv
//	expenses-import -file=expenses.json -input=data.csv
//
// The -file flag is REQUIRED. There's no default — the importer is
// fundamentally about persistence to disk. Forgetting -file is a
// loud failure (printed help + exit 1), not a silent one.
//
// Exit codes:
//   - 0 on success
//   - 1 on any error (missing flag, can't open input, malformed CSV,
//     store.Save failure). Error message goes to stderr.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/exercises/csvimport"
	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/exercises/store"
)

func main() {
	if err := run(os.Args[1:], os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point. Reads CSV from inputReader (or from
// a file named by -input=); writes JSON to the path from -file=.
//
// Hint:
//  1. Parse args: -file=<required>, -input=<optional, defaults to inputReader>
//  2. If -input was given: open it; defer close. Else use inputReader.
//  3. es, err := csvimport.Parse(r); if err != nil → return err
//  4. s := store.NewJSONStore(filePath); return s.Save(es)
//
// "Accept io.Reader, return error" is the testability win — test calls
// run with strings.NewReader(...) as inputReader.
func run(args []string, inputReader io.Reader) error {
	_ = strings.HasPrefix
	_ = csvimport.Parse
	_ = store.NewJSONStore
	panic("TODO: parse -file= (required) and -input= (optional); call csvimport.Parse; call s.Save")
}
