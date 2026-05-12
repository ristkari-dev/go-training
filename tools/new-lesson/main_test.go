package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseName(t *testing.T) {
	t.Run("valid simple", func(t *testing.T) {
		got, err := parseName("01-hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := lessonInfo{Number: "01", Slug: "hello", Name: "01-hello", Title: "Hello", Phase: 1, PhaseName: "Foundations"}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("valid multi-word", func(t *testing.T) {
		got, err := parseName("05-slices-and-maps")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := lessonInfo{Number: "05", Slug: "slices-and-maps", Name: "05-slices-and-maps", Title: "Slices And Maps", Phase: 1, PhaseName: "Foundations"}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("phase boundaries", func(t *testing.T) {
		cases := []struct {
			name      string
			phase     int
			phaseName string
		}{
			{"08-capstone", 1, "Foundations"},
			{"09-pointers", 2, "Idiomatic Go"},
			{"15-structure", 2, "Idiomatic Go"},
			{"16-goroutines", 3, "Concurrency & Systems"},
			{"22-profiling", 3, "Concurrency & Systems"},
			{"23-http-server", 4, "Production & Distributed"},
			{"29-capstone-final", 4, "Production & Distributed"},
		}
		for _, c := range cases {
			got, err := parseName(c.name)
			if err != nil {
				t.Errorf("%s: unexpected error: %v", c.name, err)
				continue
			}
			if got.Phase != c.phase || got.PhaseName != c.phaseName {
				t.Errorf("%s: got phase=%d/%q, want %d/%q", c.name, got.Phase, got.PhaseName, c.phase, c.phaseName)
			}
		}
	})

	t.Run("rejects out-of-range lesson number", func(t *testing.T) {
		if _, err := parseName("30-extra"); err == nil {
			t.Error("expected error for lesson number outside 1-29")
		}
	})

	t.Run("rejects invalid forms", func(t *testing.T) {
		bad := []string{
			"",
			"hello",
			"1-hello",
			"001-hello",
			"01_hello",
			"01-Hello",
			"01-hello-",
			"-hello",
			"01-",
		}
		for _, name := range bad {
			if _, err := parseName(name); err == nil {
				t.Errorf("expected error for %q", name)
			}
		}
	})
}

func TestScaffoldCreatesExpectedTree(t *testing.T) {
	dest := t.TempDir()
	if err := scaffoldLesson("01-hello", dest); err != nil {
		t.Fatalf("scaffoldLesson: %v", err)
	}
	want := []string{
		"01-hello/README.md",
		"01-hello/slides/index.html",
		"01-hello/slides/slides.md",
		"01-hello/slides/assets/.gitkeep",
		"01-hello/exercises/warmup.go",
		"01-hello/exercises/warmup_test.go",
		"01-hello/exercises/main.go",
		"01-hello/exercises/main_test.go",
		"01-hello/solutions/warmup.go",
		"01-hello/solutions/warmup_test.go",
		"01-hello/solutions/main.go",
		"01-hello/solutions/main_test.go",
	}
	for _, rel := range want {
		p := filepath.Join(dest, rel)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestScaffoldSubstitutesLessonInfo(t *testing.T) {
	dest := t.TempDir()
	if err := scaffoldLesson("05-slices-and-maps", dest); err != nil {
		t.Fatalf("scaffoldLesson: %v", err)
	}
	readme, err := os.ReadFile(filepath.Join(dest, "05-slices-and-maps", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Lesson 05", "Slices And Maps", "lessons/05-slices-and-maps/exercises"} {
		if !strings.Contains(string(readme), want) {
			t.Errorf("README missing %q\n---\n%s", want, readme)
		}
	}
	slides, err := os.ReadFile(filepath.Join(dest, "05-slices-and-maps", "slides", "slides.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<div class="lesson-number">05</div>`,
		`Phase 1 — Foundations`,
		`<h1>Slices And Maps</h1>`,
		"lessons/05-slices-and-maps/exercises",
	} {
		if !strings.Contains(string(slides), want) {
			t.Errorf("slides.md missing %q\n---\n%s", want, slides)
		}
	}
}

func TestScaffoldRefusesExistingDest(t *testing.T) {
	dest := t.TempDir()
	if err := scaffoldLesson("01-hello", dest); err != nil {
		t.Fatalf("first scaffoldLesson: %v", err)
	}
	err := scaffoldLesson("01-hello", dest)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected already-exists error, got %v", err)
	}
}

func TestCLIRejectsMissingName(t *testing.T) {
	if code := runWithArgs([]string{"new-lesson"}); code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
}

func TestCLIScaffoldsHappyPath(t *testing.T) {
	dest := t.TempDir()
	code := runWithArgs([]string{"new-lesson", "-name", "07-packages", "-out", dest})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if _, err := os.Stat(filepath.Join(dest, "07-packages", "README.md")); err != nil {
		t.Fatalf("README not created: %v", err)
	}
}
