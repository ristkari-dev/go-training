// Package bench is the lesson 22 warm-up: a deterministic synthetic
// log-line generator used to feed benchmarks.
//
// Benchmarks need a fixed, repeatable input so ns/op is comparable
// across runs. GenLines builds n well-formed log lines that
// logparse.Parse accepts — no randomness, no I/O.
package bench

import (
	"fmt"
	"strings"
)

// levels cycles through the three valid log levels.
var levels = []string{"INFO", "WARN", "ERROR"}

// GenLines returns n synthetic log lines joined by '\n' (one trailing
// newline-free block suitable for strings.NewReader). Each line is
// well-formed: "2006-01-02T15:04:SS LEVEL message ...".
//
// Hint:
//
//	var b strings.Builder
//	for i := 0; i < n; i++ {
//	    fmt.Fprintf(&b, "2026-01-02T15:04:%02d %s message number %d\n",
//	        i%60, levels[i%3], i)
//	}
//	return b.String()
func GenLines(n int) string {
	_ = fmt.Fprintf
	_ = strings.Builder{}
	_ = levels
	panic("TODO: build n well-formed log lines into a strings.Builder; return the string")
}
