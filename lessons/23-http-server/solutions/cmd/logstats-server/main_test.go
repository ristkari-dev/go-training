package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ristkari-dev/go-training/lessons/23-http-server/solutions/internal/logstats"
)

func testRouter() http.Handler {
	return newRouter(logstats.NewStore(), slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func TestIngestThenStats(t *testing.T) {
	srv := httptest.NewServer(testRouter())
	defer srv.Close()

	body := `{"lines":[
		"2026-01-02T15:04:05 INFO ok",
		"2026-01-02T15:04:06 INFO again",
		"2026-01-02T15:04:07 WARN slow",
		"garbage line"
	]}`
	resp, err := http.Post(srv.URL+"/ingest", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ingest status = %d", resp.StatusCode)
	}
	var ir ingestResponse
	_ = json.NewDecoder(resp.Body).Decode(&ir)
	resp.Body.Close()
	if ir.Accepted != 4 || ir.Parsed != 3 || ir.Failed != 1 {
		t.Errorf("ingest resp = %+v, want accepted=4 parsed=3 failed=1", ir)
	}

	resp2, err := http.Get(srv.URL + "/stats")
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	var sr statsResponse
	_ = json.NewDecoder(resp2.Body).Decode(&sr)
	resp2.Body.Close()
	if sr.Total != 3 || sr.Counts["INFO"] != 2 || sr.Counts["WARN"] != 1 {
		t.Errorf("stats = %+v, want total=3 INFO=2 WARN=1", sr)
	}
}

func TestIngestBadJSON(t *testing.T) {
	srv := httptest.NewServer(testRouter())
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/ingest", "application/json", strings.NewReader("{not json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("healthz = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, httptest.NewRequest("GET", "/ingest", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /ingest = %d, want 405", rec.Code)
	}
}

func TestConcurrentIngestRace(t *testing.T) {
	srv := httptest.NewServer(testRouter())
	defer srv.Close()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Post(srv.URL+"/ingest", "application/json",
				strings.NewReader(`{"lines":["2026-01-02T15:04:05 ERROR e"]}`))
			if err == nil {
				resp.Body.Close()
			}
		}()
	}
	wg.Wait()
	resp, _ := http.Get(srv.URL + "/stats")
	var sr statsResponse
	_ = json.NewDecoder(resp.Body).Decode(&sr)
	resp.Body.Close()
	if sr.Counts["ERROR"] != 50 {
		t.Errorf("ERROR = %d, want 50", sr.Counts["ERROR"])
	}
}

func TestServeLifecycle(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- serve(ctx, ln, io.Discard) }()

	url := fmt.Sprintf("http://%s/healthz", ln.Addr().String())
	ready := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if resp, err := http.Get(url); err == nil {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server never became ready")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("serve returned %v, want nil", err)
		}
	case <-time.After(10 * time.Second):
		// Generous ceiling: serve waits on srv.Shutdown (up to 5s), so
		// the test timeout must exceed it. With zero in-flight requests
		// Shutdown returns near-instantly.
		t.Fatal("serve did not return after cancel")
	}
}
