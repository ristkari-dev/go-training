package otelx

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSetupShutdown(t *testing.T) {
	shutdown, err := Setup(context.Background(), io.Discard)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}

func TestInstrument(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp)))
	reader := sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	h := Instrument(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), logger)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/stats", nil))

	// Span recorded with the right name.
	spans := exp.GetSpans()
	if len(spans) != 1 || spans[0].Name != "GET /stats" {
		t.Fatalf("spans = %+v", spans)
	}
	// Trace/log correlation: the span's trace ID appears in the log.
	if tid := spans[0].SpanContext.TraceID().String(); !strings.Contains(buf.String(), tid) {
		t.Errorf("log missing correlated trace_id %s: %s", tid, buf.String())
	}
	// Request counter incremented.
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http_requests_total" {
				if d, ok := m.Data.(metricdata.Sum[int64]); ok {
					for _, dp := range d.DataPoints {
						total += dp.Value
					}
				}
			}
		}
	}
	if total != 1 {
		t.Errorf("http_requests_total = %d, want 1", total)
	}
}
