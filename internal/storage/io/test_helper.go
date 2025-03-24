// File: internal/storage/io/test_helper.go

package io

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHelper provides utilities for testing file operations
type TestHelper struct {
	t          *testing.T
	mockFS     *MockFileSystem
	originalFS FileSystem
}

// NewTestHelper creates a new TestHelper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:      t,
		mockFS: NewMockFileSystem(),
	}
}

// SetupDir creates a directory in the mock filesystem
func (h *TestHelper) SetupDir(path string) {
	h.mockFS.SetupDir(path)
}

// SetupFile creates a file with the given content in the mock filesystem
func (h *TestHelper) SetupFile(path, content string) {
	h.mockFS.SetupFile(path, []byte(content))
}

// InjectError sets up an error for a specific operation and path
func (h *TestHelper) InjectError(op string, path string, err error) {
	h.mockFS.InjectError(op, path, err)
}

// UseWithinTest temporarily replaces the default filesystem with the mock filesystem
// for the duration of the test function
func (h *TestHelper) UseWithinTest(testFunc func()) {
	// Save the original filesystem and replace with mock
	h.originalFS = SetDefaultFS(h.mockFS)

	// Run the test function
	testFunc()

	// Restore the original filesystem
	SetDefaultFS(h.originalFS)
}

// AssertFileExists asserts that a file exists in the mock filesystem
func (h *TestHelper) AssertFileExists(path string) {
	_, err := h.mockFS.Stat(path)
	if os.IsNotExist(err) {
		h.t.Errorf("Expected file to exist at %s, but it doesn't", path)
	} else if err != nil {
		h.t.Errorf("Error checking file existence at %s: %v", path, err)
	}
}

// AssertFileDoesNotExist asserts that a file doesn't exist in the mock filesystem
func (h *TestHelper) AssertFileDoesNotExist(path string) {
	_, err := h.mockFS.Stat(path)
	if err == nil {
		h.t.Errorf("Expected file to not exist at %s, but it does", path)
	} else if !os.IsNotExist(err) {
		h.t.Errorf("Error checking file non-existence at %s: %v", path, err)
	}
}

// AssertFileContent asserts that a file has the expected content
func (h *TestHelper) AssertFileContent(path, expected string) {
	content, err := h.mockFS.ReadFile(path)
	if err != nil {
		h.t.Errorf("Failed to read file at %s: %v", path, err)
		return
	}

	if string(content) != expected {
		h.t.Errorf("File content mismatch.\nExpected: %s\nActual: %s", expected, string(content))
	}
}

// AssertDirExists asserts that a directory exists in the mock filesystem
func (h *TestHelper) AssertDirExists(path string) {
	info, err := h.mockFS.Stat(path)
	if os.IsNotExist(err) {
		h.t.Errorf("Expected directory to exist at %s, but it doesn't", path)
	} else if err != nil {
		h.t.Errorf("Error checking directory existence at %s: %v", path, err)
	} else if !info.IsDir() {
		h.t.Errorf("Path %s exists but is not a directory", path)
	}
}

// CreateTempDir creates a temp directory structure in the mock filesystem
func (h *TestHelper) CreateTempDir(subDirs ...string) map[string]string {
	paths := make(map[string]string)

	// Create base temp directory
	baseDir := filepath.Join("/", "temp")
	h.SetupDir(baseDir)
	paths["base"] = baseDir

	// Create subdirectories
	for _, subDir := range subDirs {
		path := filepath.Join(baseDir, subDir)
		h.SetupDir(path)
		paths[subDir] = path
	}

	return paths
}

// GetFS returns the mock filesystem
func (h *TestHelper) GetFS() *MockFileSystem {
	return h.mockFS
}
