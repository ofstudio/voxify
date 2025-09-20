package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// MoveFile moves a file from sourcePath to destPath, handling cross-filesystem moves
func MoveFile(sourcePath, destPath string) error {
	// First attempt: try using os.Rename (works within same filesystem)
	err := os.Rename(sourcePath, destPath)
	if err == nil {
		return nil // Successfully moved with rename
	}

	// Init if the error is due to cross-filesystem operation
	var linkError *os.LinkError
	if errors.As(err, &linkError) {
		if errors.Is(linkError.Err, syscall.EXDEV) {
			// Cross-filesystem move required - copy then delete
			return copyAndRemove(sourcePath, destPath)
		}
	}

	// Return other errors (permissions, non-existent file, etc.)
	return err
}

// copyAndRemove handles cross-filesystem moves by copying then deleting the original
func copyAndRemove(sourcePath, destPath string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Copy the file contents
	if err := CopyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Remove the original file
	if err := os.Remove(sourcePath); err != nil {
		// Attempt to clean up the copied file if removal fails
		if err = os.Remove(destPath); err != nil {
			return err
		}
		return fmt.Errorf("failed to remove original file after copy: %v", err)
	}

	return nil
}
