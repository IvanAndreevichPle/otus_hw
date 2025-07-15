package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

// Global mutex to prevent race conditions in captureOutput
var captureMutex sync.Mutex

func captureOutput(f func()) string {
	captureMutex.Lock()
	defer captureMutex.Unlock()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// TestLoggerLevels проверяет, что для каждого уровня логирования выводятся только нужные сообщения.
func TestLoggerLevels(t *testing.T) {
	cases := []struct {
		level   string
		calls   []func(l *Logger)
		expects []string
	}{
		{"error", []func(l *Logger){func(l *Logger) { l.Error("err") }, func(l *Logger) { l.Warn("warn") }, func(l *Logger) { l.Info("info") }, func(l *Logger) { l.Debug("debug") }}, []string{"[ERROR] err\n"}},
		{"warn", []func(l *Logger){func(l *Logger) { l.Error("err") }, func(l *Logger) { l.Warn("warn") }, func(l *Logger) { l.Info("info") }, func(l *Logger) { l.Debug("debug") }}, []string{"[ERROR] err\n", "[WARN] warn\n"}},
		{"info", []func(l *Logger){func(l *Logger) { l.Error("err") }, func(l *Logger) { l.Warn("warn") }, func(l *Logger) { l.Info("info") }, func(l *Logger) { l.Debug("debug") }}, []string{"[ERROR] err\n", "[WARN] warn\n", "[INFO] info\n"}},
		{"debug", []func(l *Logger){func(l *Logger) { l.Error("err") }, func(l *Logger) { l.Warn("warn") }, func(l *Logger) { l.Info("info") }, func(l *Logger) { l.Debug("debug") }}, []string{"[ERROR] err\n", "[WARN] warn\n", "[INFO] info\n", "[DEBUG] debug\n"}},
	}
	for _, c := range cases {
		l := New(c.level)
		out := captureOutput(func() {
			for _, call := range c.calls {
				call(l)
			}
		})
		lines := strings.Split(out, "\n")
		var filtered []string
		for _, line := range lines {
			if line != "" {
				filtered = append(filtered, line+"\n")
			}
		}
		if len(filtered) != len(c.expects) {
			t.Errorf("level %s: expected %d lines, got %d", c.level, len(c.expects), len(filtered))
			continue
		}
		for i, exp := range c.expects {
			if filtered[i] != exp {
				t.Errorf("level %s: expected %q, got %q", c.level, exp, filtered[i])
			}
		}
	}
}

// TestLoggerParseLevelDefault проверяет, что неизвестный уровень по умолчанию становится info.
func TestLoggerParseLevelDefault(t *testing.T) {
	l := New("unknown")
	out := captureOutput(func() { l.Info("test") })
	if !strings.Contains(out, "[INFO] test") {
		t.Errorf("expected default level to be info")
	}
}

// TestLoggerCaseInsensitiveLevel проверяет, что уровень логирования нечувствителен к регистру.
func TestLoggerCaseInsensitiveLevel(t *testing.T) {
	l := New("InFo")
	out := captureOutput(func() { l.Info("case") })
	if !strings.Contains(out, "[INFO] case") {
		t.Errorf("expected case-insensitive level to work")
	}
}

// TestLoggerEmptyMessage проверяет, что даже для пустого сообщения выводится префикс уровня.
func TestLoggerEmptyMessage(t *testing.T) {
	l := New("debug")
	out := captureOutput(func() { l.Info("") })
	if !strings.Contains(out, "[INFO]") {
		t.Errorf("expected prefix even for empty message")
	}
}

// TestLoggerNoOutputForLowerLevels проверяет, что сообщения ниже текущего уровня не выводятся.
func TestLoggerNoOutputForLowerLevels(t *testing.T) {
	l := New("error")
	out := captureOutput(func() { l.Warn("should not print") })
	if out != "" {
		t.Errorf("expected no output for lower level, got: %q", out)
	}
}

// TestLoggerConcurrency проверяет, что логгер не паникует при конкурентном использовании.
func TestLoggerConcurrency(t *testing.T) {
	l := New("debug")
	var wg sync.WaitGroup
	calls := 100
	wg.Add(calls)
	for i := 0; i < calls; i++ {
		go func(i int) {
			defer wg.Done()
			captureOutput(func() { l.Info(fmt.Sprintf("msg %d", i)) })
		}(i)
	}
	wg.Wait()
}
