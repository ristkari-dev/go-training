package logparse

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func mkTime(s string) time.Time {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestParse(t *testing.T) {
	t.Run("single-entry", func(t *testing.T) {
		in := "2026-05-21T14:30:00 INFO connection accepted from 192.168.1.5"
		got, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		want := []LogEntry{
			{Time: mkTime("2026-05-21T14:30:00"), Level: "INFO", Message: "connection accepted from 192.168.1.5"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Parse = %+v, want %+v", got, want)
		}
	})

	t.Run("multiple-entries", func(t *testing.T) {
		in := strings.Join([]string{
			"2026-05-21T14:30:00 INFO server started",
			"2026-05-21T14:31:00 WARN high memory usage",
			"2026-05-21T14:32:00 ERROR connection refused",
		}, "\n")
		got, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("got %d entries, want 3", len(got))
		}
		levels := []string{got[0].Level, got[1].Level, got[2].Level}
		want := []string{"INFO", "WARN", "ERROR"}
		if !reflect.DeepEqual(levels, want) {
			t.Errorf("levels = %v, want %v", levels, want)
		}
	})

	t.Run("blank-lines-skipped", func(t *testing.T) {
		in := "\n2026-05-21T14:30:00 INFO a\n\n2026-05-21T14:31:00 WARN b\n"
		got, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d entries (blank lines should be skipped), want 2", len(got))
		}
	})

	t.Run("malformed-line-returns-partial-and-error", func(t *testing.T) {
		in := strings.Join([]string{
			"2026-05-21T14:30:00 INFO ok",
			"garbage not a log line",
			"2026-05-21T14:32:00 WARN also ok",
		}, "\n")
		out, err := Parse(strings.NewReader(in))
		if err == nil {
			t.Fatal("expected error from malformed line")
		}
		if !strings.Contains(err.Error(), "line 2") {
			t.Errorf("error should mention line 2, got %v", err)
		}
		if len(out) != 1 {
			t.Errorf("expected 1 partial entry from line 1, got %d", len(out))
		}
	})
}

func TestCountByLevel(t *testing.T) {
	t.Run("mixed", func(t *testing.T) {
		entries := []LogEntry{
			{Level: "INFO"},
			{Level: "WARN"},
			{Level: "INFO"},
			{Level: "ERROR"},
			{Level: "INFO"},
		}
		got := CountByLevel(entries)
		want := map[string]int{"INFO": 3, "WARN": 1, "ERROR": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CountByLevel = %v, want %v", got, want)
		}
	})

	t.Run("all-same", func(t *testing.T) {
		entries := []LogEntry{
			{Level: "INFO"}, {Level: "INFO"}, {Level: "INFO"},
		}
		got := CountByLevel(entries)
		want := map[string]int{"INFO": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("CountByLevel = %v, want %v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := CountByLevel(nil)
		if len(got) != 0 {
			t.Errorf("CountByLevel(nil) = %v, want empty", got)
		}
	})
}

func TestParseLine(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		e, err := ParseLine("2026-05-21T14:30:00 INFO hello world")
		if err != nil {
			t.Fatalf("ParseLine: %v", err)
		}
		if e.Level != "INFO" || e.Message != "hello world" {
			t.Errorf("ParseLine = %+v", e)
		}
	})
	for _, bad := range []string{"", "0", "garbage", "2026-05-21T14:30:00 DEBUG x", "2026-13-02T15:04:05 INFO x"} {
		if _, err := ParseLine(bad); err == nil {
			t.Errorf("ParseLine(%q) = nil error, want error", bad)
		}
	}
}
