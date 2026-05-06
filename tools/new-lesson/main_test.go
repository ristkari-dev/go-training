package main

import (
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
