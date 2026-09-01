package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"

	"ssh-skill/internal/types"
)

// newSFTPClient 创建启用并发传输的 SFTP 客户端。
// conn 是本包 Client wrapper；内部取 conn.Client 传给 sftp.NewClient。
// maxConcurrent 是单文件并发读写请求数上限。
func newSFTPClient(conn *Client, maxConcurrent int) (*sftp.Client, error) {
	sftpClient, err := sftp.NewClient(conn.Client,
		sftp.UseConcurrentWrites(true),
		sftp.UseConcurrentReads(true),
		// MaxPacketUnchecked: pkg/sftp 默认单包 32KB (1<<15)；提到 64KB 以摊薄高延迟链路的 RTT。
		// 原始代码未设置此 option，使用默认 32KB。
		sftp.MaxPacketUnchecked(64*1024),
		sftp.MaxConcurrentRequestsPerFile(maxConcurrent),
	)
	if err != nil {
		return nil, fmt.Errorf("start sftp: %w", err)
	}
	return sftpClient, nil
}

// uploadFile 通过已建好的 sftpClient 把本地文件传到远端，返回实际写入字节数。
// 启用并发写时 pkg/sftp 可能留下文件空洞，必须在 io.Copy 完成后调用
// Truncate 把远程文件修正到实际写入大小。
func uploadFile(sftpClient *sftp.Client, localPath, remotePath string, onProgress types.ProgressCallback) (int64, error) {
	start := time.Now()

	// Open local file.
	localFile, err := os.Open(localPath)
	if err != nil {
		return 0, fmt.Errorf("open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	localInfo, err := localFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat local file %s: %w", localPath, err)
	}

	// Ensure remote directory exists.
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		if err := sftpClient.MkdirAll(remoteDir); err != nil {
			return 0, fmt.Errorf("create remote dir %s: %w", remoteDir, err)
		}
	}

	// Create remote file.
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return 0, fmt.Errorf("create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// Wrap local file with progress tracking and use buffered copy.
	src := newProgressReader(localFile, localInfo.Size(), onProgress)
	buf := make([]byte, transferBufferSize)
	written, err := io.CopyBuffer(remoteFile, src, buf)
	if err != nil {
		return written, fmt.Errorf("transfer: %w", err)
	}

	// Concurrent writes may leave the remote file with a trailing hole;
	// truncate to the actual number of bytes written.
	if err := remoteFile.Truncate(written); err != nil {
		return written, fmt.Errorf("truncate remote file %s: %w", remotePath, err)
	}

	// Final progress tick at 100%.
	if onProgress != nil {
		onProgress(written, localInfo.Size(), time.Since(start))
	}

	return written, nil
}

// downloadFile 通过已建好的 sftpClient 把远程文件拉到本地，返回实际写入字节数。
func downloadFile(sftpClient *sftp.Client, remotePath, localPath string, onProgress types.ProgressCallback) (int64, error) {
	start := time.Now()

	// Open remote file.
	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return 0, fmt.Errorf("open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	remoteInfo, err := remoteFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat remote file %s: %w", remotePath, err)
	}

	// Ensure local directory exists.
	localDir := filepath.Dir(localPath)
	if localDir != "." {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return 0, fmt.Errorf("create local dir %s: %w", localDir, err)
		}
	}

	// Create local file.
	localFile, err := os.Create(localPath)
	if err != nil {
		return 0, fmt.Errorf("create local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Wrap local file with progress tracking; remoteFile.WriteTo drives the copy.
	dst := newProgressWriter(localFile, remoteInfo.Size(), onProgress)
	// 原始: io.CopyBuffer(localFile, src, buf) — dst=os.File 实现 ReaderFrom，
	// os.File.ReadFrom 不认识 progressReader → 回退 32KB 串行读，仅 ~2.5MB/s。
	// 改为走 remoteFile.WriteTo(progressWriter) 以触发 pkg/sftp 并发读。
	written, err := remoteFile.WriteTo(dst)
	if err != nil {
		return written, fmt.Errorf("transfer: %w", err)
	}

	// Final progress tick at 100%.
	if onProgress != nil {
		onProgress(written, remoteInfo.Size(), time.Since(start))
	}

	return written, nil
}
