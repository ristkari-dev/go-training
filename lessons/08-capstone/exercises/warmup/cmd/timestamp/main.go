// Package main is a tiny binary that prints the current time via the
// clock package. Demonstrates the cmd/<binaryname>/main.go convention.
//
// Run with:
//
//	go run ./lessons/08-capstone/exercises/warmup/cmd/timestamp
//
// Once clock.Now() is implemented, output is the current UTC time formatted
// as "2026-05-18 14:30:00" (whatever the current time is).
package main

import (
	"fmt"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/warmup/clock"
)

func main() {
	fmt.Println(clock.Now())
}
