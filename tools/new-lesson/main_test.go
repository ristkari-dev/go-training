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
		want := lessonInfo{Number: "01", Slug: "hello", Name: "01-hello", Title: "Hello"}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("valid multi-word", func(t *testing.T) {
		got, err := parseName("05-slices-and-maps")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := lessonInfo{Number: "05", Slug: "slices-and-maps", Name: "05-slices-and-maps", Title: "Slices And Maps"}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
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
		"01-hello/exercises/main.go",
		"01-hello/exercises/main_test.go",
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
