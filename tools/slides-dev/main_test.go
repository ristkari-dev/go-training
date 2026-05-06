package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "lessons/01-hello/slides/slides.md"), "# Hello\n")
	mustWrite(t, filepath.Join(dir, "lessons/01-hello/slides/index.html"), "<html>hi</html>")
	mustWrite(t, filepath.Join(dir, "shared/reveal/dist/reveal.js"), "console.log('reveal');")
	return dir
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHandlerServesLessonSlides(t *testing.T) {
	repo := setupRepo(t)
	h, err := buildHandler(repo, "01-hello")
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/lessons/01-hello/slides/slides.md", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Hello") {
		t.Errorf("body: %q does not contain 'Hello'", rec.Body.String())
	}
}

func TestHandlerServesRevealAssets(t *testing.T) {
	repo := setupRepo(t)
	h, err := buildHandler(repo, "01-hello")
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/shared/reveal/dist/reveal.js", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
}

func TestHandlerRedirectsRoot(t *testing.T) {
	repo := setupRepo(t)
	h, err := buildHandler(repo, "01-hello")
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status: got %d, want 302", rec.Code)
	}
	want := "/lessons/01-hello/slides/"
	if got := rec.Header().Get("Location"); got != want {
		t.Errorf("Location: got %q, want %q", got, want)
	}
}

func TestHandlerErrorsOnMissingLesson(t *testing.T) {
	repo := setupRepo(t)
	if _, err := buildHandler(repo, "99-nope"); err == nil {
		t.Error("expected error for missing lesson")
	}
}

func TestHandlerErrorsOnMissingReveal(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "lessons/01-hello/slides/slides.md"), "# Hi")
	if _, err := buildHandler(dir, "01-hello"); err == nil {
		t.Error("expected error when shared/reveal is absent")
	}
}
