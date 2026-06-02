// Package buildinfo holds build metadata stamped at link time via
// -ldflags "-X". Reference implementation.
package buildinfo

import "fmt"

var (
	Version = "dev"     // -ldflags -X .../buildinfo.Version=...
	Commit  = "none"    // -ldflags -X .../buildinfo.Commit=...
	Date    = "unknown" // -ldflags -X .../buildinfo.Date=...
)

// String renders a one-line version banner.
func String() string {
	return fmt.Sprintf("logstatsd %s (commit %s, built %s)", Version, Commit, Date)
}
