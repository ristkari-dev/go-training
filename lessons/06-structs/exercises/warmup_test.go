package exercises

import "testing"

// TestWarmupPointDistanceFromOrigin is a SKELETON. Fill in the cases and
// the t.Run body. Distances on floats need a tolerance — see the doc-comment
// hint about absolute-difference comparison.
//
// Cases to cover: a 3-4-5 right triangle (exact), the origin (0), at least
// one case with both negative coordinates, and a small case where the
// answer is irrational (use math.Abs(got - want) < 1e-9 for the comparison).
func TestWarmupPointDistanceFromOrigin(t *testing.T) {
	cases := []struct {
		name string
		p    WarmupPoint
		want float64
	}{
		// TODO: add at least 4 cases. Use the examples in the
		// DistanceFromOrigin doc comment as a starting point.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got := tc.p.DistanceFromOrigin()
			//   diff := got - tc.want
			//   if diff < 0 { diff = -diff }    // math.Abs without the import
			//   if diff > 1e-9 { t.Errorf("...") }
			//
			// Or import math and use math.Abs.
			_ = tc
		})
	}
}
