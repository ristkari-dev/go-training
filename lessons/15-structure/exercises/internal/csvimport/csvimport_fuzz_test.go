package csvimport

import (
	"testing"
)

// FuzzParseLine is a SKELETON. The full fuzz function exists in the
// solutions tree. Here the seeds + body are placeholders.
//
// Lesson 15 introduces the syntax; Phase 3 lesson 22 covers fuzzing in
// depth. To actually run the fuzzer:
//
//	go test -fuzz=FuzzParseLine ./lessons/15-structure/solutions/internal/csvimport/
//
// (Run from the solutions tree where the implementation is real.)
func FuzzParseLine(f *testing.F) {
	// TODO: f.Add three valid seeds; f.Fuzz a function that calls
	// parseLine and asserts no panic.
	f.Skip("fuzz body not implemented in exercises tree; see solutions")
}
