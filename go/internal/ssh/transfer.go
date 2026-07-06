package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"

	"ssh-mcp/internal/types"
)

// Upload copies a local file to the remote server via SFTP.
// The cfg must have its password already decrypted (done by the CLI layer).
func Upload(ctx context.Context, cfg *types.ServerConfig, localPath, remotePath string) (*types.FileTransferResult, error) {

	result := &types.FileTransferResult{
		ServerID: cfg.ID,
		Path:     remotePath,
	}

	start := time.Now()

	client, err := Connect(ctx, cfg)
	if err != nil {
		return result, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return result, fmt.Errorf("start sftp: %w", err)
	}
	defer sftpClient.Close()

	// Open local file.
	localFile, err := os.Open(localPath)
	if err != nil {
		return result, fmt.Errorf("open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	localInfo, err := localFile.Stat()
	if err != nil {
		return result, fmt.Errorf("stat local file %s: %w", localPath, err)
	}

	// Ensure remote directory exists.
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		if err := sftpClient.MkdirAll(remoteDir); err != nil {
			return result, fmt.Errorf("create remote dir %s: %w", remoteDir, err)
		}
	}

	// Create remote file.
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return result, fmt.Errorf("create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	written, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return result, fmt.Errorf("transfer: %w", err)
	}

	result.SizeBytes = written
	result.DurationMs = time.Since(start).Milliseconds()
	_ = localInfo // size already captured via io.Copy return value

	return result, nil
}

// Download copies a remote file to the local machine via SFTP.
// The cfg must have its password already decrypted (done by the CLI layer).
func Download(ctx context.Context, cfg *types.ServerConfig, remotePath, localPath string) (*types.FileTransferResult, error) {

	result := &types.FileTransferResult{
		ServerID: cfg.ID,
		Path:     remotePath,
	}

	start := time.Now()

	client, err := Connect(ctx, cfg)
	if err != nil {
		return result, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return result, fmt.Errorf("start sftp: %w", err)
	}
	defer sftpClient.Close()

	// Open remote file.
	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return result, fmt.Errorf("open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Ensure local directory exists.
	localDir := filepath.Dir(localPath)
	if localDir != "." {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return result, fmt.Errorf("create local dir %s: %w", localDir, err)
		}
	}

	// Create local file.
	localFile, err := os.Create(localPath)
	if err != nil {
		return result, fmt.Errorf("create local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	written, err := io.Copy(localFile, remoteFile)
	if err != nil {
		return result, fmt.Errorf("transfer: %w", err)
	}

	result.SizeBytes = written
	result.DurationMs = time.Since(start).Milliseconds()

	return result, nil
}
