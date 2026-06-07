// Package httpsrv builds the logstats HTTP router — the service's HTTP
// face, extracted so the unified daemon can mount it alongside gRPC.
package httpsrv

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/ristkari-dev/go-training/lessons/28-observability/exercises/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/28-observability/exercises/internal/logstats"
	"github.com/ristkari-dev/go-training/lessons/28-observability/exercises/internal/otelx"
)

const maxIngestBytes = 1 << 20

type ingestRequest struct {
	Lines []string `json:"lines"`
}

type ingestResponse struct {
	Accepted int `json:"accepted"`
	Parsed   int `json:"parsed"`
	Failed   int `json:"failed"`
}

type statsResponse struct {
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
}

// Router returns the HTTP handler for the logstats service (POST /ingest,
// GET /stats, GET /healthz), wrapped in OpenTelemetry instrumentation
// (span + request metrics + trace-correlated logging) and panic recovery.
func Router(store *logstats.Store, logger *slog.Logger) http.Handler {
	// ingested_lines_total: a metric recorded where the domain event
	// happens (parsing). Created once; the MeterProvider must already be
	// set (the daemon calls otelx.Setup before building the router).
	ingested, _ := otel.Meter("logstatsd").Int64Counter("ingested_lines_total")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", ingestHandler(store, ingested))
	mux.HandleFunc("GET /stats", statsHandler(store))
	mux.HandleFunc("GET /healthz", healthHandler)
	return otelx.Instrument(withRecovery(mux, logger), logger)
}

func ingestHandler(store *logstats.Store, ingested metric.Int64Counter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ingestRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxIngestBytes))
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		delta := map[string]int{}
		parsed, failed := 0, 0
		for _, line := range req.Lines {
			e, err := logparse.ParseLine(line)
			if err != nil {
				failed++
				continue
			}
			delta[e.Level]++
			parsed++
		}
		store.Merge(delta)
		ingested.Add(r.Context(), int64(parsed+failed))
		writeJSON(w, http.StatusOK, ingestResponse{
			Accepted: len(req.Lines), Parsed: parsed, Failed: failed,
		})
	}
}

func statsHandler(store *logstats.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		counts, total := store.Snapshot()
		writeJSON(w, http.StatusOK, statsResponse{Counts: counts, Total: total})
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withRecovery(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This yields a clean 500 only if the handler hasn't written a
		// response yet — once WriteHeader is called the status is
		// committed and can't be changed. Our handlers do all fallible
		// work (decode, parse) BEFORE writing, so a panic lands here
		// before any bytes go out.
		defer func() {
			if v := recover(); v != nil {
				logger.Error("panic recovered", "value", v, "path", r.URL.Path)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
