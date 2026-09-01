package ssh

import (
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// DefaultKeepAliveInterval is the keepalive heartbeat interval.
const DefaultKeepAliveInterval = 15 * time.Second

// StartKeepAlive starts a background goroutine that periodically sends
// keepalive@openssh.com requests to keep the SSH connection alive (prevents
// NAT/firewall idle timeout from silently dropping long-running tasks).
// It returns a stop function for terminating the goroutine when the connection
// is closed; the stop function is idempotent. If the connection is already
// dead, SendRequest returns an error and the goroutine exits on its own
// (reconnecting is the caller's responsibility).
func StartKeepAlive(client *ssh.Client, interval time.Duration) func() {
	if client == nil {
		return func() {}
	}
	if interval <= 0 {
		interval = DefaultKeepAliveInterval
	}

	stop := make(chan struct{})
	var once sync.Once

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// SendRequest fails once the connection is dead or closed
				// locally; exit silently without logging or panicking.
				if _, _, err := client.SendRequest("keepalive@openssh.com", true, nil); err != nil {
					return
				}
			case <-stop:
				return
			}
		}
	}()

	return func() {
		once.Do(func() { close(stop) })
	}
}
