// Package config loads logstatsd configuration with precedence:
// defaults < JSON file < environment < explicitly-set flags.
package config

import (
	"flag"
	"fmt"
	"log/slog"
)

// Config holds the logstatsd runtime configuration.
type Config struct {
	HTTPAddr  string `json:"http_addr"`
	GRPCAddr  string `json:"grpc_addr"`
	LogLevel  string `json:"log_level"`
	AuthToken string `json:"-"` // secret: env-only, never logged
}

// LogValue implements slog.LogValuer so logging a Config never leaks the
// secret token — it shows "***" when set, "" when empty.
func (c Config) LogValue() slog.Value {
	token := ""
	if c.AuthToken != "" {
		token = "***"
	}
	return slog.GroupValue(
		slog.String("http_addr", c.HTTPAddr),
		slog.String("grpc_addr", c.GRPCAddr),
		slog.String("log_level", c.LogLevel),
		slog.String("auth_token", token),
	)
}

func validLevel(s string) bool {
	switch s {
	case "debug", "info", "warn", "error":
		return true
	}
	return false
}

// Load resolves configuration with precedence: defaults < JSON file <
// env < explicitly-set flags. getenv is injected for testability.
//
// Hint:
//  1. cfg := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", LogLevel: "info"} // defaults
//  2. fs := flag.NewFlagSet(...); define -config -http-addr -grpc-addr -log-level; fs.Parse(args)
//  3. file layer: path from -config or getenv("LOGSTATS_CONFIG"); os.ReadFile + json.Unmarshal into &cfg
//  4. env layer: LOGSTATS_HTTP_ADDR / _GRPC_ADDR / _LOG_LEVEL / _AUTH_TOKEN (non-empty overrides)
//  5. flag layer: fs.Visit(...) — only EXPLICITLY-SET flags override (not defaulted ones)
//  6. validate cfg.LogLevel; return error if invalid
func Load(args []string, getenv func(string) string) (Config, error) {
	_ = flag.NewFlagSet
	_ = fmt.Errorf
	_ = validLevel
	panic("TODO: defaults < file < env < explicitly-set flags; validate level")
}
