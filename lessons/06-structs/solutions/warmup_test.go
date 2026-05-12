package solutions

import (
	"math"
	"testing"
)

func TestWarmupPointDistanceFromOrigin(t *testing.T) {
	cases := []struct {
		name string
		p    WarmupPoint
		want float64
	}{
		{"3-4-5", WarmupPoint{X: 3, Y: 4}, 5},
		{"origin", WarmupPoint{X: 0, Y: 0}, 0},
		{"unit-x", WarmupPoint{X: 1, Y: 0}, 1},
		{"unit-y", WarmupPoint{X: 0, Y: 1}, 1},
		{"diagonal", WarmupPoint{X: 1, Y: 1}, math.Sqrt2},
		{"both-negative", WarmupPoint{X: -3, Y: -4}, 5},
		{"large", WarmupPoint{X: 300, Y: 400}, 500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.p.DistanceFromOrigin()
			if math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("WarmupPoint{%g, %g}.DistanceFromOrigin() = %g, want %g (tolerance 1e-9)",
					tc.p.X, tc.p.Y, got, tc.want)
			}
		})
	}
}
