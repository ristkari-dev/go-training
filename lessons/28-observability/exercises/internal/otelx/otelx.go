// Package otelx wires OpenTelemetry for logstatsd: tracer + meter
// providers exporting to stdout (no collector needed), plus an HTTP
// middleware that records a span, request metrics, and a
// trace-correlated logger.
package otelx

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Setup builds global tracer + meter providers that export to w (stdout
// in production) and returns a shutdown func that flushes both. IMPLEMENT THIS.
//
// Hint:
//
//	traceExp, _ := stdouttrace.New(stdouttrace.WithWriter(w))
//	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp)); otel.SetTracerProvider(tp)
//	metricExp, _ := stdoutmetric.New(stdoutmetric.WithWriter(w))
//	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp))); otel.SetMeterProvider(mp)
//	return func(ctx) error { return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx)) }, nil
//
// (Imports you'll add: errors, the stdouttrace/stdoutmetric exporters,
// sdktrace "go.opentelemetry.io/otel/sdk/trace", sdkmetric "go.opentelemetry.io/otel/sdk/metric".)
func Setup(ctx context.Context, w io.Writer) (func(context.Context) error, error) {
	_ = w
	panic("TODO: build TracerProvider (stdouttrace) + MeterProvider (stdoutmetric), set global, return flushing shutdown")
}

// statusRecorder captures the response status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Instrument wraps next with a per-request span, request counter +
// latency histogram, and a trace-correlated request logger. Provided.
func Instrument(next http.Handler, logger *slog.Logger) http.Handler {
	tracer := otel.Tracer("logstatsd/http")
	meter := otel.Meter("logstatsd/http")
	reqs, _ := meter.Int64Counter("http_requests_total")
	dur, _ := meter.Float64Histogram("http_request_duration_seconds")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		// Trace/log correlation: stamp the trace ID on a request logger.
		reqLogger := logger.With("trace_id", span.SpanContext().TraceID().String(),
			"method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(rec, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.status_code", rec.status))
		// status is low-cardinality — safe as a metric attribute (the raw
		// path is NOT, so we don't use it here).
		reqs.Add(ctx, 1, metric.WithAttributes(attribute.Int("status", rec.status)))
		dur.Record(ctx, time.Since(start).Seconds())
		reqLogger.Info("request handled", "status", rec.status,
			"duration_ms", time.Since(start).Milliseconds())
	})
}
