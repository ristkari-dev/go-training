// Package slicesx is a small slices-style helper library. It demonstrates
// Go generics: type parameters, unconstrained type parameters (any),
// multiple type parameters, and type inference at the call site.
//
// In real code you'd use the stdlib slices package (Go 1.21+) — it has
// Filter (via custom code: slices.DeleteFunc with inversion) and there's
// an "x/exp/slices.Map" community variant. We reimplement them here to
// teach the generics machinery from scratch.
//
// Both functions are unconstrained (T any, U any). The lesson teaches
// that MOST generic functions need only `any` — only operations that
// need <, ==, etc. require richer constraints (see maxg.Max).
package slicesx

// Filter returns a new slice containing the elements of xs for which
// pred returns true. The output preserves input order.
//
// The output slice is pre-allocated with capacity len(xs) — best case
// (no matches) wastes len(xs) ints; worst case (all match) does zero
// re-allocations. Reasonable tradeoff for unknown match rate.
//
// Examples:
//
//	Filter([]int{1, 2, 3, 4}, func(x int) bool { return x%2 == 0 }) → [2, 4]
//	Filter([]string{"a", "bb", "ccc"}, func(s string) bool { return len(s) > 1 }) → ["bb", "ccc"]
//
// Type inference: callers write Filter(xs, pred) — the compiler infers T
// from the element type of xs.
//
// Hint:
//  1. out := make([]T, 0, len(xs))
//  2. for _, x := range xs → if pred(x) → out = append(out, x)
//  3. return out
func Filter[T any](xs []T, pred func(T) bool) []T {
	panic("TODO: pre-allocate with cap len(xs); append elements where pred is true")
}

// Map returns a new slice where each element is the result of applying
// f to the corresponding input element. The output type U is
// (potentially) different from the input type T.
//
// This is where multiple type parameters earn their keep: the function
// has one input type and one output type, and they can be different.
//
// Examples:
//
//	Map([]int{1, 2, 3}, strconv.Itoa)              → ["1", "2", "3"]   // T=int, U=string
//	Map([]string{"hi", "world"}, strings.ToUpper)  → ["HI", "WORLD"]   // T=U=string
//	Map([]string{"hi", "world"}, len)              → [2, 5]             // T=string, U=int
//
// Output is pre-allocated to exactly len(xs) — we know the output size
// in advance.
//
// Type inference: callers write Map(xs, f) — both T and U are inferred
// from the function signature.
//
// Hint:
//  1. out := make([]U, len(xs))
//  2. for i, x := range xs → out[i] = f(x)
//  3. return out
func Map[T, U any](xs []T, f func(T) U) []U {
	panic("TODO: pre-allocate with size len(xs); transform each element")
}
