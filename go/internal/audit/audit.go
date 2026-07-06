// Package audit provides JSONL audit logging for ssh-mcp command executions.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"ssh-mcp/internal/types"
)

// Logger writes audit entries in JSONL format to a file.
// It is safe for concurrent use.
type Logger struct {
	mu   sync.Mutex
	path string
}

// NewLogger creates a new audit logger writing to the given file path.
func NewLogger(path string) (*Logger, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("audit: create dir %s: %w", dir, err)
	}
	return &Logger{path: path}, nil
}

// Log writes a single audit entry as a JSON line to the log file.
func (l *Logger) Log(entry *types.AuditEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("audit: open file %s: %w", l.path, err)
	}
	defer f.Close()

	line := append(data, '\n')
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("audit: write entry: %w", err)
	}

	return nil
}

// Path returns the path to the audit log file.
func (l *Logger) Path() string {
	return l.path
}
