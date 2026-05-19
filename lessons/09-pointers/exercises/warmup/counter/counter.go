// Package counter is the lesson 09 warm-up: a tiny demo of value-receiver
// vs pointer-receiver semantics.
//
// Counter has one unexported field `n` (the running count). It exposes:
//   - Inc()         — POINTER receiver. Mutates the underlying value.
//   - Value() int   — VALUE receiver. Reads the count; can't mutate.
//
// The whole lesson is: "if you want a method to mutate the receiver,
// declare it with a pointer receiver. Value receivers get a COPY."
package counter

// Counter holds an int count. The field is unexported — the only way to
// read it from outside this package is via Value().
type Counter struct {
	n int
}

// Inc increments the counter by 1.
//
// Pointer receiver — the call site's Counter is mutated. After
// `c.Inc()`, the caller's `c.Value()` will return 1 more than before.
//
// Examples:
//
//	var c Counter
//	c.Inc()
//	c.Inc()
//	c.Value()   → 2
//
// Hint: this is a one-liner. `c.n++`.
func (c *Counter) Inc() {
	panic("TODO: increment c.n by 1")
}

// Value returns the current count.
//
// Value receiver — it can READ c.n but mutations to c here would be
// invisible to the caller (because c is a copy). For a read-only method,
// value receivers are the conventional choice.
//
// Examples:
//
//	var c Counter
//	c.Value()   → 0
//	c.Inc()
//	c.Value()   → 1
//
// Hint: one-liner — return c.n.
func (c Counter) Value() int {
	panic("TODO: return c.n")
}
