package cli

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const progressBarWidth = 30

// renderProgress writes an inline progress bar to stderr using carriage return.
// Format: "Uploading: [====>     ] 45%  2.3MB/5.1MB  1.2MB/s  00:03"
func renderProgress(label string, transferred, total int64, elapsed time.Duration) {
	if total <= 0 {
		fmt.Fprintf(os.Stderr, "\r%s: %s ... %s   ",
			label, formatSize(transferred), formatDuration(elapsed))
		return
	}

	pct := float64(transferred) / float64(total)
	if pct > 1.0 {
		pct = 1.0
	}

	filled := int(pct * progressBarWidth)
	if filled > progressBarWidth {
		filled = progressBarWidth
	}
	bar := ""
	if filled > 0 {
		bar = strings.Repeat("=", filled-1) + ">"
	}
	empty := strings.Repeat("-", progressBarWidth-filled)

	speed := float64(0)
	if elapsed.Seconds() > 0 {
		speed = float64(transferred) / elapsed.Seconds()
	}
	eta := "--:--"
	if speed > 0 && total > transferred {
		remaining := float64(total-transferred) / speed
		eta = formatDuration(time.Duration(remaining) * time.Second)
	}

	fmt.Fprintf(os.Stderr, "\r%s: [%s%s] %3.0f%%  %s/%s  %s/s  %s   ",
		label, bar, empty, pct*100,
		formatSize(transferred), formatSize(total),
		formatSize(int64(speed)), eta)
}

// finishProgress clears the progress line and prints a completion summary to stderr.
func finishProgress(label string, total int64, elapsed time.Duration) {
	speed := float64(0)
	if elapsed.Seconds() > 0 {
		speed = float64(total) / elapsed.Seconds()
	}
	fmt.Fprintf(os.Stderr, "\r%-80s\r", "") // Clear the progress line.
	fmt.Fprintf(os.Stderr, "%s: done — %s in %s (%.1f MB/s)\n",
		label, formatSize(total), formatDuration(elapsed), speed/1024/1024)
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
