package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ssh-skill/internal/ssh"
)

func cmdTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	serverID := fs.String("server", "", "Server identifier")
	fs.Parse(args)

	if *serverID == "" {
		fmt.Fprintln(os.Stderr, "Required flag: --server")
		return fmt.Errorf("missing --server")
	}

	app, err := Load()
	if err != nil {
		return err
	}

	cfg, err := resolveServer(app, *serverID)
	if err != nil {
		return err
	}

	start := time.Now()
	client, err := ssh.Connect(context.Background(), cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection to %s (%s@%s:%d) FAILED: %v\n", cfg.ID, cfg.User, cfg.Host, cfg.Port, err)
		return err
	}
	client.Close()

	elapsed := time.Since(start)
	fmt.Printf("Connection to %s (%s@%s:%d) OK — %dms\n", cfg.ID, cfg.User, cfg.Host, cfg.Port, elapsed.Milliseconds())

	return nil
}
