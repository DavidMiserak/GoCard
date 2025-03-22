// Package io provides file system operations for the GoCard storage system.
package io

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}
	return nil
}

// GetAbsolutePath returns the absolute path for a given path
func GetAbsolutePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}
	return absPath, nil
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists checks if a directory exists at the given path
// Returns a boolean indicating existence and any error encountered
func DirectoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// ReadFileContent reads the content of a file at the given path
func ReadFileContent(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return content, nil
}

// WriteFileContent writes content to a file at the given path
func WriteFileContent(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := EnsureDirectoryExists(dir); err != nil {
		return err
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}
	return nil
}

// DeleteFile deletes a file at the given path
func DeleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	return nil
}

// MoveFile moves a file from sourcePath to targetPath
func MoveFile(sourcePath, targetPath string) error {
	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := EnsureDirectoryExists(targetDir); err != nil {
		return err
	}

	// Move the file
	if err := os.Rename(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", sourcePath, targetPath, err)
	}
	return nil
}

// RenameDirectory renames a directory
func RenameDirectory(oldPath, newPath string) error {
	// Ensure parent directory of new path exists
	newParentDir := filepath.Dir(newPath)
	if err := EnsureDirectoryExists(newParentDir); err != nil {
		return err
	}

	// Rename the directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename directory from %s to %s: %w", oldPath, newPath, err)
	}
	return nil
}

// DeleteDirectory removes a directory and all its contents
func DeleteDirectory(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete directory %s: %w", path, err)
	}
	return nil
}
