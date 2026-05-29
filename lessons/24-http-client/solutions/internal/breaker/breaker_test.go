package breaker

import (
	"errors"
	"testing"
	"time"
)

// fakeClock is a manually-advanced time source for deterministic tests.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time      { return c.t }
func (c *fakeClock) add(d time.Duration) { c.t = c.t.Add(d) }

func TestBreakerOpensAfterThreshold(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(3, time.Minute, WithClock(clk.now))
	boom := errors.New("boom")

	for i := 0; i < 2; i++ {
		if err := b.Call(func() error { return boom }); err != boom {
			t.Fatalf("call %d = %v, want boom", i, err)
		}
		if b.State() != Closed {
			t.Fatalf("after %d failures state = %s, want closed", i+1, b.State())
		}
	}
	_ = b.Call(func() error { return boom }) // 3rd consecutive failure trips it
	if b.State() != Open {
		t.Fatalf("state = %s, want open", b.State())
	}
	called := false
	if err := b.Call(func() error { called = true; return nil }); !errors.Is(err, ErrOpen) {
		t.Errorf("open call = %v, want ErrOpen", err)
	}
	if called {
		t.Error("fn was invoked while breaker open")
	}
}

func TestBreakerHalfOpenToClosed(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(1, time.Minute, WithClock(clk.now))
	b.Call(func() error { return errors.New("boom") }) // opens (maxFailures=1)
	if b.State() != Open {
		t.Fatalf("state = %s, want open", b.State())
	}
	clk.add(time.Minute) // cooldown elapses
	if err := b.Call(func() error { return nil }); err != nil {
		t.Fatalf("trial call = %v, want nil", err)
	}
	if b.State() != Closed {
		t.Errorf("state = %s, want closed", b.State())
	}
}

func TestBreakerHalfOpenToOpen(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(1, time.Minute, WithClock(clk.now))
	boom := errors.New("boom")
	b.Call(func() error { return boom }) // opens
	clk.add(time.Minute)                 // cooldown elapses → next call is a trial
	if err := b.Call(func() error { return boom }); err != boom {
		t.Fatalf("trial = %v, want boom", err)
	}
	if b.State() != Open {
		t.Errorf("state = %s, want open (trial failure re-opens)", b.State())
	}
	if err := b.Call(func() error { return nil }); !errors.Is(err, ErrOpen) {
		t.Errorf("call right after re-open = %v, want ErrOpen", err)
	}
}

func TestBreakerStillOpenBeforeCooldown(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	b := New(1, time.Minute, WithClock(clk.now))
	b.Call(func() error { return errors.New("boom") }) // opens
	clk.add(30 * time.Second)                          // less than cooldown
	if err := b.Call(func() error { return nil }); !errors.Is(err, ErrOpen) {
		t.Errorf("call before cooldown = %v, want ErrOpen", err)
	}
}
