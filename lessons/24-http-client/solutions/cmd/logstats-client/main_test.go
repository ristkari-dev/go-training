package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accepted":3,"parsed":2,"failed":1}`))
	}))
	defer srv.Close()

	var out bytes.Buffer
	in := strings.NewReader("2026-01-02T15:04:05 INFO ok\n2026-01-02T15:04:06 ERROR boom\nbad\n")
	if err := run(context.Background(), srv.URL, in, &out); err != nil {
		t.Fatalf("run = %v", err)
	}
	if !strings.Contains(out.String(), "accepted=3 parsed=2 failed=1") {
		t.Errorf("out = %q", out.String())
	}
}
