package errgroupx

import (
	"context"
	"errors"
	"testing"
)

// TestSuccess is a SKELETON. Spawn three goroutines that all return
// nil; assert Wait returns nil.
func TestSuccess(t *testing.T) {
	// TODO:
	//   g, _ := WithContext(context.Background())
	//   for i := 0; i < 3; i++ { g.Go(func() error { return nil }) }
	//   if err := g.Wait(); err != nil → t.Errorf("...")
	_ = context.Background
}

// TestFirstErrorWins is a SKELETON. Spawn three goroutines; one returns
// a sentinel error; the others succeed. Assert Wait returns that error.
func TestFirstErrorWins(t *testing.T) {
	// TODO:
	//   g, _ := WithContext(context.Background())
	//   sentinel := errors.New("first fail")
	//   g.Go(func() error { return nil })
	//   g.Go(func() error { return sentinel })
	//   g.Go(func() error { return nil })
	//   if err := g.Wait(); !errors.Is(err, sentinel) → t.Errorf("...")
	_ = errors.Is
}

// TestCtxCancelledOnError is a SKELETON. Verify that when one goroutine
// returns an error, the group's ctx is cancelled (other goroutines can
// see ctx.Done() fire).
func TestCtxCancelledOnError(t *testing.T) {
	// TODO:
	//   g, ctx := WithContext(context.Background())
	//   sentinel := errors.New("fail")
	//   g.Go(func() error { return sentinel })
	//   <-ctx.Done()  // should fire quickly
	//   if !errors.Is(ctx.Err(), context.Canceled) → t.Errorf("...")
}
