// internal/storage/io/mock_filesystem_test.go
package io

import (
	"errors"
	"os"
	"testing"
)

func TestMockFileSystem(t *testing.T) {
	fs := NewMockFileSystem()

	// Test writing and reading a file
	testData := []byte("test data")
	err := fs.WriteFile("/test/file.txt", testData, 0644)
	if err != nil {
		t.Errorf("WriteFile error: %v", err)
	}

	readData, err := fs.ReadFile("/test/file.txt")
	if err != nil {
		t.Errorf("ReadFile error: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("Expected %q, got %q", testData, readData)
	}

	// Test parent directory was created automatically
	info, err := fs.Stat("/test")
	if err != nil {
		t.Errorf("Stat error on directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected IsDir() to be true for directory")
	}

	// Test explicitly creating directories
	err = fs.MkdirAll("/test/dir1/dir2", 0755)
	if err != nil {
		t.Errorf("MkdirAll error: %v", err)
	}

	// Test Stat on directory
	info, err = fs.Stat("/test/dir1")
	if err != nil {
		t.Errorf("Stat error on directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected IsDir() to be true for directory")
	}

	// Test Stat on file
	info, err = fs.Stat("/test/file.txt")
	if err != nil {
		t.Errorf("Stat error on file: %v", err)
	}

	if info.IsDir() {
		t.Error("Expected IsDir() to be false for file")
	}

	if info.Size() != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), info.Size())
	}

	// Test Remove
	err = fs.Remove("/test/file.txt")
	if err != nil {
		t.Errorf("Remove error: %v", err)
	}

	_, err = fs.Stat("/test/file.txt")
	if !os.IsNotExist(err) {
		t.Errorf("Expected ErrNotExist after remove, got %v", err)
	}

	// Test Rename
	if err := fs.WriteFile("/test/old.txt", []byte("old data"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Add the missing rename operation
	if err := fs.Rename("/test/old.txt", "/test/dir1/new.txt"); err != nil {
		t.Errorf("Rename error: %v", err)
	}

	_, err = fs.Stat("/test/old.txt")
	if !os.IsNotExist(err) {
		t.Errorf("Expected ErrNotExist for old file after rename, got %v", err)
	}

	_, err = fs.Stat("/test/dir1/new.txt")
	if err != nil {
		t.Errorf("Stat error on renamed file: %v", err)
	}

	// Test RemoveAll
	if err := fs.WriteFile("/test/dir1/file1.txt", []byte("file1"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := fs.WriteFile("/test/dir1/dir2/file2.txt", []byte("file2"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = fs.RemoveAll("/test/dir1")
	if err != nil {
		t.Errorf("RemoveAll error: %v", err)
	}

	_, err = fs.Stat("/test/dir1")
	if !os.IsNotExist(err) {
		t.Errorf("Expected ErrNotExist for directory after RemoveAll, got %v", err)
	}

	// Test error injection
	expectedErr := errors.New("injected error")
	fs.SetupFile("/error/file.txt", []byte("data"))
	fs.InjectError("read", "/error/file.txt", expectedErr)

	_, err = fs.ReadFile("/error/file.txt")
	if err != expectedErr {
		t.Errorf("Expected injected error, got %v", err)
	}
}
