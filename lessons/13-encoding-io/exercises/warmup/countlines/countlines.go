// Package countlines is the lesson 13 warm-up: counting lines via
// bufio.Scanner from any io.Reader.
//
// CountLines demonstrates two of the lesson's core ideas at once:
//
//  1. Accept io.Reader, not a path. The function doesn't open or close
//     anything — it just reads from whatever you hand it. Callers can
//     hand it *os.File (for a real file), strings.NewReader (for tests
//     with literal input), or any other type that satisfies io.Reader.
//
//  2. The canonical bufio.Scanner loop: for s.Scan() { ... } followed by
//     a final s.Err() check. The Scan() method returns false on either
//     EOF (normal) OR error (problem). The Err() call distinguishes —
//     it returns nil on clean EOF, non-nil if Scan stopped because of
//     an error.
package countlines

import (
	"bufio"
	"io"
)

// CountLines returns the number of lines read from r.
//
// "Line" means "anything separated by '\n'" — bufio.Scanner's default
// split function (bufio.ScanLines) handles \r\n correctly too.
//
// Edge cases:
//   - Empty reader → (0, nil)
//   - "foo" (no trailing newline) → (1, nil)        // one line, no terminator
//   - "foo\n" (trailing newline) → (1, nil)         // still one line
//   - "foo\nbar\n" → (2, nil)
//   - Reader that errors mid-stream → (partial count, non-nil error)
//
// Hint:
//  1. s := bufio.NewScanner(r)
//  2. var n int
//  3. for s.Scan() { n++ }
//  4. if err := s.Err(); err != nil { return n, err }
//  5. return n, nil
//
// The Err() check at step 4 is the easy thing to forget. Without it,
// a reader that fails mid-stream looks identical to clean EOF — both
// make Scan() return false. Always end the scan loop with Err().
func CountLines(r io.Reader) (int, error) {
	_ = bufio.NewScanner
	_ = io.EOF
	panic("TODO: bufio.NewScanner, scan loop, Err() check")
}
