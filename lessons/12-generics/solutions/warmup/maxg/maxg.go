// Package maxg is the lesson 12 warm-up reference implementation.
package maxg

import (
	"cmp"
	"errors"
)

// Max returns the largest element of xs (cmp.Ordered constraint).
func Max[T cmp.Ordered](xs []T) (T, error) {
	var zero T
	if len(xs) == 0 {
		return zero, errors.New("maxg: empty slice")
	}
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m, nil
}
