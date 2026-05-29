package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRequestLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := WithRequestLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hi"))
	}), logger)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/foo", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status passthrough = %d, want 418", rec.Code)
	}
	out := buf.String()
	for _, want := range []string{`"method":"GET"`, `"path":"/foo"`, `"status":418`} {
		if !strings.Contains(out, want) {
			t.Errorf("log missing %s in: %s", want, out)
		}
	}
}

func TestWithRequestLogDefaultStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	h := WithRequestLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}), logger)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(buf.String(), `"status":200`) {
		t.Errorf("default status not logged as 200: %s", buf.String())
	}
}
