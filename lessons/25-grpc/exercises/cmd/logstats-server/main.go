// Package main is the lesson 25 logstats HTTP service.
//
// Endpoints:
//
//	POST /ingest   {"lines":["<log line>", ...]}  → parse + accumulate
//	GET  /stats    → {"counts":{...},"total":N}
//	GET  /healthz  → {"status":"ok"}
//
// You implement newRouter + the three handlers. The server lifecycle
// (run/serve), middleware, and writeJSON are provided.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/internal/logstats"
)

const maxIngestBytes = 1 << 20 // 1 MiB request-body cap

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

// newRouter builds the mux + middleware. IMPLEMENT THIS.
//
// Hint:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("POST /ingest", ingestHandler(store))
//	mux.HandleFunc("GET /stats", statsHandler(store))
//	mux.HandleFunc("GET /healthz", healthHandler)
//	return withRequestLog(withRecovery(mux, logger), logger)
func newRouter(store *logstats.Store, logger *slog.Logger) http.Handler {
	_ = store
	_ = logger
	panic("TODO: build mux with POST /ingest, GET /stats, GET /healthz; wrap in middleware")
}

// ingestHandler decodes {"lines":[...]}, parses each line with
// logparse.ParseLine (lenient: count failures, don't reject), merges
// valid levels into the store, and returns the accepted/parsed/failed
// breakdown. Bad JSON or an oversized body → 400. IMPLEMENT THIS.
func ingestHandler(store *logstats.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = maxIngestBytes
		_ = logparse.ParseLine
		panic("TODO: decode (MaxBytesReader), ParseLine each, store.Merge, writeJSON breakdown")
	}
}

// statsHandler returns the current counts + total. IMPLEMENT THIS.
func statsHandler(store *logstats.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("TODO: store.Snapshot(); writeJSON statsResponse")
	}
}

// healthHandler reports liveness. IMPLEMENT THIS.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	panic("TODO: writeJSON 200 {\"status\":\"ok\"}")
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

func run(ctx context.Context, addr string, stdout io.Writer) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return serve(ctx, ln, stdout)
}

func serve(ctx context.Context, ln net.Listener, stdout io.Writer) error {
	logger := slog.New(slog.NewJSONHandler(stdout, nil))
	store := logstats.NewStore()
	srv := &http.Server{Handler: newRouter(store, logger)}

	// Tie the shutdown goroutine to serve's lifetime: serveCancel on
	// return guarantees it exits even if Serve fails for a reason other
	// than our own Shutdown (otherwise it would block on <-ctx.Done()
	// for the parent context's whole lifetime — a goroutine leak).
	serveCtx, serveCancel := context.WithCancel(ctx)
	defer serveCancel()
	go func() {
		<-serveCtx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	logger.Info("listening", "addr", ln.Addr().String())
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if err := run(ctx, addr, os.Stdout); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
