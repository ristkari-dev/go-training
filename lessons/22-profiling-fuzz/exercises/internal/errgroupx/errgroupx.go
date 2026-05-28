// Package errgroupx is the lesson 22 "manage N workers" library —
// a lesson-local reimplementation of golang.org/x/sync/errgroup.
//
// The pattern: spawn N goroutines that may return errors. When the
// first one fails, cancel the rest via a shared context, then return
// that first error. If they all succeed, return nil.
//
// API mirrors x/sync/errgroup exactly:
//
//	g, ctx := errgroupx.WithContext(parent)
//	for _, item := range items {
//	    item := item
//	    g.Go(func() error {
//	        return process(ctx, item)
//	    })
//	}
//	if err := g.Wait(); err != nil { return err }
//
// In production code, use the real golang.org/x/sync/errgroup. This
// package teaches the pattern from the inside.
package errgroupx

import (
	"context"
	"sync"
)

// Group is a collection of goroutines working on subtasks that are
// part of the same overall task.
//
// A zero Group is valid, but a Group created by WithContext also
// holds an associated context that is cancelled on the first error
// from any subtask or on Wait returning.
type Group struct {
	cancel  func(error)
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

// WithContext returns a new Group and a derived Context. The Context
// is cancelled when the first non-nil error is returned by any Go-
// spawned function, or when Wait returns, whichever occurs first.
//
// Hint:
//  1. ctx, cancel := context.WithCancelCause(parent)  // Go 1.20+
//  2. return &Group{cancel: cancel}, ctx
//
// WithCancelCause is like WithCancel but lets you record WHY the
// context was cancelled (useful for debugging; matches real errgroup).
func WithContext(ctx context.Context) (*Group, context.Context) {
	_ = context.WithCancelCause
	panic("TODO: WithCancelCause + return Group with cancel set")
}

// Go calls f in a new goroutine. The first call to return a non-nil
// error cancels the group's context (if any).
//
// Hint:
//  1. g.wg.Add(1)
//  2. go func() {
//     defer g.wg.Done()
//     if err := f(); err != nil {
//     g.errOnce.Do(func() {
//     g.err = err
//     if g.cancel != nil {
//     g.cancel(err)
//     }
//     })
//     }
//     }()
//
// errOnce ensures we only record the FIRST error. Subsequent failures
// are silently dropped (the first one represents the cause).
func (g *Group) Go(f func() error) {
	_ = sync.Once{}
	panic("TODO: wg.Add + spawn goroutine + errOnce.Do capture")
}

// Wait blocks until all functions launched by Go have returned, then
// returns the first non-nil error (if any) from them.
//
// Hint:
//  1. g.wg.Wait()
//  2. if g.cancel != nil { g.cancel(g.err) }
//  3. return g.err
//
// The cancel call ensures the ctx is cancelled even if no error
// occurred — releases timers and any goroutines watching ctx.Done.
func (g *Group) Wait() error {
	panic("TODO: wg.Wait + cancel + return g.err")
}
