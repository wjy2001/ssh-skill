package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTempConfigDir points SSH_SKILL_CONFIG_DIR at a fresh temp dir for the
// duration of the test, so cli.Load() reads/writes an isolated vault.
func withTempConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("SSH_SKILL_CONFIG_DIR", dir)
	return dir
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestCmdAddPasswordServer(t *testing.T) {
	withTempConfigDir(t)

	err := cmdAdd([]string{
		"--id", "prod",
		"--name", "Production",
		"--host", "10.0.0.1",
		"--port", "2222",
		"--user", "deploy",
		"--auth-type", "password",
		"--password", "s3cret",
	})
	if err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}

	// Verify the vault file exists and is non-empty.
	dir := os.Getenv("SSH_SKILL_CONFIG_DIR")
	vaultPath := filepath.Join(dir, "servers.json.age")
	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("stat vault: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("vault file is empty; Save() did not write encrypted payload")
	}

	// Verify vault file is NOT plaintext JSON (must be encrypted bytes).
	raw, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if strings.Contains(string(raw), "s3cret") {
		t.Fatal("plaintext password leaked into vault file on disk")
	}
}

func TestCmdAddDuplicateIDRejected(t *testing.T) {
	withTempConfigDir(t)

	args := []string{"--id", "dup", "--host", "10.0.0.1", "--auth-type", "agent"}
	if err := cmdAdd(args); err != nil {
		t.Fatalf("first add: %v", err)
	}
	err := cmdAdd(args)
	if err == nil {
		t.Fatal("expected error for duplicate id, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("duplicate error message: %v", err)
	}
}

func TestCmdAddMissingRequiredFlags(t *testing.T) {
	withTempConfigDir(t)
	// Missing --host and --auth-type.
	err := cmdAdd([]string{"--id", "x"})
	if err == nil {
		t.Fatal("expected error for missing required flags, got nil")
	}
}

func TestCmdListEmptyAndPopulated(t *testing.T) {
	withTempConfigDir(t)

	// Empty vault: should print "No servers configured."
	out := captureStdout(t, func() {
		_ = cmdList(nil)
	})
	if !strings.Contains(out, "No servers configured") {
		t.Errorf("empty list output: %q", out)
	}

	// Add one server and verify list shows it.
	_ = cmdAdd([]string{
		"--id", "prod", "--name", "Production",
		"--host", "10.0.0.1", "--auth-type", "agent",
	})
	out = captureStdout(t, func() {
		_ = cmdList(nil)
	})
	if !strings.Contains(out, "prod") {
		t.Errorf("populated list missing id 'prod': %q", out)
	}
	if !strings.Contains(out, "10.0.0.1") {
		t.Errorf("populated list missing host: %q", out)
	}
}

func TestCmdRemove(t *testing.T) {
	withTempConfigDir(t)

	_ = cmdAdd([]string{"--id", "del-me", "--host", "10.0.0.1", "--auth-type", "agent"})

	err := cmdRemove([]string{"--id", "del-me"})
	if err != nil {
		t.Fatalf("cmdRemove: %v", err)
	}

	// Removing again should fail.
	err = cmdRemove([]string{"--id", "del-me"})
	if err == nil {
		t.Fatal("expected error when removing non-existent server, got nil")
	}
}

func TestCmdExecServerNotFound(t *testing.T) {
	withTempConfigDir(t)

	// No servers added; exec must report server not found, NOT nil-panic
	// while building the audit entry (regression: previously ssh.Exec returned
	// nil result and cli/exec.go accessed result.ServerID).
	err := cmdExec([]string{"--server", "ghost", "--command", "uptime"})
	if err == nil {
		t.Fatal("expected error for non-existent server, got nil")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error should mention server id 'ghost': %v", err)
	}
}

func TestCmdExecMissingFlags(t *testing.T) {
	withTempConfigDir(t)
	err := cmdExec([]string{"--server", "x"}) // missing --command
	if err == nil {
		t.Fatal("expected error for missing --command flag")
	}
}

func TestCmdVaultInit(t *testing.T) {
	dir := withTempConfigDir(t)

	out := captureStdout(t, func() {
		_ = cmdVault([]string{"init"})
	})
	if !strings.Contains(out, "Vault initialized") {
		t.Errorf("vault init output: %q", out)
	}

	// Verify key file and vault file created on disk.
	keyPath := filepath.Join(dir, ".vault-key")
	vaultPath := filepath.Join(dir, "servers.json.age")
	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("key file not created: %v", err)
	}
	if _, err := os.Stat(vaultPath); err != nil {
		t.Errorf("vault file not created: %v", err)
	}
}
