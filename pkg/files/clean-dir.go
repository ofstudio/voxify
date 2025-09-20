package files

import (
	"fmt"
	"os"
	"path/filepath"
)

// CleanDir removes all files and subdirectories in the specified directory.
func CleanDir(dirPath string) error {
	// Init if the directory exists
	if err := IsDir(dirPath); err != nil {
		return err
	}

	// Read all entries in the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	// Iterate and remove each entry
	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			// Remove directory and its contents
			err = os.RemoveAll(fullPath)
		} else {
			// Remove file
			err = os.Remove(fullPath)
		}
		if err != nil {
			return fmt.Errorf("failed to remove '%s': %w", fullPath, err)
		}
	}

	return nil
}
