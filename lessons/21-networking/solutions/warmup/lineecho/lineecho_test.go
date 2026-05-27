package lineecho

import (
	"bufio"
	"fmt"
	"net"
	"testing"
)

func TestLineEcho(t *testing.T) {
	cases := []struct {
		name string
		send []string
		want []string
	}{
		{"single", []string{"hello"}, []string{"HELLO"}},
		{"multiple", []string{"hi", "world"}, []string{"HI", "WORLD"}},
		{"mixed-case", []string{"Hello World"}, []string{"HELLO WORLD"}},
		{"unicode", []string{"héllo"}, []string{"HÉLLO"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			serverEnd, clientEnd := net.Pipe()

			done := make(chan error, 1)
			go func() {
				done <- LineEcho(serverEnd)
				_ = serverEnd.Close()
			}()

			// Client writer: send all lines, then we'll close after reads.
			go func() {
				w := bufio.NewWriter(clientEnd)
				for _, line := range tc.send {
					fmt.Fprintln(w, line)
				}
				_ = w.Flush()
			}()

			// Read responses from client side.
			r := bufio.NewScanner(clientEnd)
			var got []string
			for i := 0; i < len(tc.want); i++ {
				if !r.Scan() {
					t.Fatalf("expected %d responses, got %d (err: %v)", len(tc.want), i, r.Err())
				}
				got = append(got, r.Text())
			}

			// Close client side to signal EOF to LineEcho.
			_ = clientEnd.Close()

			// Wait for LineEcho to finish.
			if err := <-done; err != nil {
				t.Errorf("LineEcho returned error: %v", err)
			}

			if !equalStrings(got, tc.want) {
				t.Errorf("LineEcho got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLineEchoEmpty(t *testing.T) {
	serverEnd, clientEnd := net.Pipe()

	done := make(chan error, 1)
	go func() {
		done <- LineEcho(serverEnd)
		_ = serverEnd.Close()
	}()

	// Close client side immediately — LineEcho sees EOF.
	_ = clientEnd.Close()

	if err := <-done; err != nil {
		t.Errorf("LineEcho with empty input returned error: %v", err)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
