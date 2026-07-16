package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ssh-skill/internal/ssh"
)

func cmdDownload(args []string) error {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	serverID := fs.String("server", "", "Server identifier")
	remotePath := fs.String("remote", "", "Remote file path")
	localPath := fs.String("local", "", "Local file path")
	fs.Parse(args)

	if *serverID == "" || *remotePath == "" || *localPath == "" {
		fmt.Fprintln(os.Stderr, "Required flags: --server, --remote, --local")
		return fmt.Errorf("missing required flags")
	}

	app, err := Load()
	if err != nil {
		return err
	}

	cfg, err := resolveServer(app, *serverID)
	if err != nil {
		return err
	}

	// Show progress bar during transfer.
	result, err := ssh.Download(context.Background(), cfg, *remotePath, *localPath,
		func(transferred, total int64, elapsed time.Duration) {
			renderProgress("Downloading", transferred, total, elapsed)
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n")
		return err
	}

	finishProgress("Downloading", result.SizeBytes, time.Duration(result.DurationMs)*time.Millisecond)
	fmt.Printf("Downloaded %d bytes from %s:%s (%dms)\n", result.SizeBytes, result.ServerID, result.Path, result.DurationMs)
	return nil
}
