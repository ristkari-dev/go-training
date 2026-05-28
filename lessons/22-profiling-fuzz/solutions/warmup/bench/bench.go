// Package bench is the lesson 22 warm-up reference implementation.
package bench

import (
	"fmt"
	"strings"
)

var levels = []string{"INFO", "WARN", "ERROR"}

// GenLines returns n deterministic, well-formed log lines.
func GenLines(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "2026-01-02T15:04:%02d %s message number %d\n", i%60, levels[i%3], i)
	}
	return b.String()
}
