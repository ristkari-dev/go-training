// Command build-index produces a static slides site under dist/: every lesson's
// slides copied into place, the shared reveal.js assets copied alongside, and a
// generated index.html landing page listing all lessons grouped by phase.
package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
)

//go:embed index.html.tmpl
var indexFS embed.FS

// lessonInfo is one entry in the master lesson list. The master list is
// authoritative for titles and blurbs; whether a lesson is published is
// determined at build time by checking whether the lesson's directory exists
// on disk.
type lessonInfo struct {
	Number string // "01"
	Slug   string // "hello"
	Title  string // "Hello, Go"
	Blurb  string // "go run · package main · fmt"
	Phase  int    // 1-4
}

// allLessons is the master list of all 29 lessons across the four phases.
// New lessons are added here when they're planned (so they appear as faded
// placeholders on the landing page until their content lands on disk).
var allLessons = []lessonInfo{
	// Phase 1 — Foundations
	{Number: "01", Slug: "hello", Title: "Hello, Go", Blurb: "go run · package main · fmt", Phase: 1},
	{Number: "02", Slug: "variables", Title: "Variables, types, operators", Blurb: "var · := · float64 · const", Phase: 1},
	{Number: "03", Slug: "control-flow", Title: "Control flow", Blurb: "if · for · switch", Phase: 1},
	{Number: "04", Slug: "functions", Title: "Functions & first tests", Blurb: "multi-return · table tests", Phase: 1},
	{Number: "05", Slug: "slices-maps", Title: "Slices and maps", Blurb: "[]T · map[K]V · range", Phase: 1},
	{Number: "06", Slug: "structs", Title: "Structs & methods", Blurb: "type T struct · methods", Phase: 1},
	{Number: "07", Slug: "packages", Title: "Packages & modules", Blurb: "go.mod · imports · gofmt", Phase: 1},
	{Number: "08", Slug: "capstone", Title: "Phase 1 capstone", Blurb: "expense tracker CLI", Phase: 1},
	// Phase 2 — Idiomatic Go
	{Number: "09", Slug: "pointers", Title: "Pointers", Blurb: "value vs reference", Phase: 2},
	{Number: "10", Slug: "interfaces", Title: "Interfaces", Blurb: "io.Reader · any", Phase: 2},
	{Number: "11", Slug: "errors", Title: "Errors", Blurb: "wrapping · errors.Is/As", Phase: 2},
	{Number: "12", Slug: "generics", Title: "Generics", Blurb: "type parameters", Phase: 2},
	{Number: "13", Slug: "encoding-io", Title: "Encoding & I/O", Blurb: "JSON · bufio · streams", Phase: 2},
	{Number: "14", Slug: "time-strings-regex", Title: "Time, strings, regex", Blurb: "stdlib literacy", Phase: 2},
	{Number: "15", Slug: "structure", Title: "Project structure", Blurb: "cmd/ · internal/", Phase: 2},
	// Phase 3 — Concurrency & Systems
	{Number: "16", Slug: "goroutines-channels", Title: "Goroutines & channels", Blurb: "go · chan · range", Phase: 3},
	{Number: "17", Slug: "select-timers", Title: "Select & timers", Blurb: "select · time.After", Phase: 3},
	{Number: "18", Slug: "sync", Title: "sync & memory model", Blurb: "Mutex · race detector", Phase: 3},
	{Number: "19", Slug: "context", Title: "context", Blurb: "cancellation · deadlines", Phase: 3},
	{Number: "20", Slug: "patterns", Title: "Concurrency patterns", Blurb: "worker pool · errgroup", Phase: 3},
	{Number: "21", Slug: "networking", Title: "Networking", Blurb: "net · TCP · syscalls", Phase: 3},
	{Number: "22", Slug: "profiling", Title: "Profiling & benchmarks", Blurb: "pprof · go test -bench", Phase: 3},
	// Phase 4 — Production & Distributed
	{Number: "23", Slug: "http-server", Title: "HTTP servers", Blurb: "net/http · slog", Phase: 4},
	{Number: "24", Slug: "http-client", Title: "HTTP clients", Blurb: "retries · timeouts", Phase: 4},
	{Number: "25", Slug: "grpc", Title: "gRPC", Blurb: "protobuf · streaming", Phase: 4},
	{Number: "26", Slug: "config", Title: "Config & shutdown", Blurb: "flags · env · signals", Phase: 4},
	{Number: "27", Slug: "container", Title: "Containerization", Blurb: "multi-stage · distroless", Phase: 4},
	{Number: "28", Slug: "observability", Title: "Observability", Blurb: "metrics · traces", Phase: 4},
	{Number: "29", Slug: "capstone-final", Title: "Course capstone", Blurb: "distributed wrap-up", Phase: 4},
}

