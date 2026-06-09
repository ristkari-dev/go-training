// Package tracemw is the lesson 29 warm-up: an HTTP middleware that
// records a tracing span for each request. It's the tracing half of the
// production otelx.Instrument middleware.
package tracemw

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// statusRecorder captures the response status code (the L23 trick).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// WithTracing wraps next so each request runs inside a span named
// "<METHOD> <path>", with the response status recorded as a span
// attribute. The span's context is threaded into the request so handlers
// can create child spans. IMPLEMENT THIS.
//
// Hint:
//
//	return http.HandlerFunc(func(w, r) {
//	    ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
//	    defer span.End()
//	    rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
//	    next.ServeHTTP(rec, r.WithContext(ctx))
//	    span.SetAttributes(attribute.Int("http.status_code", rec.status))
//	})
func WithTracing(next http.Handler, tracer trace.Tracer) http.Handler {
	_ = attribute.Int
	_ = statusRecorder{}
	panic("TODO: start a span named METHOD+path, thread ctx, record status attr on End")
}
