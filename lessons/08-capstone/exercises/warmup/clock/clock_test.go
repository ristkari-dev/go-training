package clock

import "testing"

// TestNow is a SKELETON. Fill in the assertion body.
//
// Time-dependent tests are tricky — Now() returns the current time,
// which differs every call. The simplest correctness check is the FORMAT:
//
//   - The returned string is exactly 19 characters long ("YYYY-MM-DD HH:MM:SS").
//   - Characters at positions 4 and 7 are '-'; position 10 is ' ';
//     positions 13 and 16 are ':'.
//
// We don't check the actual time — only that the format is right. A more
// rigorous test would inject a "now" function as a dependency (lesson 10's
// interfaces unlock that pattern); for Phase 1, format-only is enough.
func TestNow(t *testing.T) {
	// TODO: got := Now()
	// TODO: assert len(got) == 19; t.Fatalf if not.
	// TODO: assert got[4] == '-' && got[7] == '-' && got[10] == ' ' &&
	//        got[13] == ':' && got[16] == ':'; t.Errorf for each.
	_ = t
}
