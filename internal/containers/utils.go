package containers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

type Container interface {
	Terminate(ctx context.Context)
}

// GetProjectRoot returns the absolute path to the project root.
// It works by finding the directory of the caller of this func and navigating up
// until it finds the go.mod file.
func GetProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("failed to get the caller information")
	}

	// Start directory for traversal
	dir := filepath.Dir(filename)

	for {
		modFile := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modFile); err == nil {
			return dir, nil // Found the project root!
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", errors.New("reached root without finding go.mod")
		}

		dir = parentDir
	}
}
