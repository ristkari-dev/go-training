package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadPrecedence(t *testing.T) {
	cfg, err := Load(nil, env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.GRPCAddr != ":9090" || cfg.LogLevel != "info" {
		t.Errorf("defaults wrong: %+v", cfg)
	}

	cfg, _ = Load(nil, env(map[string]string{"LOGSTATS_HTTP_ADDR": ":7000", "LOGSTATS_LOG_LEVEL": "debug"}))
	if cfg.HTTPAddr != ":7000" || cfg.LogLevel != "debug" {
		t.Errorf("env override failed: %+v", cfg)
	}

	cfg, _ = Load([]string{"-http-addr=:6000"}, env(map[string]string{"LOGSTATS_HTTP_ADDR": ":7000"}))
	if cfg.HTTPAddr != ":6000" {
		t.Errorf("flag should beat env: %+v", cfg)
	}
}

func TestLoadFileLayer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(path, []byte(`{"http_addr":":5000","grpc_addr":":5001","log_level":"warn"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load([]string{"-config=" + path}, env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":5000" || cfg.GRPCAddr != ":5001" || cfg.LogLevel != "warn" {
		t.Errorf("file layer failed: %+v", cfg)
	}

	cfg, _ = Load([]string{"-config=" + path, "-log-level=error"},
		env(map[string]string{"LOGSTATS_GRPC_ADDR": ":5999"}))
	if cfg.HTTPAddr != ":5000" {
		t.Errorf("file value lost: %+v", cfg)
	}
	if cfg.GRPCAddr != ":5999" {
		t.Errorf("env should beat file: %+v", cfg)
	}
	if cfg.LogLevel != "error" {
		t.Errorf("flag should beat file: %+v", cfg)
	}
}

func TestLoadInvalidLevel(t *testing.T) {
	if _, err := Load([]string{"-log-level=verbose"}, env(nil)); err == nil {
		t.Error("expected error for invalid level")
	}
}

func TestConfigRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	cfg := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", LogLevel: "info", AuthToken: "super-secret-xyz"}
	logger.Info("config loaded", "config", cfg)

	out := buf.String()
	if strings.Contains(out, "super-secret-xyz") {
		t.Errorf("secret leaked in log: %s", out)
	}
	if !strings.Contains(out, "***") {
		t.Errorf("redaction marker missing: %s", out)
	}
}
