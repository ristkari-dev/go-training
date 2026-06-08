// Package buildinfo holds build metadata stamped at link time via
// -ldflags "-X". The defaults apply to a plain `go build`/`go run`; a
// release build overrides them (see the Dockerfile and the Makefile
// `build` target).
package buildinfo

import "fmt"

var (
	Version = "dev"     // -ldflags -X .../buildinfo.Version=...
	Commit  = "none"    // -ldflags -X .../buildinfo.Commit=...
	Date    = "unknown" // -ldflags -X .../buildinfo.Date=...
)

// String renders a one-line version banner:
//
//	logstatsd <version> (commit <commit>, built <date>)
//
// Hint: fmt.Sprintf("logstatsd %s (commit %s, built %s)", Version, Commit, Date)
func String() string {
	_ = fmt.Sprintf
	panic("TODO: return a one-line banner with Version, Commit, Date")
}
