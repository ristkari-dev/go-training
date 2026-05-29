// Package shipper is a resilient HTTP client that POSTs log lines to the
// logstats server's /ingest, with timeouts, selective retry, and a
// circuit breaker. Reference implementation.
package shipper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ristkari-dev/go-training/lessons/24-http-client/solutions/internal/breaker"
	"github.com/ristkari-dev/go-training/lessons/24-http-client/solutions/warmup/retry"
)

type ingestRequest struct {
	Lines []string `json:"lines"`
}

// Result mirrors the server's ingest response.
type Result struct {
	Accepted int `json:"accepted"`
	Parsed   int `json:"parsed"`
	Failed   int `json:"failed"`
}

// permanentError marks a non-retryable failure (4xx other than 429).
type permanentError struct{ status int }

func (e *permanentError) Error() string {
	return fmt.Sprintf("shipper: permanent failure: HTTP %d", e.status)
}

// transientError marks a retryable failure (5xx, 429, network).
type transientError struct {
	status int // 0 for a network error
	err    error
}

func (e *transientError) Error() string {
	if e.status != 0 {
		return fmt.Sprintf("shipper: transient failure: HTTP %d", e.status)
	}
	return fmt.Sprintf("shipper: transient failure: %v", e.err)
}

type Shipper struct {
	url         string
	client      *http.Client
	breaker     *breaker.Breaker
	maxAttempts int
	baseBackoff time.Duration
}

type Option func(*Shipper)

func WithBreaker(b *breaker.Breaker) Option  { return func(s *Shipper) { s.breaker = b } }
func WithMaxAttempts(n int) Option           { return func(s *Shipper) { s.maxAttempts = n } }
func WithBaseBackoff(d time.Duration) Option { return func(s *Shipper) { s.baseBackoff = d } }
func WithClient(c *http.Client) Option       { return func(s *Shipper) { s.client = c } }

func New(url string, opts ...Option) *Shipper {
	s := &Shipper{
		url:         url,
		client:      &http.Client{Timeout: 10 * time.Second},
		breaker:     breaker.New(5, 30*time.Second),
		maxAttempts: 3,
		baseBackoff: 100 * time.Millisecond,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Ship POSTs lines to the server's /ingest, retrying transient
// failures with backoff, short-circuiting via the breaker. Returns the
// parsed Result on success.
//
// Layering: the breaker counts only TRANSIENT failures (5xx/429/network)
// — a 4xx is our bad request, not the dependency being down, so it must
// not trip the breaker. Permanent (4xx) errors and an open breaker both
// stop the retry loop immediately via retry.Permanent.
func (s *Shipper) Ship(ctx context.Context, lines []string) (Result, error) {
	body, err := json.Marshal(ingestRequest{Lines: lines})
	if err != nil {
		return Result{}, err
	}

	var result Result
	runErr := retry.Do(ctx, s.maxAttempts, s.baseBackoff, func() error {
		var permanent error
		callErr := s.breaker.Call(func() error {
			res, e := s.doOnce(ctx, body)
			if e != nil {
				var p *permanentError
				if errors.As(e, &p) {
					permanent = e // capture; don't let it trip the breaker
					return nil    // breaker sees a healthy dependency
				}
				return e // transient → breaker counts it, retry will retry
			}
			result = res
			return nil
		})
		if permanent != nil {
			return retry.Permanent(permanent) // 4xx → stop now
		}
		if errors.Is(callErr, breaker.ErrOpen) {
			return retry.Permanent(callErr) // open → fail fast, no backoff
		}
		return callErr // nil (success) or transient (retry)
	})
	return result, runErr
}

// doOnce performs a single POST and classifies the outcome.
func (s *Shipper) doOnce(ctx context.Context, body []byte) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return Result{}, &transientError{err: err} // network error → transient
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		var r Result
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return Result{}, &transientError{err: err}
		}
		return r, nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return Result{}, &transientError{status: resp.StatusCode} // 429 → retry with backoff
	case resp.StatusCode >= 500:
		return Result{}, &transientError{status: resp.StatusCode}
	default: // 4xx other than 429
		return Result{}, &permanentError{status: resp.StatusCode}
	}
}
