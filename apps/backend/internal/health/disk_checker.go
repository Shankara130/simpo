package health

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type diskChecker struct {
	path string
}

// NewDiskChecker creates a new disk health checker for the specified path
// Story 6.2, Task 1.3: Add disk usage checker with threshold monitoring
func NewDiskChecker(path string) Checker {
	// Use root path as default if empty
	if path == "" {
		path = "/"
	}
	return &diskChecker{path: path}
}

func (d *diskChecker) Name() string {
	return "disk"
}

func (d *diskChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	// Get filesystem statistics
	var stat syscall.Statfs_t
	err := syscall.Statfs(d.path, &stat)
	if err != nil {
		slog.Error("Failed to get disk stats", "path", d.path, "error", err)
		return CheckResult{
			Status:  CheckFail,
			Message: fmt.Sprintf("Failed to get disk stats for %s: %v", d.path, err),
		}
	}

	// Calculate disk metrics
	// Bavail is available blocks to non-privileged user
	// Blocks is total file system blocks
	// Bsize is block size in bytes
	totalBytes := stat.Blocks * uint64(stat.Bsize)

	// PATCH: Check for integer overflow (Story 6.3 code review)
	if stat.Blocks > 0 && totalBytes/stat.Blocks != uint64(stat.Bsize) {
		slog.Error("Integer overflow in disk size calculation", "path", d.path, "blocks", stat.Blocks, "bsize", stat.Bsize)
		return CheckResult{
			Status:  CheckFail,
			Message: fmt.Sprintf("Disk size calculation overflow for %s", d.path),
		}
	}

	// PATCH: Check for division by zero (Story 6.3 code review)
	if totalBytes == 0 {
		slog.Error("Invalid filesystem stats: total bytes is zero", "path", d.path)
		return CheckResult{
			Status:  CheckFail,
			Message: fmt.Sprintf("Invalid filesystem stats for %s: total bytes is zero", d.path),
		}
	}

	freeBytes := stat.Bavail * uint64(stat.Bsize)
	usedBytes := totalBytes - freeBytes

	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)
	usedGB := float64(usedBytes) / (1024 * 1024 * 1024)
	freeGB := float64(freeBytes) / (1024 * 1024 * 1024)
	freePercentage := (float64(freeBytes) / float64(totalBytes)) * 100

	// Determine status based on thresholds
	// Story 6.2: Critical if < 10% free, Warning if < 20% free
	status := CheckPass
	message := "Disk space healthy"

	if freePercentage < 10.0 {
		status = CheckFail
		message = fmt.Sprintf("Critical: Only %.2f%% disk space remaining (%.2f GB free of %.2f GB total)",
			freePercentage, freeGB, totalGB)
	} else if freePercentage < 20.0 {
		status = CheckWarn
		message = fmt.Sprintf("Warning: Disk space below 20%% (%.2f%% free, %.2f GB free of %.2f GB total)",
			freePercentage, freeGB, totalGB)
	}

	duration := time.Since(start)

	return CheckResult{
		Status:  status,
		Message: message,
		ResponseTime: duration.String(),
		Details: map[string]interface{}{
			"used_gb":         usedGB,
			"total_gb":        totalGB,
			"free_gb":         freeGB,
			"free_percentage": freePercentage,
		},
	}
}

// GetAbsolutePath returns the absolute path for the disk checker
// Useful for testing and debugging
func (d *diskChecker) GetAbsolutePath() (string, error) {
	return filepath.Abs(d.path)
}

// Exists checks if the path exists
// Useful for validation before checking disk stats
func (d *diskChecker) Exists() bool {
	_, err := os.Stat(d.path)
	return err == nil
}
