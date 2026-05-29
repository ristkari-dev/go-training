// Package breaker implements a circuit breaker: it fails fast when a
// dependency is unhealthy, giving it room to recover. Reference impl.
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

	mu               sync.Mutex
	state            State
	failures         int
	openedAt         time.Time
	halfOpenInFlight bool // a trial call is currently being attempted
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

// Call runs fn unless the breaker is open. It updates state based on
// the result. When open and the cooldown hasn't elapsed it returns
// ErrOpen without calling fn.
func (b *Breaker) Call(fn func() error) error {
	if !b.allow() {
		return ErrOpen
	}
	err := fn()
	b.record(err)
	return err
}

// allow reports whether a call may proceed. It transitions Open →
// HalfOpen when the cooldown has elapsed and admits exactly ONE trial:
// while that trial is in flight, further callers are denied (so a
// recovering dependency gets a single probe, not a fresh stampede).
func (b *Breaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case Open:
		if b.now().Sub(b.openedAt) >= b.cooldown {
			b.state = HalfOpen
			b.halfOpenInFlight = true
			return true // the single trial
		}
		return false
	case HalfOpen:
		return false // a trial is already in flight
	default: // Closed
		return true
	}
}

func (b *Breaker) record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.halfOpenInFlight = false // the trial (if any) has completed
	if err != nil {
		b.failures++
		// A failure in HalfOpen, or hitting the threshold in Closed,
		// (re)opens the breaker.
		if b.state == HalfOpen || b.failures >= b.maxFailures {
			b.state = Open
			b.openedAt = b.now()
		}
		return
	}
	// Success resets.
	b.failures = 0
	b.state = Closed
}

// State returns the current state (mainly for tests/introspection).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
