// Package main is the logstats-forward client: it reads log lines from
// stdin and ships them to an aggregator's /ingest endpoint via the
// forwarder (outbox + at-least-once retry + stable idempotency keys).
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/exercises/internal/forwarder"
)

func main() {
	addr := flag.String("addr", "http://127.0.0.1:8080/ingest", "aggregator /ingest endpoint URL")
	id := flag.String("id", "fwd1", "forwarder id (tags idempotency keys)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, *addr, *id, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// run reads non-empty log lines from stdin and forwards them as one batch
// to addr, then prints a one-line summary to stdout.
func run(ctx context.Context, addr, id string, stdin io.Reader, stdout io.Writer) error {
	var lines []string
	sc := bufio.NewScanner(stdin)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if err := forwarder.New(addr, id).Send(ctx, lines); err != nil {
		return fmt.Errorf("forward: %w", err)
	}
	fmt.Fprintf(stdout, "forwarded %d lines to %s\n", len(lines), addr)
	return nil
}
