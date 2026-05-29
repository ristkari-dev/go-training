package shipper

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ristkari-dev/go-training/lessons/24-http-client/solutions/internal/breaker"
)

func testShipper(url string, opts ...Option) *Shipper {
	base := []Option{
		WithMaxAttempts(4),
		WithBaseBackoff(time.Microsecond), // keep tests fast
	}
	return New(url, append(base, opts...)...)
}

func TestShipSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accepted":2,"parsed":2,"failed":0}`))
	}))
	defer srv.Close()

	res, err := testShipper(srv.URL).Ship(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("Ship = %v", err)
	}
	if res.Accepted != 2 || res.Parsed != 2 {
		t.Errorf("res = %+v", res)
	}
}

func TestShipRetriesTransientThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 → transient
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accepted":1,"parsed":1,"failed":0}`))
	}))
	defer srv.Close()

	res, err := testShipper(srv.URL).Ship(context.Background(), []string{"x"})
	if err != nil {
		t.Fatalf("Ship = %v", err)
	}
	if res.Parsed != 1 {
		t.Errorf("res = %+v", res)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("server calls = %d, want 3", calls)
	}
}

func TestShipPermanentNoRetry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest) // 400 → permanent
	}))
	defer srv.Close()

	_, err := testShipper(srv.URL).Ship(context.Background(), []string{"x"})
	if err == nil {
		t.Fatal("Ship = nil, want permanent error")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("server calls = %d, want 1 (no retry on 4xx)", calls)
	}
}

func TestShip429IsTransient(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests) // 429 → transient
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accepted":1,"parsed":1,"failed":0}`))
	}))
	defer srv.Close()

	_, err := testShipper(srv.URL).Ship(context.Background(), []string{"x"})
	if err != nil {
		t.Fatalf("Ship = %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("server calls = %d, want 2 (429 retried)", calls)
	}
}

func TestShipBreakerOpensAndFailsFast(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError) // always 500
	}))
	defer srv.Close()

	b := breaker.New(3, time.Minute) // opens after 3 consecutive failures
	s := testShipper(srv.URL, WithBreaker(b), WithMaxAttempts(10))

	_, err := s.Ship(context.Background(), []string{"x"})
	if err == nil {
		t.Fatal("Ship = nil, want error")
	}
	if got := atomic.LoadInt32(&calls); got > 3 {
		t.Errorf("server calls = %d, want <= 3 (breaker should stop the storm)", got)
	}
	if b.State() != breaker.Open {
		t.Errorf("breaker state = %s, want open", b.State())
	}
	before := atomic.LoadInt32(&calls)
	_, err = s.Ship(context.Background(), []string{"y"})
	if !errors.Is(err, breaker.ErrOpen) {
		t.Errorf("second Ship = %v, want ErrOpen", err)
	}
	if atomic.LoadInt32(&calls) != before {
		t.Errorf("server was hit while breaker open")
	}
}

func TestShipContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	_, err := testShipper(srv.URL, WithBaseBackoff(time.Second)).Ship(ctx, []string{"x"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Ship = %v, want context.Canceled", err)
	}
}
