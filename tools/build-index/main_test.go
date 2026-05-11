package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAllLessonsCovers29Lessons(t *testing.T) {
	if len(allLessons) != 29 {
		t.Errorf("allLessons has %d entries, want 29", len(allLessons))
	}
	seen := map[string]bool{}
	for _, l := range allLessons {
		if seen[l.Number] {
			t.Errorf("duplicate lesson number %q", l.Number)
		}
		seen[l.Number] = true
		if l.Phase < 1 || l.Phase > 4 {
			t.Errorf("lesson %s: phase %d outside 1-4", l.Number, l.Phase)
		}
		if l.Title == "" || l.Slug == "" || l.Blurb == "" {
			t.Errorf("lesson %s: empty Title/Slug/Blurb", l.Number)
		}
	}
}

func TestCollectPublishedSlugs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "01-hello", "slides", "slides.md"), "# hi")
	writeFile(t, filepath.Join(dir, "02-variables", "slides", "index.html"), "<html></html>")
	// A dir without slides/ — should be ignored.
	if err := os.MkdirAll(filepath.Join(dir, "03-broken"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := collectPublishedSlugs(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"01-hello": true, "02-variables": true}
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for k := range want {
		if !got[k] {
			t.Errorf("missing %q", k)
		}
	}
	if got["03-broken"] {
		t.Errorf("03-broken should not appear (no slides/ subdir)")
	}
}

func TestCollectPublishedSlugsHandlesMissingDir(t *testing.T) {
	got, err := collectPublishedSlugs(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d, want 0", len(got))
	}
}

func TestBuildPhases(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "01-hello", "slides", "slides.md"), "# hi")
	writeFile(t, filepath.Join(dir, "02-variables", "slides", "slides.md"), "# hi")

	phases, err := buildPhases(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(phases) != 4 {
		t.Fatalf("got %d phases, want 4", len(phases))
	}
	// Phase 1 should have 8 lessons.
	if len(phases[0].Lessons) != 8 {
		t.Errorf("phase 1 has %d lessons, want 8", len(phases[0].Lessons))
	}
	// Phase 1 lessons should be sorted by Number.
	for i, want := range []string{"01", "02", "03", "04", "05", "06", "07", "08"} {
		if phases[0].Lessons[i].Number != want {
			t.Errorf("phase 1 lesson %d: got %q, want %q", i, phases[0].Lessons[i].Number, want)
		}
	}
	// Published flags.
	if !phases[0].Lessons[0].Published {
		t.Error("01-hello should be Published")
	}
	if !phases[0].Lessons[1].Published {
		t.Error("02-variables should be Published")
	}
	if phases[0].Lessons[2].Published {
		t.Error("03 should NOT be Published")
	}
}

func TestRenderIndexIncludesAllPhases(t *testing.T) {
	phases, err := buildPhases(t.TempDir()) // empty dir; everything unpublished
	if err != nil {
		t.Fatal(err)
	}
	body, err := renderIndex(indexData{Phases: phases})
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, want := range []string{"Foundations", "Idiomatic Go", "Concurrency", "Production"} {
		if !strings.Contains(s, want) {
			t.Errorf("index missing phase %q", want)
		}
	}
	// Lesson titles from allLessons appear.
	for _, want := range []string{"Hello, Go", "Variables, types, operators", "Goroutines"} {
		if !strings.Contains(s, want) {
			t.Errorf("index missing lesson %q", want)
		}
	}
}

func TestCopyTreeMirrorsStructure(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "a", "b", "c.txt"), "hello")
	writeFile(t, filepath.Join(src, "a", "d.txt"), "world")
	if err := copyTree(src, dst); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dst, "a", "b", "c.txt"))
	if err != nil || string(body) != "hello" {
		t.Errorf("a/b/c.txt: got %q, err %v", body, err)
	}
}

func TestBuildEndToEnd(t *testing.T) {
	root := t.TempDir()
	lessonsDir := filepath.Join(root, "lessons")
	sharedDir := filepath.Join(root, "shared", "reveal")
	outDir := filepath.Join(root, "dist")

	writeFile(t, filepath.Join(lessonsDir, "01-hello", "slides", "index.html"), "<html>hi</html>")
	writeFile(t, filepath.Join(lessonsDir, "01-hello", "slides", "slides.md"), "# Hello\n")
	writeFile(t, filepath.Join(sharedDir, "dist", "reveal.js"), "console.log('reveal');")

	if err := build(lessonsDir, sharedDir, outDir); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{
		"index.html",
		"lessons/01-hello/slides/index.html",
		"lessons/01-hello/slides/slides.md",
		"shared/reveal/dist/reveal.js",
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
	idx, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(idx), "Hello, Go") {
		t.Errorf("index.html missing 'Hello, Go'\n%s", idx)
	}
}
