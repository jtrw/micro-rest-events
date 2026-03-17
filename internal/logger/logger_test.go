package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestNewHandler_Enabled(t *testing.T) {
	h := NewHandler(&bytes.Buffer{}, slog.LevelWarn)
	if h.Enabled(context.TODO(), slog.LevelDebug) {
		t.Error("DEBUG should be disabled for WarnLevel handler")
	}
	if h.Enabled(context.TODO(), slog.LevelInfo) {
		t.Error("INFO should be disabled for WarnLevel handler")
	}
	if !h.Enabled(context.TODO(), slog.LevelWarn) {
		t.Error("WARN should be enabled")
	}
	if !h.Enabled(context.TODO(), slog.LevelError) {
		t.Error("ERROR should be enabled")
	}
}

func TestNewDebugHandler_Enabled(t *testing.T) {
	h := NewDebugHandler(&bytes.Buffer{})
	if !h.Enabled(context.TODO(), slog.LevelDebug) {
		t.Error("DEBUG should be enabled for debug handler")
	}
}

func TestHandle_OutputFormat(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, slog.LevelInfo)
	logger := slog.New(h)

	logger.Info("hello world", "key", "value", "num", 42)

	out := buf.String()
	if !strings.Contains(out, "[INFO]") {
		t.Errorf("expected [INFO] in output, got: %s", out)
	}
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected message in output, got: %s", out)
	}
	if !strings.Contains(out, "key=value") {
		t.Errorf("expected key=value in output, got: %s", out)
	}
	if !strings.Contains(out, "num=42") {
		t.Errorf("expected num=42 in output, got: %s", out)
	}
}

func TestHandle_StringWithSpacesQuoted(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, slog.LevelInfo)
	logger := slog.New(h)

	logger.Info("msg", "phrase", "hello world")

	out := buf.String()
	if !strings.Contains(out, `phrase="hello world"`) {
		t.Errorf("expected quoted value, got: %s", out)
	}
}

func TestHandle_StringWithoutSpaceNotQuoted(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, slog.LevelInfo)
	logger := slog.New(h)

	logger.Info("msg", "key", "simple")

	out := buf.String()
	if strings.Contains(out, `"simple"`) {
		t.Errorf("expected unquoted value, got: %s", out)
	}
	if !strings.Contains(out, "key=simple") {
		t.Errorf("expected key=simple, got: %s", out)
	}
}

func TestHandle_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, slog.LevelError)
	logger := slog.New(h)

	logger.Info("should not appear")
	logger.Warn("should not appear")
	logger.Error("should appear")

	out := buf.String()
	if strings.Contains(out, "should not appear") {
		t.Errorf("filtered levels should not appear in output, got: %s", out)
	}
	if !strings.Contains(out, "should appear") {
		t.Errorf("ERROR level should appear in output, got: %s", out)
	}
}

func TestHandle_TimeFormat(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, slog.LevelInfo)

	r := slog.NewRecord(time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC), slog.LevelInfo, "msg", 0)
	_ = h.Handle(context.TODO(), r)

	out := buf.String()
	if !strings.Contains(out, "2024/01/15 10:30:45") {
		t.Errorf("expected formatted time, got: %s", out)
	}
}

func TestWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, slog.LevelInfo)
	h2 := h.WithAttrs([]slog.Attr{slog.String("service", "myapp")})
	logger := slog.New(h2)

	logger.Info("started")

	out := buf.String()
	if !strings.Contains(out, "service=myapp") {
		t.Errorf("expected pre-attached attr in output, got: %s", out)
	}
}

func TestWithAttrs_PreservesParent(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, slog.LevelInfo)
	h.WithAttrs([]slog.Attr{slog.String("a", "1")})

	var buf2 bytes.Buffer
	h2 := NewHandler(&buf2, slog.LevelInfo)
	h2 = h2.WithAttrs([]slog.Attr{slog.String("x", "1")})
	h3 := h2.WithAttrs([]slog.Attr{slog.String("y", "2")})

	logger := slog.New(h3)
	logger.Info("check")

	out := buf2.String()
	if !strings.Contains(out, "x=1") || !strings.Contains(out, "y=2") {
		t.Errorf("expected merged attrs, got: %s", out)
	}
}

func TestWithGroup(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, slog.LevelInfo)
	// WithGroup is a no-op, just verify it returns a handler
	h2 := h.WithGroup("mygroup")
	if h2 == nil {
		t.Error("WithGroup should return non-nil handler")
	}
}

func TestNewDebugHandler_OutputContainsLevel(t *testing.T) {
	var buf bytes.Buffer
	h := NewDebugHandler(&buf)
	logger := slog.New(h)

	logger.Error("boom", "code", 500)

	out := buf.String()
	// Colors are stripped in non-TTY, but text should still be present
	if !strings.Contains(out, "boom") {
		t.Errorf("expected message in debug output, got: %s", out)
	}
	if !strings.Contains(out, "code=500") {
		t.Errorf("expected attr in debug output, got: %s", out)
	}
}

func TestContainsSpace(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"hello", false},
		{"hello world", true},
		{"tab\there", true},
		{"new\nline", true},
		{"", false},
	}
	for _, tt := range tests {
		got := containsSpace(tt.input)
		if got != tt.want {
			t.Errorf("containsSpace(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
