package ssh

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"

	"ssh-mcp/internal/types"
)

// Exec runs a command on the remote server and returns the result.
// It handles target validation (server must exist in the vault) and timeout control.
func Exec(ctx context.Context, vault *types.Vault, serverID, command string, timeout time.Duration) (*types.ExecResult, error) {
	cfg, err := FindServer(vault, serverID)
	if err != nil {
		return nil, err
	}

	result := &types.ExecResult{
		ServerID: serverID,
		Command:  command,
		ExitCode: -1,
	}

	start := time.Now()

	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	connCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := Connect(connCtx, cfg)
	if err != nil {
		return result, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return result, fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	result.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		if exitErr, ok := err.(*ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
		}
		// err is set but we still return the result — the caller inspects ExitCode.
	}

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// If no exit error was detected but err is set, return it.
	if result.ExitCode == -1 && err != nil {
		result.Stderr += fmt.Sprintf("\n[ssh-mcp] %v", err)
	}

	return result, nil
}

// ExitError wraps ssh.ExitError for consistent type assertions.
type ExitError struct {
	*ssh.ExitError
}

// ExitStatus returns the exit code from the remote command.
func (e *ExitError) ExitStatus() int {
	return e.ExitError.ExitStatus()
}
