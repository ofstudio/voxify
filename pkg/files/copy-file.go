package files

import (
	"io"
	"os"
)

// CopyFile copies the contents of src to dst
func CopyFile(src, dst string) error {
	// Open source file for reading
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer destFile.Close()

	// Copy contents from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Flush writes to disk
	err = destFile.Sync()
	if err != nil {
		return err
	}

	// Preserve file permissions from original file
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
