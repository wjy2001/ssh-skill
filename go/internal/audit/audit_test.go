package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ssh-skill/internal/types"
)

func TestLogWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	entry := &types.AuditEntry{
		Timestamp:  "2026-07-02T15:30:45Z",
		ServerID:   "prod-web",
		ServerHost: "10.0.1.100",
		Command:    "systemctl restart nginx",
		ExitCode:   0,
		StdoutLen:  1287,
		StderrLen:  0,
		DurationMs: 320,
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log: %v", err)
	}

	// Read back and verify.
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var parsed types.AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}

	if parsed.ServerID != entry.ServerID || parsed.Command != entry.Command {
		t.Fatalf("round-trip mismatch: %+v vs %+v", parsed, entry)
	}
}

func TestLogAppendsMultipleEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	for i := 0; i < 3; i++ {
		entry := &types.AuditEntry{
			Timestamp: "2026-07-02T15:30:45Z",
			ServerID:  "test",
			Command:   "cmd",
		}
		if err := logger.Log(entry); err != nil {
			t.Fatalf("Log entry %d: %v", i, err)
		}
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatalf("line %d is not valid JSON: %s", i, line)
		}
	}
}

func TestConcurrentLogging(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	const goroutines = 10
	done := make(chan bool, goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			entry := &types.AuditEntry{
				Timestamp: "2026-07-02T15:30:45Z",
				ServerID:  "test",
				Command:   "cmd",
			}
			if err := logger.Log(entry); err != nil {
				t.Errorf("goroutine %d: Log: %v", id, err)
			}
			done <- true
		}(g)
	}

	for g := 0; g < goroutines; g++ {
		<-done
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != goroutines {
		t.Fatalf("expected %d lines, got %d (possible data race)", goroutines, len(lines))
	}
}
