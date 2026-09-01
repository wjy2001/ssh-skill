package ssh

import (
	"context"
	"fmt"
	"io"
	"time"

	"ssh-skill/internal/types"
)

// transferBufferSize is the buffer size used for file transfers (256KB).
// Larger than io.Copy default (32KB) for better throughput on high-latency links.
const transferBufferSize = 256 * 1024

// defaultTransferRetries is the maximum number of whole-transfer attempts.
// A failed attempt (connect or transfer error) is retried with a fresh SSH
// connection; the attempt count is the total, so 3 means at most 3 tries.
const defaultTransferRetries = 3

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

// Size 暴露文件总大小，使 pkg/sftp File.ReadFrom 能探测 remain 从而启用并发写。
func (pr *progressReader) Size() int64 { return pr.total }

// progressWriter wraps an io.Writer and calls onProgress with transfer stats.
// Used on the download path, where pkg/sftp File.WriteTo drives the copy and
// issues concurrent reads on the remote side (the latency-bound direction).
type progressWriter struct {
	writer   io.Writer
	total    int64
	written  int64
	start    time.Time
	lastTime time.Time
	onTick   types.ProgressCallback
}

func newProgressWriter(w io.Writer, total int64, onTick types.ProgressCallback) *progressWriter {
	now := time.Now()
	return &progressWriter{
		writer:   w,
		total:    total,
		start:    now,
		lastTime: now,
		onTick:   onTick,
	}
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.written += int64(n)

	// Same 100ms throttle as progressReader: keep the terminal responsive
	// without flooding it.
	if pw.onTick != nil && time.Since(pw.lastTime) >= 100*time.Millisecond {
		pw.onTick(pw.written, pw.total, time.Since(pw.start))
		pw.lastTime = time.Now()
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
	var lastErr error
	for attempt := 1; attempt <= defaultTransferRetries; attempt++ {
		if attempt > 1 {
			// From the second attempt on, back off briefly before retrying
			// the whole transfer from scratch.
			select {
			case <-time.After(2 * time.Second * time.Duration(attempt-1)):
			case <-ctx.Done():
				return result, ctx.Err()
			}
		}

		// A fresh SSH connection per attempt: failures leave the previous
		// connection closed, so no handles leak across attempts.
		client, err := Connect(ctx, cfg)
		if err != nil {
			lastErr = err
			continue
		}
		sftpClient, err := newSFTPClient(client, 16) // max concurrent requests per file
		if err != nil {
			client.Close()
			lastErr = err
			continue
		}
		written, err := uploadFile(sftpClient, localPath, remotePath, onProgress)
		sftpClient.Close()
		client.Close()
		if err != nil {
			lastErr = err
			// Whole-transfer retry: reconnect and upload from 0
			// (sftpClient.Create overwrites the half-transferred file).
			continue
		}

		result.SizeBytes = written
		result.DurationMs = time.Since(start).Milliseconds()
		if onProgress != nil {
			onProgress(written, written, time.Since(start)) // 100% final tick
		}
		return result, nil
	}
	return result, fmt.Errorf("upload attempt %d/%d: %w", defaultTransferRetries, defaultTransferRetries, lastErr)
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
	var lastErr error
	for attempt := 1; attempt <= defaultTransferRetries; attempt++ {
		if attempt > 1 {
			// From the second attempt on, back off briefly before retrying
			// the whole transfer from scratch.
			select {
			case <-time.After(2 * time.Second * time.Duration(attempt-1)):
			case <-ctx.Done():
				return result, ctx.Err()
			}
		}

		// A fresh SSH connection per attempt: failures leave the previous
		// connection closed, so no handles leak across attempts.
		client, err := Connect(ctx, cfg)
		if err != nil {
			lastErr = err
			continue
		}
		sftpClient, err := newSFTPClient(client, 16) // max concurrent requests per file
		if err != nil {
			client.Close()
			lastErr = err
			continue
		}
		written, err := downloadFile(sftpClient, remotePath, localPath, onProgress)
		sftpClient.Close()
		client.Close()
		if err != nil {
			lastErr = err
			// Whole-transfer retry: reconnect and download into a freshly
			// created local file (os.Create truncates the half-download).
			continue
		}

		result.SizeBytes = written
		result.DurationMs = time.Since(start).Milliseconds()
		if onProgress != nil {
			onProgress(written, written, time.Since(start)) // 100% final tick
		}
		return result, nil
	}
	return result, fmt.Errorf("download attempt %d/%d: %w", defaultTransferRetries, defaultTransferRetries, lastErr)
}
