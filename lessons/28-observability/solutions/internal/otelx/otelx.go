// Package otelx wires OpenTelemetry for logstatsd. Reference impl.
package otelx

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Setup builds global tracer + meter providers exporting to w and
// returns a shutdown func that flushes both.
func Setup(ctx context.Context, w io.Writer) (func(context.Context) error, error) {
	traceExp, err := stdouttrace.New(stdouttrace.WithWriter(w))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExp))
	otel.SetTracerProvider(tp)

	metricExp, err := stdoutmetric.New(stdoutmetric.WithWriter(w))
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)))
	otel.SetMeterProvider(mp)

	return func(ctx context.Context) error {
		return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx))
	}, nil
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Instrument wraps next with a per-request span, request counter +
// latency histogram, and a trace-correlated request logger.
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
		reqLogger := logger.With("trace_id", span.SpanContext().TraceID().String(),
			"method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(rec, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.status_code", rec.status))
		reqs.Add(ctx, 1, metric.WithAttributes(attribute.Int("status", rec.status)))
		dur.Record(ctx, time.Since(start).Seconds())
		reqLogger.Info("request handled", "status", rec.status,
			"duration_ms", time.Since(start).Milliseconds())
	})
}
