// Package main is the logstatsd daemon: it serves the logstats HTTP and
// gRPC faces from one process over a shared store, configured by
// flags>env>file, and drains both gracefully on SIGINT/SIGTERM.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/internal/httpsrv"
	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/26-config/solutions/proto/logstatspb"
	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/warmup/config"
)

const drainTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load(os.Args[1:], os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newLogger(level string, w io.Writer) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: l}))
}

// run wires config → servers → listeners, logs the (redacted) config,
// and hands off to serve. Provided.
func run(ctx context.Context, cfg config.Config, stdout io.Writer) error {
	logger := newLogger(cfg.LogLevel, stdout)
	logger.Info("starting logstatsd", "config", cfg) // redacted via Config.LogValue

	store := logstats.NewStore()

	httpLis, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("http listen: %w", err)
	}
	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	httpSrv := &http.Server{Handler: httpsrv.Router(store, logger)}
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcsrv.LoggingUnaryInterceptor(logger)),
		grpc.StreamInterceptor(grpcsrv.LoggingStreamInterceptor(logger)),
	)
	pb.RegisterLogStatsServer(grpcSrv, grpcsrv.New(store))

	logger.Info("listening", "http", httpLis.Addr().String(), "grpc", grpcLis.Addr().String())
	return serve(ctx, httpSrv, httpLis, grpcSrv, grpcLis)
}

// serve runs both servers concurrently and drains BOTH when ctx is
// cancelled. If either fails first, stop the other and return the error.
func serve(ctx context.Context, httpSrv *http.Server, httpLis net.Listener, grpcSrv *grpc.Server, grpcLis net.Listener) error {
	errCh := make(chan error, 2)
	go func() {
		if err := httpSrv.Serve(httpLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	go func() {
		if err := grpcSrv.Serve(grpcLis); err != nil {
			errCh <- err
		}
	}()

	drain := func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
		defer cancel()
		_ = httpSrv.Shutdown(shutCtx)
		grpcSrv.GracefulStop()
	}

	select {
	case <-ctx.Done():
		drain()
		return nil
	case err := <-errCh:
		drain()
		return err
	}
}
