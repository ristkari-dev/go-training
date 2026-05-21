// Package util is a tiny demo of the internal/ access rule.
//
// This package lives under warmup/internal/util/. Per Go's module
// system, packages under an internal/ directory can ONLY be imported
// from code inside the parent directory or its descendants — i.e.,
// from anything inside warmup/. Code in a sibling directory (say,
// lessons/14-time-strings-regex/...) trying to import this package
// would fail at compile time:
//
//	package warmup/internal/util is not allowed: use of internal package
//
// The compiler enforces it. There's no opt-out.
package util

import "fmt"

// Money formats a euro amount as "€<2-decimal>". E.g., Money(4.5) →
// "€4.50".
//
// Hint: return fmt.Sprintf("€%.2f", amount)
func Money(amount float64) string {
	_ = fmt.Sprintf
	panic("TODO: fmt.Sprintf with %.2f")
}
