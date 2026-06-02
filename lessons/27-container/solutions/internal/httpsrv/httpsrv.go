// Package httpsrv builds the logstats HTTP router — the service's HTTP
// face, extracted so the unified daemon can mount it alongside gRPC.
package httpsrv

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ristkari-dev/go-training/lessons/27-container/solutions/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/27-container/solutions/internal/logstats"
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
// GET /stats, GET /healthz), wrapped in logging + recovery middleware.
func Router(store *logstats.Store, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", ingestHandler(store))
	mux.HandleFunc("GET /stats", statsHandler(store))
	mux.HandleFunc("GET /healthz", healthHandler)
	return withRequestLog(withRecovery(mux, logger), logger)
}

func ingestHandler(store *logstats.Store) http.HandlerFunc {
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

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func withRequestLog(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration_ms", time.Since(start).Milliseconds())
	})
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
