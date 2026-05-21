// Package slicesx is the lesson 12 main package reference implementation.
package slicesx

// Filter returns a new slice containing the elements of xs for which pred
// returns true. Preserves input order. Pre-allocates with cap len(xs).
func Filter[T any](xs []T, pred func(T) bool) []T {
	out := make([]T, 0, len(xs))
	for _, x := range xs {
		if pred(x) {
			out = append(out, x)
		}
	}
	return out
}

// Map returns a new slice where each element is f applied to the
// corresponding input element. Pre-allocates with exact size len(xs).
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}
