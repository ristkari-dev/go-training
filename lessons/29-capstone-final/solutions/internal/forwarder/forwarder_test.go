package forwarder

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/httpsrv"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/logstats"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/warmup/dedup"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestRetriesTransient(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	f := New(srv.URL+"/ingest", "fwd1")
	if err := f.Send(context.Background(), []string{"2026-01-02T15:04:05 INFO a"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

// TestEffectivelyOnce — the capstone. The aggregator PROCESSES a batch
// but its ack is "lost" (the wrapper returns 503 the first time it sees
// a key, replaying the real response on the retry). The forwarder
// retries the SAME key → the aggregator dedups → lines counted ONCE.
func TestEffectivelyOnce(t *testing.T) {
	store := logstats.NewStore()
	real := httpsrv.Router(store, dedup.New(1000), discardLogger())

	var mu sync.Mutex
	failed := map[string]bool{}
	agg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		rec := httptest.NewRecorder()
		real.ServeHTTP(rec, r) // real handler merges + records the key (first time)
		mu.Lock()
		first := !failed[key]
		failed[key] = true
		mu.Unlock()
		if first {
			w.WriteHeader(http.StatusServiceUnavailable) // "lose the ack"
			return
		}
		for k, vs := range rec.Header() { // replay the real response on retry
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.Code)
		_, _ = w.Write(rec.Body.Bytes())
	}))
	defer agg.Close()

	f := New(agg.URL+"/ingest", "fwd1")
	if err := f.Send(context.Background(), []string{
		"2026-01-02T15:04:05 INFO a", "2026-01-02T15:04:06 WARN b",
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if _, total := store.Snapshot(); total != 2 {
		t.Errorf("total = %d, want 2 (effectively-once despite redelivery)", total)
	}
}
