package buildinfo

import (
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	if Version != "dev" || Commit != "none" || Date != "unknown" {
		t.Errorf("defaults = %q / %q / %q", Version, Commit, Date)
	}
}

func TestString(t *testing.T) {
	s := String()
	for _, want := range []string{"logstatsd", "dev", "none", "unknown"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() = %q, missing %q", s, want)
		}
	}
}
