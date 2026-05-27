// Package lineecho is the lesson 21 warm-up reference implementation.
package lineecho

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// LineEcho reads lines, writes each back uppercased.
func LineEcho(conn net.Conn) error {
	r := bufio.NewScanner(conn)
	w := bufio.NewWriter(conn)

	for r.Scan() {
		upper := strings.ToUpper(r.Text())
		if _, err := fmt.Fprintln(w, upper); err != nil {
			return err
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return r.Err()
}
