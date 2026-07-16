package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"

	"ssh-skill/internal/types"
)

// Exec runs a command on the remote server described by cfg and returns the
// result. The cfg.Auth field MUST already contain decrypted credentials —
// password decryption is the caller's responsibility (cli layer). This avoids
// a double lookup against the raw (still-encrypted) vault.
//
// ExitCode handling: a non-zero remote exit produces *ssh.ExitError from
// session.Run(); we surface its ExitStatus() and still return the captured
// stdout/stderr. Other errors (connection, session creation) leave ExitCode
// at -1 and are also returned via err so the caller can distinguish them.
func Exec(ctx context.Context, cfg *types.ServerConfig, command string, timeout time.Duration) (*types.ExecResult, error) {
	result := &types.ExecResult{
		ServerID: cfg.ID,
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
		result.DurationMs = time.Since(start).Milliseconds()
		return result, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		result.DurationMs = time.Since(start).Milliseconds()
		return result, fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	runErr := session.Run(command)
	result.DurationMs = time.Since(start).Milliseconds()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Non-zero remote exit: extract the real exit code and DO NOT treat as
	// a Go-side error — the caller inspects ExitCode. We still return nil err
	// so the CLI can print stdout/stderr cleanly.
	var exitErr *ssh.ExitError
	if errors.As(runErr, &exitErr) {
		result.ExitCode = exitErr.ExitStatus()
		return result, nil
	}

	// Genuine error (connection lost, parse failure, etc.): keep ExitCode at -1
	// and surface the error to the caller.
	if runErr != nil {
		result.Stderr += fmt.Sprintf("\n[ssh-skill] %v", runErr)
		return result, runErr
	}

	// Clean success.
	result.ExitCode = 0
	return result, nil
}
