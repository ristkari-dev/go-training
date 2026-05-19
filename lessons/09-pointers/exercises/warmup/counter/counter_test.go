package counter

import "testing"

// TestCounter is a SKELETON. Fill in the cases and the t.Run body.
//
// Each case describes a sequence of Inc() calls then asserts the final
// Value(). Cases to cover: zero (no Inc calls), one Inc, several Inc calls,
// and a sanity check that two separate Counters don't interfere.
//
// Hint: the cleanest shape is a table of (name, incCount, want):
//
//	cases := []struct {
//		name     string
//		incCount int
//		want     int
//	}{ ... }
//
// Then in the loop: declare a fresh `var c Counter`, call c.Inc() in a for
// loop incCount times, assert c.Value() == want.
func TestCounter(t *testing.T) {
	cases := []struct {
		name     string
		incCount int
		want     int
	}{
		// TODO: at least 4 cases. zero, one, many, large-n.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: var c Counter
			//   for i := 0; i < tc.incCount; i++ { c.Inc() }
			//   if got := c.Value(); got != tc.want { t.Errorf(...) }
			_ = tc
		})
	}
}

// TestCountersDoNotInterfere is a SKELETON. Verify that two separate
// Counter values are independent — incrementing one does not affect the
// other. This catches a class of bug where Counter was implemented as a
// pointer-typed field (or used a package-level var).
func TestCountersDoNotInterfere(t *testing.T) {
	// TODO:
	//   var a, b Counter
	//   a.Inc(); a.Inc(); a.Inc()
	//   b.Inc()
	//   if got := a.Value(); got != 3 { t.Errorf(...) }
	//   if got := b.Value(); got != 1 { t.Errorf(...) }
	_ = t
}
