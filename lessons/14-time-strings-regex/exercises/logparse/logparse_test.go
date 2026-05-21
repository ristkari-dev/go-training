package logparse

import (
	"strings"
	"testing"
)

// TestParse is a SKELETON. Cover at least:
//   - Single well-formed line → one entry, no error
//   - Multiple well-formed lines → three entries
//   - Blank lines interspersed → skipped, count matches non-blank lines
//   - One malformed line in the middle → partial result + wrapped error mentioning line number
func TestParse(t *testing.T) {
	// TODO: build inputs with strings.NewReader; call Parse; assert structure.
	_ = strings.NewReader
}

// TestCountByLevel is a SKELETON. Cover at least:
//   - Mixed levels → correct counts per level
//   - All same level → single map entry with full count
//   - Empty input → empty map (or nil — either acceptable)
func TestCountByLevel(t *testing.T) {
	// TODO: build []LogEntry directly; call CountByLevel; assert map content.
}
