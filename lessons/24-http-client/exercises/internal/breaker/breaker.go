// Package breaker implements a circuit breaker: it fails fast when a
// dependency is unhealthy, giving it room to recover.
package breaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned by Call when the breaker is open (failing fast).
var ErrOpen = errors.New("breaker: circuit open")

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Closed:
		return "closed"
	case Open:
		return "open"
	case HalfOpen:
		return "half-open"
	}
	return "unknown"
}

// Breaker is a circuit breaker: closed → open (after maxFailures
// consecutive failures) → half-open (after cooldown) → closed (on a
// trial success) or open (on a trial failure).
type Breaker struct {
	maxFailures int
	cooldown    time.Duration
	now         func() time.Time

	mu       sync.Mutex
	state    State
	failures int
	openedAt time.Time
}

// Option configures a Breaker.
type Option func(*Breaker)

// WithClock overrides the time source (for tests).
func WithClock(now func() time.Time) Option {
	return func(b *Breaker) { b.now = now }
}

func New(maxFailures int, cooldown time.Duration, opts ...Option) *Breaker {
	b := &Breaker{
		maxFailures: maxFailures,
		cooldown:    cooldown,
		now:         time.Now,
		state:       Closed,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Call runs fn unless the breaker is open (then returns ErrOpen without
// calling fn). It updates state from the result: a success closes the
// breaker and resets the failure count; a failure increments it and
// opens the breaker at the threshold (or immediately, if half-open).
//
// Hint: gate on allow(); call fn; record(err); return.
func (b *Breaker) Call(fn func() error) error {
	_ = ErrOpen
	panic("TODO: if !allow() return ErrOpen; err := fn(); record(err); return err")
}

// allow reports whether a call may proceed, transitioning Open →
// HalfOpen when the cooldown has elapsed (allowing one trial).
func (b *Breaker) allow() bool {
	panic("TODO: lock; if Open and now-openedAt >= cooldown → HalfOpen, allow one trial; Open → deny; else allow")
}

// record updates state from a call's outcome.
func (b *Breaker) record(err error) {
	panic("TODO: lock; on error → failures++, open if HalfOpen or failures>=max (stamp openedAt); on success → reset to Closed")
}

// State returns the current state (mainly for tests/introspection).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
