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

// transferBufferSize is the buffer size used for file transfers (256KB).
// Larger than io.Copy default (32KB) for better throughput on high-latency links.
const transferBufferSize = 256 * 1024

// progressReader wraps an io.Reader and calls onProgress with transfer stats.
type progressReader struct {
	reader   io.Reader
	total    int64
	read     int64
	start    time.Time
	lastTime time.Time
	onTick   types.ProgressCallback
}

func newProgressReader(r io.Reader, total int64, onTick types.ProgressCallback) *progressReader {
	now := time.Now()
	return &progressReader{
		reader:   r,
		total:    total,
		start:    now,
		lastTime: now,
		onTick:   onTick,
	}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.read += int64(n)

	// Fire callback up to 10 times per second to keep terminal responsive
	// without flooding it.
	if pr.onTick != nil && time.Since(pr.lastTime) >= 100*time.Millisecond {
		pr.onTick(pr.read, pr.total, time.Since(pr.start))
		pr.lastTime = time.Now()
	}
	return n, err
}

// Upload copies a local file to the remote server via SFTP.
// The cfg must have its password already decrypted (done by the CLI layer).
// If onProgress is non-nil, it is called periodically with transfer progress.
func Upload(ctx context.Context, cfg *types.ServerConfig, localPath, remotePath string, onProgress types.ProgressCallback) (*types.FileTransferResult, error) {

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

	sftpClient, err := sftp.NewClient(client.Client)
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

	// Wrap local file with progress tracking and use buffered copy.
	src := newProgressReader(localFile, localInfo.Size(), onProgress)
	buf := make([]byte, transferBufferSize)
	written, err := io.CopyBuffer(remoteFile, src, buf)
	if err != nil {
		return result, fmt.Errorf("transfer: %w", err)
	}

	// Final progress tick at 100%.
	if onProgress != nil {
		onProgress(written, localInfo.Size(), time.Since(start))
	}

	result.SizeBytes = written
	result.DurationMs = time.Since(start).Milliseconds()

	return result, nil
}

// Download copies a remote file to the local machine via SFTP.
// The cfg must have its password already decrypted (done by the CLI layer).
// If onProgress is non-nil, it is called periodically with transfer progress.
func Download(ctx context.Context, cfg *types.ServerConfig, remotePath, localPath string, onProgress types.ProgressCallback) (*types.FileTransferResult, error) {

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

	sftpClient, err := sftp.NewClient(client.Client)
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

	remoteInfo, err := remoteFile.Stat()
	if err != nil {
		return result, fmt.Errorf("stat remote file %s: %w", remotePath, err)
	}

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

	// Wrap remote file with progress tracking and use buffered copy.
	src := newProgressReader(remoteFile, remoteInfo.Size(), onProgress)
	buf := make([]byte, transferBufferSize)
	written, err := io.CopyBuffer(localFile, src, buf)
	if err != nil {
		return result, fmt.Errorf("transfer: %w", err)
	}

	// Final progress tick at 100%.
	if onProgress != nil {
		onProgress(written, remoteInfo.Size(), time.Since(start))
	}

	result.SizeBytes = written
	result.DurationMs = time.Since(start).Milliseconds()

	return result, nil
}
