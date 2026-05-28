// Package middleware is the lesson 23 warm-up: an HTTP logging
// middleware. A middleware is a function that wraps an http.Handler
// and returns a new http.Handler — the foundation of request logging,
// auth, recovery, and more.
package middleware

import (
	"log/slog"
	"net/http"
)

// WithRequestLog wraps next so that every request is logged with its
// method, path, status code, and duration. Returns a new handler.
//
// The trick: http.ResponseWriter doesn't expose the status code after
// the fact, so wrap it in a small type that records the code passed to
// WriteHeader (defaulting to 200, which net/http uses when a handler
// writes a body without calling WriteHeader).
//
// Hint:
//  1. return http.HandlerFunc(func(w, r) {
//  2. start := time.Now()
//  3. rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
//  4. next.ServeHTTP(rec, r)
//  5. logger.Info("request", "method", r.Method, "path", r.URL.Path,
//     "status", rec.status, "duration_ms", time.Since(start).Milliseconds())
//  6. })
//
// You'll need a statusRecorder type embedding http.ResponseWriter with
// an overridden WriteHeader that stores the code.
func WithRequestLog(next http.Handler, logger *slog.Logger) http.Handler {
	_ = slog.LevelInfo
	panic("TODO: wrap next; record status via a ResponseWriter wrapper; log method/path/status/duration")
}
