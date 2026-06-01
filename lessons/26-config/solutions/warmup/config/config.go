// Package config loads logstatsd configuration with precedence:
// defaults < JSON file < environment < explicitly-set flags.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	HTTPAddr  string `json:"http_addr"`
	GRPCAddr  string `json:"grpc_addr"`
	LogLevel  string `json:"log_level"`
	AuthToken string `json:"-"`
}

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

func Load(args []string, getenv func(string) string) (Config, error) {
	cfg := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", LogLevel: "info"} // defaults

	fs := flag.NewFlagSet("logstatsd", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON config file")
	httpAddr := fs.String("http-addr", "", "HTTP listen address")
	grpcAddr := fs.String("grpc-addr", "", "gRPC listen address")
	logLevel := fs.String("log-level", "", "log level (debug|info|warn|error)")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	// Layer 1: JSON config file (from -config flag or LOGSTATS_CONFIG env).
	path := *configPath
	if path == "" {
		path = getenv("LOGSTATS_CONFIG")
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("config: read %s: %w", path, err)
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("config: parse %s: %w", path, err)
		}
	}

	// Layer 2: environment overrides.
	if v := getenv("LOGSTATS_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := getenv("LOGSTATS_GRPC_ADDR"); v != "" {
		cfg.GRPCAddr = v
	}
	if v := getenv("LOGSTATS_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := getenv("LOGSTATS_AUTH_TOKEN"); v != "" {
		cfg.AuthToken = v
	}

	// Layer 3: explicitly-set flags win (fs.Visit skips defaulted flags).
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "http-addr":
			cfg.HTTPAddr = *httpAddr
		case "grpc-addr":
			cfg.GRPCAddr = *grpcAddr
		case "log-level":
			cfg.LogLevel = *logLevel
		}
	})

	if !validLevel(cfg.LogLevel) {
		return Config{}, fmt.Errorf("config: invalid log level %q", cfg.LogLevel)
	}
	return cfg, nil
}
