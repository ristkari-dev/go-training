package shipper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestShip is a SKELETON. Use httptest servers to cover: success;
// transient (5xx) retried then succeeds; permanent (400) not retried;
// 429 retried; breaker opens after repeated 5xx and then fails fast.
func TestShip(t *testing.T) {
	// TODO: httptest.NewServer(...) returning 200/500/400/429; assert
	// Ship behavior + server hit counts (use sync/atomic counters).
	_ = http.StatusOK
	_ = httptest.NewServer
	_ = context.Background
}
