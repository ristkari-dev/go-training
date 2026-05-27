// Package errgroupx is the lesson 20 reference implementation —
// a tiny errgroup mirroring golang.org/x/sync/errgroup.
package errgroupx

import (
	"context"
	"sync"
)

// Group manages a collection of goroutines working on subtasks.
type Group struct {
	cancel  func(error)
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

// WithContext returns a Group and a derived Context. The Context is
// cancelled on the first error or when Wait returns.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}

// Go spawns f as a goroutine. First error cancels the group's ctx.
func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(err)
				}
			})
		}
	}()
}

// Wait blocks until all Go-spawned functions return, then returns the
// first non-nil error (if any).
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}
