package main

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/httpsrv"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/logstats"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/warmup/dedup"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// TestRun forwards two stdin lines to an in-process aggregator and asserts
// the call succeeds (the binary's run wires stdin → forwarder → /ingest).
func TestRun(t *testing.T) {
	srv := httptest.NewServer(httpsrv.Router(logstats.NewStore(), dedup.New(1000), discardLogger()))
	defer srv.Close()

	in := strings.NewReader("line1\nline2\n")
	if err := run(context.Background(), srv.URL+"/ingest", "fwd1", in, io.Discard); err != nil {
		t.Fatalf("run: %v", err)
	}
}
