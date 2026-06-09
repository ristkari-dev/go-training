package tracemw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestWithTracing(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp)))
	tracer := otel.Tracer("test")

	h := WithTracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), tracer)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Name != "GET /x" {
		t.Errorf("span name = %q, want %q", spans[0].Name, "GET /x")
	}
	var found bool
	for _, a := range spans[0].Attributes {
		if string(a.Key) == "http.status_code" && a.Value.AsInt64() == http.StatusTeapot {
			found = true
		}
	}
	if !found {
		t.Errorf("span missing http.status_code=418 attribute: %+v", spans[0].Attributes)
	}
}
