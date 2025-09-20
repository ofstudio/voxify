package files

import (
	"fmt"
	"os"
)

func IsDir(dirPath string) error {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %w", err)
	}
	if err != nil {
		return fmt.Errorf("error accessing directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}
	return nil
}
