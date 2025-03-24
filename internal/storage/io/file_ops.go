// File: internal/storage/io/file_ops.go

package io

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(dirPath string) error {
	return EnsureDirectoryExistsWithFS(GetDefaultFS(), dirPath)
}

// EnsureDirectoryExistsWithFS creates a directory if it doesn't exist using the provided filesystem
func EnsureDirectoryExistsWithFS(fs FileSystem, dirPath string) error {
	_, err := fs.Stat(dirPath)
	if os.IsNotExist(err) {
		if err := fs.MkdirAll(dirPath, 0755); err != nil {
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
	return FileExistsWithFS(GetDefaultFS(), path)
}

// FileExistsWithFS checks if a file exists at the given path using the provided filesystem
func FileExistsWithFS(fs FileSystem, path string) bool {
	info, err := fs.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists checks if a directory exists at the given path
func DirectoryExists(path string) (bool, error) {
	return DirectoryExistsWithFS(GetDefaultFS(), path)
}

// DirectoryExistsWithFS checks if a directory exists using the provided filesystem
func DirectoryExistsWithFS(fs FileSystem, path string) (bool, error) {
	info, err := fs.Stat(path)
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
	return ReadFileContentWithFS(GetDefaultFS(), path)
}

// ReadFileContentWithFS reads the content of a file using the provided filesystem
func ReadFileContentWithFS(fs FileSystem, path string) ([]byte, error) {
	content, err := fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return content, nil
}

// WriteFileContent writes content to a file at the given path
func WriteFileContent(path string, content []byte) error {
	return WriteFileContentWithFS(GetDefaultFS(), path, content)
}

// WriteFileContentWithFS writes content to a file using the provided filesystem
func WriteFileContentWithFS(fs FileSystem, path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := EnsureDirectoryExistsWithFS(fs, dir); err != nil {
		return err
	}

	if err := fs.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}
	return nil
}

// DeleteFile deletes a file at the given path
func DeleteFile(path string) error {
	return DeleteFileWithFS(GetDefaultFS(), path)
}

// DeleteFileWithFS deletes a file using the provided filesystem
func DeleteFileWithFS(fs FileSystem, path string) error {
	if err := fs.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	return nil
}

// MoveFile moves a file from sourcePath to targetPath
func MoveFile(sourcePath, targetPath string) error {
	return MoveFileWithFS(GetDefaultFS(), sourcePath, targetPath)
}

// MoveFileWithFS moves a file using the provided filesystem
func MoveFileWithFS(fs FileSystem, sourcePath, targetPath string) error {
	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := EnsureDirectoryExistsWithFS(fs, targetDir); err != nil {
		return err
	}

	// Move the file
	if err := fs.Rename(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", sourcePath, targetPath, err)
	}
	return nil
}

// RenameDirectory renames a directory
func RenameDirectory(oldPath, newPath string) error {
	return RenameDirectoryWithFS(GetDefaultFS(), oldPath, newPath)
}

// RenameDirectoryWithFS renames a directory using the provided filesystem
func RenameDirectoryWithFS(fs FileSystem, oldPath, newPath string) error {
	// Ensure parent directory of new path exists
	newParentDir := filepath.Dir(newPath)
	if err := EnsureDirectoryExistsWithFS(fs, newParentDir); err != nil {
		return err
	}

	// Rename the directory
	if err := fs.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename directory from %s to %s: %w", oldPath, newPath, err)
	}
	return nil
}

// DeleteDirectory removes a directory and all its contents
func DeleteDirectory(path string) error {
	return DeleteDirectoryWithFS(GetDefaultFS(), path)
}

// DeleteDirectoryWithFS removes a directory and all its contents using the provided filesystem
func DeleteDirectoryWithFS(fs FileSystem, path string) error {
	if err := fs.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete directory %s: %w", path, err)
	}
	return nil
}
