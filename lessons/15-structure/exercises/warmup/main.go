// Package main is the lesson 15 warm-up: importing an internal/ package
// from a sibling directory in the SAME module.
//
// This binary compiles because warmup/main.go and
// warmup/internal/util/format.go share a common ancestor (warmup/).
// The internal/ rule allows imports from descendants of the parent of
// the internal/ directory.
//
// Try moving main.go up one level (to lessons/15-structure/exercises/)
// and re-running `go build`. The build will fail:
//
//	use of internal package github.com/.../warmup/internal/util
//	  not allowed
//
// That's the rule, enforced by the compiler.
package main

import (
	"fmt"

	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/warmup/internal/util"
)

func main() {
	fmt.Println(util.Money(4.50))
}
