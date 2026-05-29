// Package main is the lesson 25 LogStats gRPC server. It registers the
// grpcsrv service (with logging interceptors) over a TCP listener,
// sharing a logstats.Store. Graceful stop on SIGINT.
//
// Usage: logstats-grpc-server 127.0.0.1:9090
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	"github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/proto/logstatspb"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := "127.0.0.1:9090"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	if err := run(ctx, addr); err != nil {
		slog.Error("grpc server failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, addr string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcsrv.LoggingUnaryInterceptor(logger)),
		grpc.StreamInterceptor(grpcsrv.LoggingStreamInterceptor(logger)),
	)
	pb.RegisterLogStatsServer(srv, grpcsrv.New(logstats.NewStore()))

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	logger.Info("grpc listening", "addr", lis.Addr().String())
	return srv.Serve(lis)
}
