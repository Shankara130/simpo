package services

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// InMemoryFileStorage is a simple in-memory implementation of FileStorageService
// Story 5.3, Task 6: Basic file storage for export files
type InMemoryFileStorage struct {
	files     map[string][]byte
	mutex     sync.RWMutex
	basePath  string
	maxSizeMB int64
}

// NewInMemoryFileStorage creates a new in-memory file storage service
func NewInMemoryFileStorage(basePath string, maxSizeMB int64) *InMemoryFileStorage {
	return &InMemoryFileStorage{
		files:     make(map[string][]byte),
		basePath:  basePath,
		maxSizeMB: maxSizeMB,
	}
}

// SaveFile saves file data to memory
func (fs *InMemoryFileStorage) SaveFile(ctx context.Context, filename string, data []byte) (string, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// Check disk space
	ok, err := fs.CheckDiskSpace(ctx, int64(len(data)))
	if !ok {
		return "", err
	}

	// Store in memory
	fs.files[filename] = data
	return filepath.Join(fs.basePath, filename), nil
}

// GetFile retrieves file data from memory
func (fs *InMemoryFileStorage) GetFile(ctx context.Context, filePath string) ([]byte, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filename := filepath.Base(filePath)
	data, ok := fs.files[filename]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

// DeleteFile removes a file from memory
func (fs *InMemoryFileStorage) DeleteFile(ctx context.Context, filePath string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filename := filepath.Base(filePath)
	delete(fs.files, filename)
	return nil
}

// CleanupOldFiles removes files older than the specified duration
func (fs *InMemoryFileStorage) CleanupOldFiles(ctx context.Context, olderThan time.Duration) error {
	// In-memory storage doesn't track creation times, so we just clear all
	// In production, this would check timestamps and selectively delete
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.files = make(map[string][]byte)
	return nil
}

// CheckDiskSpace checks if there's enough disk space for the required bytes
func (fs *InMemoryFileStorage) CheckDiskSpace(ctx context.Context, requiredBytes int64) (bool, error) {
	// Simple check: ensure we don't exceed max size
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	var totalSize int64
	for _, data := range fs.files {
		totalSize += int64(len(data))
	}

	maxBytes := fs.maxSizeMB * 1024 * 1024
	if totalSize+requiredBytes > maxBytes {
		return false, nil
	}
	return true, nil
}
