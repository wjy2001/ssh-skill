package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ssh-skill/internal/ssh"
)

func cmdUpload(args []string) error {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	serverID := fs.String("server", "", "Server identifier")
	localPath := fs.String("local", "", "Local file path")
	remotePath := fs.String("remote", "", "Remote file path")
	fs.Parse(args)

	if *serverID == "" || *localPath == "" || *remotePath == "" {
		fmt.Fprintln(os.Stderr, "Required flags: --server, --local, --remote")
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
	result, err := ssh.Upload(context.Background(), cfg, *localPath, *remotePath,
		func(transferred, total int64, elapsed time.Duration) {
			renderProgress("Uploading", transferred, total, elapsed)
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n")
		return err
	}

	finishProgress("Uploading", result.SizeBytes, time.Duration(result.DurationMs)*time.Millisecond)
	fmt.Printf("Uploaded %d bytes to %s:%s (%dms)\n", result.SizeBytes, result.ServerID, result.Path, result.DurationMs)
	return nil
}
