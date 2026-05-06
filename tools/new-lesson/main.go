// Command new-lesson scaffolds a new go-training lesson folder from
// the embedded template under tools/new-lesson/template.
package main

import (
	"fmt"
	"regexp"
	"strings"
)

type lessonInfo struct {
	Number string // "01"
	Slug   string // "hello"
	Name   string // "01-hello"
	Title  string // "Hello"
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

func main() {
	// Wired in Task 7.
}
