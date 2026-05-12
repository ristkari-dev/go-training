// Package exercises is the starter code for lesson 06: Structs and methods.
//
// This file holds the WARM-UP exercise: a small struct with one method,
// to build muscle memory for "type T struct { ... }" and "func (t T) M()".
// The matching test ships as a skeleton — fill in the cases and assertion body.
package exercises

import "math"

// WarmupPoint is a 2D point with float64 coordinates.
//
// We use the `Warmup` prefix to keep this type out of the way of the
// main exercise's Expense struct. Real Go code would just call this Point.
type WarmupPoint struct {
	X float64
	Y float64
}

// DistanceFromOrigin returns the Euclidean distance from the origin (0, 0)
// to p — i.e. sqrt(X*X + Y*Y).
//
// Examples:
//
//	WarmupPoint{X: 3, Y: 4}.DistanceFromOrigin()  → 5
//	WarmupPoint{X: 0, Y: 0}.DistanceFromOrigin()  → 0
//	WarmupPoint{X: 1, Y: 1}.DistanceFromOrigin()  → 1.4142135…
//
// Hint: use math.Sqrt. p.X * p.X + p.Y * p.Y is the squared distance.
func (p WarmupPoint) DistanceFromOrigin() float64 {
	_ = math.Sqrt // keep the import compiling until you use it
	panic("TODO: return math.Sqrt(p.X*p.X + p.Y*p.Y)")
}