// phaseInfo names the four phases for rendering.
type phaseInfo struct {
	Number int
	Name   string
}

var phases = []phaseInfo{
	{Number: 1, Name: "Foundations"},
	{Number: 2, Name: "Idiomatic Go"},
	{Number: 3, Name: "Concurrency & Systems"},
	{Number: 4, Name: "Production & Distributed"},
}

// lessonView is what the template iterates over.
type lessonView struct {
	Number    string
	Title     string
	Blurb     string
	Name      string // "01-hello" — used for slides/ URL when Published
	Published bool
}

// phaseView is a phase plus its lessons.
type phaseView struct {
	Number  int
	Name    string
	Lessons []lessonView
}

// indexData is the root template payload.
type indexData struct {
	Phases []phaseView
}

// collectPublishedSlugs walks lessonsDir to find which "NN-slug" directories
// exist on disk. Returns a set keyed by full lesson name (e.g. "01-hello").
func collectPublishedSlugs(lessonsDir string) (map[string]bool, error) {
	entries, err := os.ReadDir(lessonsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	out := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Confirm a slides/ subdir exists; otherwise it isn't really published.
		if _, err := os.Stat(filepath.Join(lessonsDir, e.Name(), "slides")); err != nil {
			continue
		}
		out[e.Name()] = true
	}
	return out, nil
}

// buildPhases returns the four phases populated with lessons from allLessons,
// each lesson marked Published if its directory exists in lessonsDir.
func buildPhases(lessonsDir string) ([]phaseView, error) {
	published, err := collectPublishedSlugs(lessonsDir)
	if err != nil {
		return nil, err
	}
	byPhase := map[int][]lessonView{}
	for _, l := range allLessons {
		name := l.Number + "-" + l.Slug
		byPhase[l.Phase] = append(byPhase[l.Phase], lessonView{
			Number:    l.Number,
			Title:     l.Title,
			Blurb:     l.Blurb,
			Name:      name,
			Published: published[name],
		})
	}
	// Sort each phase's lessons by Number (string sort works because they're
	// all two-digit zero-padded).
	for k := range byPhase {
		sort.Slice(byPhase[k], func(i, j int) bool {
			return byPhase[k][i].Number < byPhase[k][j].Number
		})
	}
	out := make([]phaseView, 0, len(phases))
	for _, p := range phases {
		out = append(out, phaseView{
			Number:  p.Number,
			Name:    p.Name,
			Lessons: byPhase[p.Number],
		})
	}
	return out, nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close() //nolint:errcheck
		return err
	}
	return out.Close()
}

func renderIndex(data indexData) ([]byte, error) {
	tmpl, err := template.ParseFS(indexFS, "index.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

func build(lessonsDir, sharedDir, outDir string) error {
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clear out dir: %w", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	phases, err := buildPhases(lessonsDir)
	if err != nil {
		return fmt.Errorf("build phases: %w", err)
	}
	// Copy every published lesson's slides.
	for _, p := range phases {
		for _, l := range p.Lessons {
			if !l.Published {
				continue
			}
			src := filepath.Join(lessonsDir, l.Name, "slides")
			dst := filepath.Join(outDir, "lessons", l.Name, "slides")
			if err := copyTree(src, dst); err != nil {
				return fmt.Errorf("copy %s slides: %w", l.Name, err)
			}
		}
	}
	// Copy shared/reveal.
	if _, err := os.Stat(sharedDir); err == nil {
		if err := copyTree(sharedDir, filepath.Join(outDir, "shared", "reveal")); err != nil {
			return fmt.Errorf("copy shared/reveal: %w", err)
		}
	}
	idx, err := renderIndex(indexData{Phases: phases})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "index.html"), idx, 0o644)
}

func main() {
	os.Exit(runWithArgs(os.Args))
}

func runWithArgs(args []string) int {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	lessonsDir := flags.String("lessons", "lessons", "directory containing lesson folders")
	sharedDir := flags.String("shared", "shared/reveal", "directory containing shared reveal.js assets")
	outDir := flags.String("out", "dist", "output directory")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if err := build(*lessonsDir, *sharedDir, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	fmt.Printf("built %s\n", *outDir)
	return 0
}
