package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ssh-skill/internal/ssh"
	"ssh-skill/internal/types"
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

	// resolveServer looks up the server and decrypts the password in-place,
	// returning a ServerConfig ready for SSH connection. We pass this cfg
	// (not app.Vault) to ssh.Exec so the decrypted password is actually used.
	cfg, err := resolveServer(app, *serverID)
	if err != nil {
		return fmt.Errorf("server '%s' not found in local configuration", *serverID)
	}

	ctx := context.Background()
	timeout := time.Duration(*timeoutSec) * time.Second

	result, execErr := ssh.Exec(ctx, cfg, *command, timeout)

	// Defensive nil check: ssh.Exec only returns nil result on programmer error,
	// but guard against regressions so an audit entry is still written.
	if result == nil {
		result = &types.ExecResult{
			ServerID: *serverID,
			Command:  *command,
			ExitCode: -1,
		}
	}

	// Write audit log entry (best-effort; does not block command output).
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

	if execErr != nil {
		fmt.Fprintln(os.Stderr, execErr)
	}

	// Print output.
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	return execErr
}
