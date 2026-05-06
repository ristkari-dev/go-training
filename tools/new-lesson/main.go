// Command new-lesson scaffolds a new go-training lesson folder from
// the embedded template under tools/new-lesson/template.
package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed all:template
var templateFS embed.FS

type lessonInfo struct {
	Number string
	Slug   string
	Name   string
	Title  string
}

var nameRe = regexp.MustCompile(`^(\d{2})-([a-z][a-z0-9]*(?:-[a-z0-9]+)*)$`)

func parseName(name string) (lessonInfo, error) {
	m := nameRe.FindStringSubmatch(name)
	if m == nil {
		return lessonInfo{}, fmt.Errorf("invalid lesson name %q: must match NN-kebab-case, e.g. 01-hello", name)
	}
	return lessonInfo{
		Number: m[1],
		Slug:   m[2],
		Name:   name,
		Title:  toTitle(m[2]),
	}, nil
}

func toTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// scaffoldLesson creates a new lesson under outBase using the embedded template.
// It refuses to overwrite an existing destination.
func scaffoldLesson(name, outBase string) error {
	info, err := parseName(name)
	if err != nil {
		return err
	}
	dest := filepath.Join(outBase, info.Name)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("destination already exists: %s", dest)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return walkTemplate(templateFS, "template", dest, info)
}

func walkTemplate(srcFS fs.FS, root, dest string, info lessonInfo) error {
	return fs.WalkDir(srcFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dest, 0o755)
		}
		target := filepath.Join(dest, strings.TrimSuffix(rel, ".tmpl"))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(srcFS, path)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return os.WriteFile(target, data, 0o644)
		}
		tmpl, err := template.New(rel).Parse(string(data))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", rel, err)
		}
		f, err := os.Create(target)
		if err != nil {
			return err
		}
		defer f.Close()
		return tmpl.Execute(f, info)
	})
}

func main() {
	// Wired in Task 7.
}
