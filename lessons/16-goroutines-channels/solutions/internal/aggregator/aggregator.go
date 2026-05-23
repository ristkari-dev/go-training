// Package aggregator is the lesson 16 reference implementation.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/16-goroutines-channels/solutions/internal/logparse"
)

type fileResult struct {
	path   string
	counts map[string]int
	err    error
}

// Walk walks dir, spawns one goroutine per file, and returns the
// combined level counts. On any error, returns partial counts +
// wrapped error.
func Walk(dir string) (map[string]int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("aggregator: read dir %s: %w", dir, err)
	}

	files := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		files = append(files, filepath.Join(dir, e.Name()))
	}

	resultsCh := make(chan fileResult, len(files))

	for _, path := range files {
		go processFile(path, resultsCh)
	}

	combined := map[string]int{}
	for i := 0; i < len(files); i++ {
		r := <-resultsCh
		if r.err != nil {
			return combined, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
		}
		for level, count := range r.counts {
			combined[level] += count
		}
	}
	return combined, nil
}

func processFile(path string, resultsCh chan<- fileResult) {
	f, err := os.Open(path)
	if err != nil {
		resultsCh <- fileResult{path: path, err: err}
		return
	}
	defer f.Close()

	entries, err := logparse.Parse(f)
	if err != nil {
		resultsCh <- fileResult{path: path, err: err}
		return
	}

	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Level]++
	}
	resultsCh <- fileResult{path: path, counts: counts}
}
