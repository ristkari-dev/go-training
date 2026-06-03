// Package main is the logstatsd daemon: it serves the logstats HTTP and
// gRPC faces from one process over a shared store, configured by
// flags>env>file, and drains both gracefully on SIGINT/SIGTERM.
package main

import (
	"context"
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

	"github.com/ristkari-dev/go-training/lessons/28-observability/exercises/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/28-observability/exercises/internal/httpsrv"
	"github.com/ristkari-dev/go-training/lessons/28-observability/exercises/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/28-observability/exercises/proto/logstatspb"
	"github.com/ristkari-dev/go-training/lessons/28-observability/exercises/warmup/buildinfo"
	"github.com/ristkari-dev/go-training/lessons/28-observability/exercises/warmup/config"
)

const drainTimeout = 10 * time.Second

func main() {
	// Pre-scan for -version/--version before config parsing so it works
	// regardless of the config FlagSet.
	for _, a := range os.Args[1:] {
		if a == "-version" || a == "--version" {
			fmt.Println(buildinfo.String())
			return
		}
	}

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

// serve runs both servers concurrently and stops BOTH when ctx is
// cancelled (graceful) or when one of them fails (immediate). IMPLEMENT THIS.
//
// Hint:
//
//	errCh := make(chan error, 2)
//	go func(){ if err := httpSrv.Serve(httpLis); err != nil && !errors.Is(err, http.ErrServerClosed) { errCh <- err } }()
//	go func(){ if err := grpcSrv.Serve(grpcLis); err != nil { errCh <- err } }()
//	select {
//	case <-ctx.Done():           // requested shutdown → drain gracefully, bounded
//	    shutCtx, cancel := context.WithTimeout(context.Background(), drainTimeout); defer cancel()
//	    httpSrv.Shutdown(shutCtx); grpcSrv.GracefulStop(); return nil
//	case err := <-errCh:         // a server crashed → stop the other IMMEDIATELY
//	    httpSrv.Close(); grpcSrv.Stop(); return err
//	}
//
// Why immediate on the error path: the service is already broken, so
// don't wait for in-flight work — an unbounded GracefulStop could hang
// on a stuck stream.
func serve(ctx context.Context, httpSrv *http.Server, httpLis net.Listener, grpcSrv *grpc.Server, grpcLis net.Listener) error {
	_ = drainTimeout
	panic("TODO: run both servers; on ctx.Done drain gracefully (http.Shutdown + grpc.GracefulStop); on server error stop both immediately (http.Close + grpc.Stop)")
}
