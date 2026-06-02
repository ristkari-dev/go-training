package httpsrv

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/internal/logstats"
)

func testRouter() http.Handler {
	return Router(logstats.NewStore(), slog.New(slog.NewJSONHandler(io.Discard, nil)))
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
