package config

import "testing"

// TestLoad is a SKELETON. Cover precedence: defaults; env over default;
// flag over env; file over default; env over file; flag over file; an
// invalid level errors. Plus a redaction test: logging a Config with an
// AuthToken must not leak it (shows "***").
func TestLoad(t *testing.T) {
	// TODO: env := func(m map[string]string) func(string) string { ... }
	//       cfg, err := Load(nil, env(nil)); assert defaults
	_ = Load
}
