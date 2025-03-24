// File: internal/testing/helpers.go

package testing

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates and manages a temporary directory for testing
func TempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "gocard-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Register cleanup to run after the test completes
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// WriteTestFile creates a test file with the given content
func WriteTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()

	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	return path
}

// CreateTestSubdirs creates a test directory structure
func CreateTestSubdirs(t *testing.T, baseDir string, subDirs ...string) map[string]string {
	t.Helper()

	paths := make(map[string]string)
	for _, sub := range subDirs {
		path := filepath.Join(baseDir, sub)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			t.Fatalf("Failed to create subdirectory %s: %v", sub, err)
		}
		paths[sub] = path
	}

	return paths
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s, but it doesn't", path)
	} else if err != nil {
		t.Errorf("Error checking file existence at %s: %v", path, err)
	}
}

// AssertFileDoesNotExist checks that a file doesn't exist
func AssertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err == nil {
		t.Errorf("Expected file to not exist at %s, but it does", path)
	} else if !os.IsNotExist(err) {
		t.Errorf("Error checking file non-existence at %s: %v", path, err)
	}
}

// AssertFileContent checks that a file has the expected content
func AssertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file at %s: %v", path, err)
		return
	}

	if string(content) != expected {
		t.Errorf("File content mismatch.\nExpected: %s\nActual: %s", expected, string(content))
	}
}
