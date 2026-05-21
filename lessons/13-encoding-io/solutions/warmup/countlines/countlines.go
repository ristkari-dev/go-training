// Package countlines is the lesson 13 warm-up reference implementation.
package countlines

import (
	"bufio"
	"io"
)

// CountLines counts lines read from r via bufio.Scanner.
func CountLines(r io.Reader) (int, error) {
	s := bufio.NewScanner(r)
	var n int
	for s.Scan() {
		n++
	}
	if err := s.Err(); err != nil {
		return n, err
	}
	return n, nil
}
