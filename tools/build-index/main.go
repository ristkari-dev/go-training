// Command build-index produces a static slides site under dist/: every lesson's
// slides copied into place, the shared reveal.js assets copied alongside, and a
// generated index.html landing page listing all lessons.
package main

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

//go:embed index.html.tmpl
var indexFS embed.FS

type lesson struct {
	Number string
	Slug   string
	Name   string
	Title  string
}

var lessonNameRe = regexp.MustCompile(`^(\d{2})-([a-z][a-z0-9]*(?:-[a-z0-9]+)*)$`)
var titleHeadingRe = regexp.MustCompile(`(?i)^#\s+lesson\s+\d+\s*[:\-]\s*(.+?)\s*$`)

func collectLessons(lessonsDir string) ([]lesson, error) {
	entries, err := os.ReadDir(lessonsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []lesson
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m := lessonNameRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		slidesDir := filepath.Join(lessonsDir, e.Name(), "slides")
		if info, err := os.Stat(slidesDir); err != nil || !info.IsDir() {
			continue
		}
		readme := filepath.Join(lessonsDir, e.Name(), "README.md")
		out = append(out, lesson{
			Number: m[1],
			Slug:   m[2],
			Name:   e.Name(),
			Title:  extractTitle(readme, m[2]),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func extractTitle(readmePath, slug string) string {
	f, err := os.Open(readmePath)
	if err != nil {
		return slugToTitle(slug)
	}
	defer f.Close() //nolint:errcheck
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := titleHeadingRe.FindStringSubmatch(scanner.Text()); m != nil {
			return strings.TrimSpace(m[1])
		}
	}
	return slugToTitle(slug)
}

func slugToTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
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

func renderIndex(lessons []lesson) ([]byte, error) {
	tmpl, err := template.ParseFS(indexFS, "index.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ Lessons []lesson }{lessons}); err != nil {
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
	lessons, err := collectLessons(lessonsDir)
	if err != nil {
		return fmt.Errorf("collect lessons: %w", err)
	}
	for _, l := range lessons {
		src := filepath.Join(lessonsDir, l.Name, "slides")
		dst := filepath.Join(outDir, "lessons", l.Name, "slides")
		if err := copyTree(src, dst); err != nil {
			return fmt.Errorf("copy %s slides: %w", l.Name, err)
		}
	}
	if _, err := os.Stat(sharedDir); err == nil {
		if err := copyTree(sharedDir, filepath.Join(outDir, "shared", "reveal")); err != nil {
			return fmt.Errorf("copy shared/reveal: %w", err)
		}
	}
	idx, err := renderIndex(lessons)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "index.html"), idx, 0o644)
}

func main() {
	// Wired in Task 3.
}
