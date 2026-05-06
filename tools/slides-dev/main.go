// Command slides-dev serves a single lesson's reveal.js deck over HTTP for local
// development. It mounts the repo's lessons/ and shared/ trees so relative paths
// inside the lesson's index.html resolve the same way they do on disk.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// buildHandler returns an http.Handler that serves lessons and shared assets
// from the repo, and redirects "/" to the requested lesson's deck.
func buildHandler(repo, lesson string) (http.Handler, error) {
	slidesDir := filepath.Join(repo, "lessons", lesson, "slides")
	if _, err := os.Stat(slidesDir); err != nil {
		return nil, fmt.Errorf("lesson slides not found at %s: %w", slidesDir, err)
	}
	revealDir := filepath.Join(repo, "shared", "reveal")
	if _, err := os.Stat(revealDir); err != nil {
		return nil, fmt.Errorf("shared reveal not found at %s: %w", revealDir, err)
	}

	deckURL := "/lessons/" + lesson + "/slides/"
	repoFS := http.FileServer(http.Dir(repo))

	mux := http.NewServeMux()
	mux.Handle("/lessons/", repoFS)
	mux.Handle("/shared/", repoFS)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, deckURL, http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	return mux, nil
}

func main() {
	os.Exit(runWithArgs(os.Args))
}

func runWithArgs(args []string) int {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	lesson := flags.String("lesson", "", "lesson name (e.g. 01-hello)")
	addr := flags.String("addr", ":8000", "listen address")
	repo := flags.String("repo", ".", "repo root containing lessons/ and shared/")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if *lesson == "" {
		fmt.Fprintln(os.Stderr, "error: -lesson is required (e.g. -lesson 01-hello)")
		return 2
	}
	h, err := buildHandler(*repo, *lesson)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	displayAddr := *addr
	if len(displayAddr) > 0 && displayAddr[0] == ':' {
		displayAddr = "localhost" + displayAddr
	}
	fmt.Printf("serving lesson %s on http://%s/  (Ctrl-C to stop)\n", *lesson, displayAddr)
	if err := http.ListenAndServe(*addr, h); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}
