package errgroupx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSuccess(t *testing.T) {
	g, _ := WithContext(context.Background())
	for range 3 {
		g.Go(func() error { return nil })
	}
	if err := g.Wait(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestFirstErrorWins(t *testing.T) {
	g, _ := WithContext(context.Background())
	sentinel := errors.New("first fail")

	g.Go(func() error { return nil })
	g.Go(func() error { return sentinel })
	g.Go(func() error { return nil })

	if err := g.Wait(); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCtxCancelledOnError(t *testing.T) {
	g, ctx := WithContext(context.Background())
	sentinel := errors.New("fail")

	g.Go(func() error { return sentinel })

	select {
	case <-ctx.Done():
		// Expected — first error cancels ctx.
	case <-time.After(time.Second):
		t.Fatal("ctx never cancelled after first error")
	}

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", ctx.Err())
	}

	_ = g.Wait()
}

func TestCtxCancelledOnWait(t *testing.T) {
	// Even with no errors, ctx should be cancelled when Wait returns
	// (releases ctx-watching goroutines that aren't part of the group).
	g, ctx := WithContext(context.Background())
	g.Go(func() error { return nil })
	_ = g.Wait()

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("expected ctx cancelled after Wait, got %v", ctx.Err())
	}
}
