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

func TestExtractTitleFromReadme(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, "# Lesson 03: Control Flow\n\nSome content.\n")
	if got := extractTitle(readme, "control-flow"); got != "Control Flow" {
		t.Errorf("got %q, want %q", got, "Control Flow")
	}
}

func TestExtractTitleFallsBackToSlug(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, "Some text without an H1.\n")
	if got := extractTitle(readme, "slices-and-maps"); got != "Slices And Maps" {
		t.Errorf("got %q, want %q", got, "Slices And Maps")
	}
}

func TestExtractTitleHandlesMissingFile(t *testing.T) {
	if got := extractTitle("/nonexistent/path/README.md", "hello"); got != "Hello" {
		t.Errorf("got %q, want %q", got, "Hello")
	}
}

func TestCollectLessonsSortedByNumber(t *testing.T) {
	dir := t.TempDir()
	for _, l := range []struct{ name, body string }{
		{"03-control-flow", "# Lesson 03: Control Flow\n"},
		{"01-hello", "# Lesson 01: Hello\n"},
		{"10-modules", "# Lesson 10: Modules\n"},
	} {
		writeFile(t, filepath.Join(dir, l.name, "README.md"), l.body)
		writeFile(t, filepath.Join(dir, l.name, "slides", "index.html"), "<html></html>")
	}
	got, err := collectLessons(dir)
	if err != nil {
		t.Fatal(err)
	}
	wantNames := []string{"01-hello", "03-control-flow", "10-modules"}
	if len(got) != len(wantNames) {
		t.Fatalf("got %d lessons, want %d", len(got), len(wantNames))
	}
	for i, w := range wantNames {
		if got[i].Name != w {
			t.Errorf("lesson %d: got %q, want %q", i, got[i].Name, w)
		}
	}
	if got[0].Title != "Hello" {
		t.Errorf("lesson 0 title: got %q, want %q", got[0].Title, "Hello")
	}
}

func TestCollectLessonsIgnoresEntriesWithoutSlides(t *testing.T) {
	dir := t.TempDir()
	// A lesson dir with no slides/ should be skipped.
	writeFile(t, filepath.Join(dir, "01-hello", "README.md"), "# Lesson 01: Hello\n")
	// A lesson dir with slides/.
	writeFile(t, filepath.Join(dir, "02-types", "README.md"), "# Lesson 02: Types\n")
	writeFile(t, filepath.Join(dir, "02-types", "slides", "index.html"), "<html></html>")
	got, err := collectLessons(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "02-types" {
		t.Errorf("got %+v, want only 02-types", got)
	}
}

func TestCollectLessonsHandlesMissingDir(t *testing.T) {
	got, err := collectLessons(filepath.Join(t.TempDir(), "lessons-doesnt-exist"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d lessons, want 0", len(got))
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
	for _, rel := range []string{"a/b/c.txt", "a/d.txt"} {
		body, err := os.ReadFile(filepath.Join(dst, rel))
		if err != nil {
			t.Errorf("%s: %v", rel, err)
		}
		if rel == "a/b/c.txt" && string(body) != "hello" {
			t.Errorf("%s body mismatch: %q", rel, body)
		}
	}
}

func TestRenderIndexLists(t *testing.T) {
	body, err := renderIndex([]lesson{
		{Number: "01", Slug: "hello", Name: "01-hello", Title: "Hello"},
		{Number: "02", Slug: "types", Name: "02-types", Title: "Types"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, want := range []string{"Hello", "Types", "lessons/01-hello/slides/", "lessons/02-types/slides/"} {
		if !strings.Contains(s, want) {
			t.Errorf("index missing %q\n---\n%s", want, s)
		}
	}
}

func TestRenderIndexEmpty(t *testing.T) {
	body, err := renderIndex(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "No lessons published yet.") {
		t.Errorf("empty index missing fallback text\n---\n%s", body)
	}
}

func TestBuildEndToEnd(t *testing.T) {
	root := t.TempDir()
	lessonsDir := filepath.Join(root, "lessons")
	sharedDir := filepath.Join(root, "shared", "reveal")
	outDir := filepath.Join(root, "dist")

	writeFile(t, filepath.Join(lessonsDir, "01-hello", "README.md"), "# Lesson 01: Hello\n")
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
	if !strings.Contains(string(idx), "Hello") {
		t.Errorf("index.html missing 'Hello'\n%s", idx)
	}
}
