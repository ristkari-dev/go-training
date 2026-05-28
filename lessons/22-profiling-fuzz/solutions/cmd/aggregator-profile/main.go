// Package main is the lesson 22 aggregator profiler reference impl.
//
// Runs WalkPool under load and writes CPU + heap profiles for analysis
// with `go tool pprof`. Optionally serves a live pprof endpoint via the
// net/http/pprof side-effect import.
//
// Usage:
//
//	aggregator-profile -n=2000 -cpuprofile=cpu.prof -memprofile=heap.prof
//	aggregator-profile -dir=/path/to/logs -duration=5s
//	aggregator-profile -httppprof=:6060 -duration=30s
//
// Then: go tool pprof -http=:8080 cpu.prof
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	_ "net/http/pprof" // registers /debug/pprof handlers on DefaultServeMux

	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/solutions/internal/aggregator"
	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/solutions/warmup/bench"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aggregator-profile", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "", "directory of log files (empty → generate synthetic)")
	n := fs.Int("n", 2000, "synthetic file count when -dir is empty")
	lines := fs.Int("lines", 50, "lines per synthetic file")
	workers := fs.Int("workers", 0, "worker count (0 = NumCPU)")
	duration := fs.Duration("duration", 0, "loop WalkPool until this elapses (0 = single pass)")
	cpuProfile := fs.String("cpuprofile", "", "write CPU profile to this file")
	memProfile := fs.String("memprofile", "", "write heap profile to this file")
	httpAddr := fs.String("httppprof", "", "serve live pprof on this addr (e.g. :6060)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Optional live pprof endpoint. The blank import registered the
	// handlers; we just need a server on DefaultServeMux.
	if *httpAddr != "" {
		go func() {
			fmt.Fprintf(stderr, "live pprof: http://%s/debug/pprof/\n", *httpAddr)
			if err := http.ListenAndServe(*httpAddr, nil); err != nil {
				fmt.Fprintln(stderr, "pprof http:", err)
			}
		}()
	}

	// Resolve the dataset directory.
	target := *dir
	if target == "" {
		d, err := generateDataset(*n, *lines)
		if err != nil {
			fmt.Fprintln(stderr, "generate dataset:", err)
			return 1
		}
		defer os.RemoveAll(d)
		target = d
	}

	// CPU profile spans the whole workload.
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintln(stderr, "create cpuprofile:", err)
			return 1
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(stderr, "start cpu profile:", err)
			return 1
		}
		defer pprof.StopCPUProfile()
	}

	// Run WalkPool once, or loop until -duration elapses (so CPU
	// profiling gathers enough samples).
	var (
		result aggregator.WalkResult
		err    error
		passes int
		start  = time.Now()
	)
	for {
		result, err = aggregator.WalkPool(context.Background(), target, 0, *workers)
		if err != nil {
			fmt.Fprintln(stderr, "walk:", err)
			return 1
		}
		passes++
		if *duration <= 0 || time.Since(start) >= *duration {
			break
		}
	}

	// Heap profile is a snapshot — take it after the workload, post-GC.
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			fmt.Fprintln(stderr, "create memprofile:", err)
			return 1
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintln(stderr, "write heap profile:", err)
			return 1
		}
	}

	total := 0
	for _, c := range result.Counts {
		total += c
	}
	fmt.Fprintf(stdout, "passes=%d files=%d entries=%d timedOut=%d\n",
		passes, *n, total, len(result.TimedOut))
	return 0
}

// generateDataset writes n files of `lines` well-formed log lines each
// into a fresh temp dir and returns its path. Caller removes it.
func generateDataset(n, lines int) (string, error) {
	dir, err := os.MkdirTemp("", "aggprofile-")
	if err != nil {
		return "", err
	}
	block := bench.GenLines(lines)
	for i := 0; i < n; i++ {
		name := filepath.Join(dir, fmt.Sprintf("log-%05d.log", i))
		if err := os.WriteFile(name, []byte(block), 0o644); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}
	return dir, nil
}
