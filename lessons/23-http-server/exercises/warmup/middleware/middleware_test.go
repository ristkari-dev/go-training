package middleware

import (
	"log/slog"
	"net/http"
	"testing"
)

// TestWithRequestLog is a SKELETON. Wrap a handler that writes a known
// status; assert the wrapped handler passes the status through AND that
// the logger recorded method/path/status.
func TestWithRequestLog(t *testing.T) {
	// TODO:
	//   var buf bytes.Buffer
	//   logger := slog.New(slog.NewJSONHandler(&buf, nil))
	//   h := WithRequestLog(http.HandlerFunc(func(w, r){ w.WriteHeader(418); w.Write([]byte("hi")) }), logger)
	//   rec := httptest.NewRecorder(); h.ServeHTTP(rec, httptest.NewRequest("GET", "/foo", nil))
	//   assert rec.Code == 418
	//   assert buf contains "method":"GET", "path":"/foo", "status":418
	_ = slog.New
	_ = http.HandlerFunc(nil)
}
