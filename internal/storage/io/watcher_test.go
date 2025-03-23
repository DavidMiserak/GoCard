// internal/storage/io/watcher_test.go
package io

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	testhelp "github.com/DavidMiserak/GoCard/internal/testing"
)

// mockFsnotifyWatcher is a mock implementation of fsnotify.Watcher
// nolint:unused
type mockFsnotifyWatcher struct {
	events       chan FileEvent
	errors       chan error
	addedPaths   []string
	removedPaths []string
	closed       bool
}

// nolint:unused
func newMockFsnotifyWatcher() *mockFsnotifyWatcher {
	return &mockFsnotifyWatcher{
		events:       make(chan FileEvent, 10),
		errors:       make(chan error, 10),
		addedPaths:   []string{},
		removedPaths: []string{},
		closed:       false,
	}
}

// nolint:unused
func (m *mockFsnotifyWatcher) Add(path string) error {
	if m.closed {
		return errors.New("watcher is closed")
	}
	m.addedPaths = append(m.addedPaths, path)
	return nil
}

// nolint:unused
func (m *mockFsnotifyWatcher) Remove(path string) error {
	if m.closed {
		return errors.New("watcher is closed")
	}
	m.removedPaths = append(m.removedPaths, path)
	return nil
}

// nolint:unused
func (m *mockFsnotifyWatcher) Close() error {
	if m.closed {
		return errors.New("watcher already closed")
	}
	m.closed = true
	close(m.events)
	close(m.errors)
	return nil
}

// Mock creation function to replace the real fsnotify.NewWatcher
// nolint:unused
func mockNewFsnotifyWatcher() (*mockFsnotifyWatcher, error) {
	return newMockFsnotifyWatcher(), nil
}

// TestNewFileWatcher tests the creation of a new file watcher
func TestNewFileWatcher(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testhelp.TempDir(t)

	// Test with valid directory
	watcher, err := NewFileWatcher(tempDir)
	if err != nil {
		t.Errorf("NewFileWatcher error with valid directory: %v", err)
	}

	if watcher == nil {
		t.Fatal("Expected watcher to be created, got nil")
	}

	// Clean up
	if err := watcher.Stop(); err != nil {
		t.Errorf("Failed to stop watcher: %v", err)
	}

	// Test with non-existent directory
	nonExistentDir := filepath.Join(tempDir, "non-existent")
	watcher, err = NewFileWatcher(nonExistentDir)
	if err == nil {
		t.Error("Expected error with non-existent directory, got nil")
		if watcher != nil {
			if err := watcher.Stop(); err != nil {
				t.Errorf("Failed to stop watcher: %v", err)
			}
		}
	}
}

func TestSetLogger(t *testing.T) {
	tempDir := testhelp.TempDir(t)
	watcher, err := NewFileWatcher(tempDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Failed to stop watcher: %v", err)
		}
	}()

	var buf bytes.Buffer
	logger := NewLogger(&buf, DEBUG)

	// Set the logger
	watcher.SetLogger(logger)

	// Verify the logger was set (indirectly)
	if watcher.logger == nil {
		t.Error("Logger was not set")
	}
}

func TestStartStop(t *testing.T) {
	tempDir := testhelp.TempDir(t)
	watcher, err := NewFileWatcher(tempDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	// Test starting
	err = watcher.Start()
	if err != nil {
		t.Errorf("Failed to start watcher: %v", err)
	}

	if !watcher.isRunning {
		t.Error("Watcher should be running after Start()")
	}

	// Test starting again (should error)
	err = watcher.Start()
	if err == nil {
		t.Error("Expected error when starting twice")
	}

	// Test stopping
	err = watcher.Stop()
	if err != nil {
		t.Errorf("Failed to stop watcher: %v", err)
	}

	if watcher.isRunning {
		t.Error("Watcher should not be running after Stop()")
	}
}

func TestEventsAndErrors(t *testing.T) {
	tempDir := testhelp.TempDir(t)
	watcher, err := NewFileWatcher(tempDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Failed to stop watcher: %v", err)
		}
	}()

	// Test Events() method
	events := watcher.Events()
	if events == nil {
		t.Error("Events() returned nil channel")
	}

	// Test Errors() method
	errors := watcher.Errors()
	if errors == nil {
		t.Error("Errors() returned nil channel")
	}
}

// TestAddDirectory tests adding directories to the watcher
func TestAddDirectory(t *testing.T) {
	// This is a more complex test that would ideally use dependency injection
	// For simplicity in this example, we'll test the actual function with real directories

	tempDir := testhelp.TempDir(t)

	// Create nested directories
	testhelp.CreateTestSubdirs(t, tempDir, "subdir1", "subdir1/nested")

	watcher, err := NewFileWatcher(tempDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Failed to stop watcher: %v", err)
		}
	}()

	// Since addDirectory is private, we need to test it through Start()
	// This will call addDirectory internally
	err = watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Add a new directory after starting
	newDir := filepath.Join(tempDir, "new-dir")
	err = os.MkdirAll(newDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create new directory: %v", err)
	}

	// Give the file system watcher some time to detect the new directory
	time.Sleep(100 * time.Millisecond)
}

func TestRemoveDirectory(t *testing.T) {
	// Similar to TestAddDirectory, this would ideally use a mock
	// but we'll test with real directories

	tempDir := testhelp.TempDir(t)

	// Create nested directories
	dirs := testhelp.CreateTestSubdirs(t, tempDir, "subdir1", "subdir1/nested")

	watcher, err := NewFileWatcher(tempDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer func() {
		if err := watcher.Stop(); err != nil {
			t.Errorf("Failed to stop watcher: %v", err)
		}
	}()

	// Start the watcher
	err = watcher.Start()
	if err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Remove a directory
	err = os.RemoveAll(dirs["subdir1"])
	if err != nil {
		t.Fatalf("Failed to remove directory: %v", err)
	}

	// Give the file system watcher some time to detect the removal
	time.Sleep(100 * time.Millisecond)
}
