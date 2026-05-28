// Package main is the lesson 22 aggregator profiler.
//
// Runs WalkPool under load and writes CPU + heap profiles for analysis
// with `go tool pprof`. Optionally serves a live pprof endpoint.
//
// Usage:
//
//	aggregator-profile -n=2000 -cpuprofile=cpu.prof -memprofile=heap.prof
//	aggregator-profile -dir=/path/to/logs -duration=5s
//	aggregator-profile -httppprof=:6060 -duration=30s   # live: localhost:6060/debug/pprof/
//
// Then: go tool pprof -http=:8080 cpu.prof
package main

import (
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the testable entry point. See the solution for the full
// implementation (flag parsing, synthetic dataset, pprof start/stop,
// WalkPool loop, heap dump).
func run(args []string, stdout, stderr io.Writer) int {
	_ = args
	_ = stdout
	_ = stderr
	panic("TODO: parse flags; build/resolve dataset; StartCPUProfile; loop WalkPool until -duration; WriteHeapProfile")
}
