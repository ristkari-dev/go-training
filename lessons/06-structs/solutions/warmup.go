// Package solutions is the reference implementation for lesson 06: Structs and methods.
//
// This file holds the warm-up reference solution.
package solutions

import "math"

// WarmupPoint is a 2D point with float64 coordinates.
type WarmupPoint struct {
	X float64
	Y float64
}

// DistanceFromOrigin returns sqrt(X*X + Y*Y).
func (p WarmupPoint) DistanceFromOrigin() float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y)
}
