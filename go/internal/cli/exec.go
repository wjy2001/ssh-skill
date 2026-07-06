package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ssh-mcp/internal/ssh"
	"ssh-mcp/internal/types"
)

func cmdExec(args []string) error {
	fs := flag.NewFlagSet("exec", flag.ExitOnError)
	serverID := fs.String("server", "", "Server identifier")
	command := fs.String("command", "", "Command to execute")
	timeoutSec := fs.Int("timeout", 30, "Command timeout in seconds")
	fs.Parse(args)

	if *serverID == "" || *command == "" {
		fmt.Fprintln(os.Stderr, "Required flags: --server, --command")
		fs.Usage()
		return fmt.Errorf("missing required flags")
	}

	app, err := Load()
	if err != nil {
		return err
	}

	// Target validation: server must exist in vault.
	cfg, err := ssh.FindServer(app.Vault, *serverID)
	if err != nil {
		return fmt.Errorf("server '%s' not found in local configuration", *serverID)
	}

	// Decrypt password if needed.
	if cfg.Auth.Method == types.AuthPassword && cfg.Auth.EncryptedPassword != "" {
		decrypted, err := decryptField(app.VaultKey, cfg.Auth.EncryptedPassword)
		if err != nil {
			return fmt.Errorf("decrypt password: %w", err)
		}
		cfg.Auth.EncryptedPassword = decrypted // in-memory plaintext for SSH connection
	}

	ctx := context.Background()
	timeout := time.Duration(*timeoutSec) * time.Second

	result, err := ssh.Exec(ctx, app.Vault, *serverID, *command, timeout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// Write audit log entry.
	auditEntry := &types.AuditEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ServerID:   result.ServerID,
		ServerHost: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Command:    result.Command,
		ExitCode:   result.ExitCode,
		StdoutLen:  len(result.Stdout),
		StderrLen:  len(result.Stderr),
		DurationMs: result.DurationMs,
	}
	if auditErr := app.AuditLog.Log(auditEntry); auditErr != nil {
		fmt.Fprintf(os.Stderr, "warning: audit log write failed: %v\n", auditErr)
	}

	// Print output.
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	return err
}
