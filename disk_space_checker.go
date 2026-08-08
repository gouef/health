package health

import (
	"context"
	"fmt"
	"syscall"
	"time"
)

var statfs = func(path string, stat *syscall.Statfs_t) error {
	return syscall.Statfs(path, stat)
}

type DiskChecker struct {
	name         string
	path         string
	minFreeBytes uint64
}

func NewDiskChecker(name string, path string, minFreeBytes uint64) *DiskChecker {
	if name == "" {
		name = "disk_space"
	}
	if path == "" {
		path = "/"
	}
	return &DiskChecker{
		name:         name,
		path:         path,
		minFreeBytes: minFreeBytes,
	}
}

func (c *DiskChecker) Name() string {
	return c.name
}

func (c *DiskChecker) Check(ctx context.Context) Result {
	start := time.Now()

	var stat syscall.Statfs_t
	err := statfs(c.path, &stat)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return Result{
			Status:       StatusDown,
			Type:         "storage",
			ResponseTime: duration,
			Error:        fmt.Sprintf("failed to get disk stats for %s: %v", c.path, err),
		}
	}

	freeBytes := stat.Bavail * uint64(stat.Bsize)
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	usedBytes := totalBytes - (stat.Bfree * uint64(stat.Bsize))

	var freePercentage float64
	if totalBytes > 0 {
		freePercentage = (float64(freeBytes) / float64(totalBytes)) * 100
	}

	details := map[string]interface{}{
		"path":            c.path,
		"free_bytes":      freeBytes,
		"free_human":      humanizeBytes(freeBytes),
		"total_bytes":     totalBytes,
		"total_human":     humanizeBytes(totalBytes),
		"used_bytes":      usedBytes,
		"free_percentage": fmt.Sprintf("%.2f%%", freePercentage),
	}

	if freeBytes < c.minFreeBytes {
		return Result{
			Status:       StatusDown,
			Type:         "storage",
			ResponseTime: duration,
			Error:        fmt.Sprintf("low disk space on %s: %s free (minimum required: %s)", c.path, humanizeBytes(freeBytes), humanizeBytes(c.minFreeBytes)),
			Details:      details,
		}
	}

	return Result{
		Status:       StatusUp,
		Type:         "storage",
		ResponseTime: duration,
		Details:      details,
	}
}

func humanizeBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
