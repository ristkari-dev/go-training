// Package forwarder ships log-line batches to an aggregator's /ingest
// with an outbox + at-least-once retry + stable idempotency keys.
package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ingestRequest struct {
	Lines []string `json:"lines"`
}

// batch is an outbox entry: a unit of work with a stable idempotency key
// reused across retries (so a redelivery dedups, not double-counts).
type batch struct {
	key   string
	lines []string
}

type Forwarder struct {
	url         string
	id          string
	client      *http.Client
	maxAttempts int
	baseBackoff time.Duration

	seq    int
	outbox []batch
}

// New returns a Forwarder posting to url (the full /ingest endpoint),
// tagging keys with id (so multiple forwarders don't collide).
func New(url, id string) *Forwarder {
	return &Forwarder{
		url: url, id: id,
		client:      &http.Client{Timeout: 5 * time.Second},
		maxAttempts: 5,
		baseBackoff: time.Microsecond,
	}
}

// Send enqueues lines into the outbox with a fresh stable key, then
// flushes the outbox (delivering pending batches in order).
func (f *Forwarder) Send(ctx context.Context, lines []string) error {
	f.seq++
	f.outbox = append(f.outbox, batch{key: fmt.Sprintf("%s-%d", f.id, f.seq), lines: lines})
	return f.flush(ctx)
}

// flush delivers outbox batches in order, stopping (and keeping the rest)
// on the first delivery error.
func (f *Forwarder) flush(ctx context.Context) error {
	for len(f.outbox) > 0 {
		if err := f.deliver(ctx, f.outbox[0]); err != nil {
			return err
		}
		f.outbox = f.outbox[1:]
	}
	return nil
}

// deliver POSTs one batch with at-least-once retry, REUSING the batch's
// idempotency key on every attempt (so a redelivery after a lost ack is
// deduped by the server). Retries transient failures (network, 5xx);
// gives up on a permanent (4xx) failure. IMPLEMENT THIS.
//
// Hint:
//
//	body, _ := json.Marshal(ingestRequest{Lines: b.lines})
//	for attempt := 0; attempt < f.maxAttempts; attempt++ {
//	    if attempt > 0 { backoff via time.NewTimer(f.baseBackoff << (attempt-1)) honoring ctx }
//	    req, _ := http.NewRequestWithContext(ctx, POST, f.url, bytes.NewReader(body))
//	    req.Header.Set("Content-Type","application/json")
//	    req.Header.Set("Idempotency-Key", b.key)   // SAME key every attempt
//	    resp, err := f.client.Do(req)
//	    if err != nil { lastErr = err; continue }   // network → retry
//	    drain+close body
//	    2xx → return nil; 5xx → retry; 4xx → return permanent error
//	}
//	return lastErr
func (f *Forwarder) deliver(ctx context.Context, b batch) error {
	_ = bytes.NewReader
	_ = io.Discard
	_ = json.Marshal
	panic("TODO: POST with retry, reusing b.key each attempt; 2xx ok, 5xx/network retry, 4xx give up")
}
