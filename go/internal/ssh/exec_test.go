package ssh

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"ssh-skill/internal/types"
)

// shellCommand returns an *exec.Cmd that runs the given shell command string
// on the local OS. Windows uses cmd.exe /c; POSIX uses /bin/sh -c. Both
// support `echo` and `exit N` semantics identically for the assertions used
// in these tests.
func shellCommand(cmdLine string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd.exe", "/c", cmdLine)
	}
	return exec.Command("/bin/sh", "-c", cmdLine)
}

// startTestSSHServer starts an in-process SSH server that authenticates with
// password "testpass" and runs each requested command via /bin/sh -c.
// Returns the listen address and a stop function. The server accepts one
// session at a time; tests are short-lived.
//
// This avoids any external Docker dependency while exercising the full
// SSH handshake, session.Run, stdout/stderr capture and ExitError path.
func startTestSSHServer(t *testing.T) (addr string, stop func()) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate ed25519 host key: %v", err)
	}
	_ = pub

	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "testuser" && string(pass) == "testpass" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			nConn, err := listener.Accept()
			if err != nil {
				return // listener closed
			}
			go func(conn net.Conn) {
				defer conn.Close()
				sconn, chans, reqs, err := ssh.NewServerConn(conn, config)
				if err != nil {
					return
				}
				defer sconn.Close()
				go ssh.DiscardRequests(reqs)
				for newChannel := range chans {
					if newChannel.ChannelType() != "session" {
						_ = newChannel.Reject(ssh.UnknownChannelType, "only session")
						continue
					}
					channel, requests, err := newChannel.Accept()
					if err != nil {
						continue
					}
					go func(ch ssh.Channel, reqs <-chan *ssh.Request) {
						defer ch.Close()
						for req := range reqs {
							if req.Type == "exec" {
								// SSH exec request payload format: uint32 big-endian
								// command length, followed by command bytes.
								if len(req.Payload) < 4 {
									_ = req.Reply(false, nil)
									continue
								}
								cmdLen := int(req.Payload[0])<<24 | int(req.Payload[1])<<16 | int(req.Payload[2])<<8 | int(req.Payload[3])
								if 4+cmdLen > len(req.Payload) {
									_ = req.Reply(false, nil)
									continue
								}
								cmdBytes := req.Payload[4 : 4+cmdLen]
								_ = req.Reply(true, nil)

								// Run the requested command on the test host OS.
								// Bound to 127.0.0.1; commands come from the test itself.
								cmd := shellCommand(string(cmdBytes))
								var stdoutBuf, stderrBuf bytes.Buffer
								cmd.Stdout = &stdoutBuf
								cmd.Stderr = &stderrBuf
								err := cmd.Run()
								_, _ = ch.Write(stdoutBuf.Bytes())
								_, _ = ch.Stderr().Write(stderrBuf.Bytes())
								exitCode := 0
								if err != nil {
									if exitErr, ok := err.(*exec.ExitError); ok {
										exitCode = exitErr.ExitCode()
									} else {
										exitCode = 1
									}
								}
								_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Code uint32 }{uint32(exitCode)}))
								return
							}
							_ = req.Reply(false, nil)
						}
					}(channel, requests)
				}
			}(nConn)
		}
	}()

	stopFn := func() {
		_ = listener.Close()
		close(done)
	}
	return listener.Addr().String(), stopFn
}

func TestExecPasswordAuthSuccess(t *testing.T) {
	addr, stop := startTestSSHServer(t)
	defer stop()

	cfg := &types.ServerConfig{
		ID:   "test",
		Host: "127.0.0.1",
		Port: portFromAddr(addr),
		User: "testuser",
		Auth: types.AuthConfig{
			Method:            types.AuthPassword,
			EncryptedPassword: "testpass", // already-decrypted plaintext (cli layer responsibility)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Happy path: command succeeds with stdout.
	result, err := Exec(ctx, cfg, "echo hello-ssh-skill", 5*time.Second)
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0; stderr=%q", result.ExitCode, result.Stderr)
	}
	if result.Stdout != "hello-ssh-skill\n" && result.Stdout != "hello-ssh-skill\r\n" {
		t.Errorf("Stdout = %q, want \"hello-ssh-skill\\n\"", result.Stdout)
	}
}

func TestExecExitCodePropagation(t *testing.T) {
	addr, stop := startTestSSHServer(t)
	defer stop()

	cfg := &types.ServerConfig{
		ID:   "test",
		Host: "127.0.0.1",
		Port: portFromAddr(addr),
		User: "testuser",
		Auth: types.AuthConfig{
			Method:            types.AuthPassword,
			EncryptedPassword: "testpass",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Command exits with non-zero code; we must NOT treat this as a Go error
	// and ExitCode must equal the remote exit status.
	result, err := Exec(ctx, cfg, "exit 42", 5*time.Second)
	if err != nil {
		t.Fatalf("Exec returned err for non-zero exit: %v (should be nil; caller inspects ExitCode)", err)
	}
	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42 (regression: previously -1 due to wrong type assertion)", result.ExitCode)
	}
}

func TestExecAuthFailureReturnsError(t *testing.T) {
	addr, stop := startTestSSHServer(t)
	defer stop()

	cfg := &types.ServerConfig{
		ID:   "test",
		Host: "127.0.0.1",
		Port: portFromAddr(addr),
		User: "testuser",
		Auth: types.AuthConfig{
			Method:            types.AuthPassword,
			EncryptedPassword: "wrong-password",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := Exec(ctx, cfg, "echo should-not-run", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if result == nil {
		t.Fatal("expected non-nil result even on auth failure")
	}
	if result.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1 on auth failure", result.ExitCode)
	}
}

func TestExecNilCfgServerIDPreserved(t *testing.T) {
	// Defensive: even if Connect fails, ServerID/Command must be set on result
	// so the audit log is still writable. Tests the I3 nil-guard indirectly.
	cfg := &types.ServerConfig{
		ID:   "unreachable",
		Host: "127.0.0.1",
		Port: 1, // unroutable port
		User: "x",
		Auth: types.AuthConfig{
			Method:            types.AuthPassword,
			EncryptedPassword: "x",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := Exec(ctx, cfg, "uptime", 1*time.Second)
	if err == nil {
		t.Fatal("expected error for unroutable host")
	}
	if result == nil {
		t.Fatal("result must not be nil even on Connect failure")
	}
	if result.ServerID != "unreachable" || result.Command != "uptime" {
		t.Errorf("result identity fields not preserved: %+v", result)
	}
	if result.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1 on connect failure", result.ExitCode)
	}
}

func portFromAddr(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return 0
	}
	return port
}
