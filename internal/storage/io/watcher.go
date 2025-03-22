// Package io provides file system operations for the GoCard storage system.
package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileEvent represents a file system event
type FileEvent struct {
	Path      string
	Operation string // "create", "write", "remove", "rename", "chmod"
}

// FileWatcher watches for changes to files in a directory
type FileWatcher struct {
	watcher       *fsnotify.Watcher
	events        chan FileEvent
	errors        chan error
	watchedDirs   map[string]bool
	rootDir       string
	isRunning     bool
	ignoredFiles  []string
	debounceDelay time.Duration
}

// NewFileWatcher creates a new FileWatcher for the given root directory
func NewFileWatcher(rootDir string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &FileWatcher{
		watcher:       watcher,
		events:        make(chan FileEvent),
		errors:        make(chan error),
		watchedDirs:   make(map[string]bool),
		rootDir:       rootDir,
		isRunning:     false,
		ignoredFiles:  []string{".git", ".DS_Store", "Thumbs.db", "desktop.ini", "~$"},
		debounceDelay: 100 * time.Millisecond,
	}, nil
}

// Start begins watching the root directory and its subdirectories
func (fw *FileWatcher) Start() error {
	if fw.isRunning {
		return fmt.Errorf("watcher is already running")
	}

	// Add root directory to watcher
	if err := fw.addDirectory(fw.rootDir); err != nil {
		return err
	}

	// Start the event processing goroutine
	go fw.processEvents()

	fw.isRunning = true
	return nil
}

// Stop stops watching and closes channels
func (fw *FileWatcher) Stop() error {
	if !fw.isRunning {
		return nil
	}

	fw.isRunning = false
	return fw.watcher.Close()
}

// Events returns the channel of file events
func (fw *FileWatcher) Events() <-chan FileEvent {
	return fw.events
}

// Errors returns the channel of errors
func (fw *FileWatcher) Errors() <-chan error {
	return fw.errors
}

// addDirectory adds a directory and its subdirectories to the watcher
func (fw *FileWatcher) addDirectory(dirPath string) error {
	// Add the directory itself
	if err := fw.watcher.Add(dirPath); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", dirPath, err)
	}
	fw.watchedDirs[dirPath] = true

	// Add all subdirectories recursively
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != dirPath {
			// Skip ignored directories
			for _, ignored := range fw.ignoredFiles {
				if strings.Contains(path, ignored) {
					return filepath.SkipDir
				}
			}

			if err := fw.watcher.Add(path); err != nil {
				return fmt.Errorf("failed to watch directory %s: %w", path, err)
			}
			fw.watchedDirs[path] = true
		}
		return nil
	})
}

// removeDirectory removes a directory from the watched list
func (fw *FileWatcher) removeDirectory(dirPath string) {
	// Remove the directory from the watcher
	_ = fw.watcher.Remove(dirPath)
	delete(fw.watchedDirs, dirPath)

	// Remove any subdirectories that were being watched
	for watchedDir := range fw.watchedDirs {
		if strings.HasPrefix(watchedDir, dirPath+string(filepath.Separator)) {
			_ = fw.watcher.Remove(watchedDir)
			delete(fw.watchedDirs, watchedDir)
		}
	}
}

// processEvents processes events from the watcher
func (fw *FileWatcher) processEvents() {
	// Create a map to debounce events
	eventDebounce := make(map[string]time.Time)

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Skip ignored files
			isIgnored := false
			for _, ignored := range fw.ignoredFiles {
				if strings.Contains(event.Name, ignored) {
					isIgnored = true
					break
				}
			}
			if isIgnored {
				continue
			}

			// Check if this event should be debounced
			lastEvent, exists := eventDebounce[event.Name]
			now := time.Now()
			if exists && now.Sub(lastEvent) < fw.debounceDelay {
				// Update the timestamp and skip this event
				eventDebounce[event.Name] = now
				continue
			}
			eventDebounce[event.Name] = now

			// Handle directory creation
			if event.Op&fsnotify.Create == fsnotify.Create {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					err := fw.addDirectory(event.Name)
					if err != nil {
						fw.errors <- err
					}
				}
			}

			// Handle directory removal
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				if _, exists := fw.watchedDirs[event.Name]; exists {
					fw.removeDirectory(event.Name)
				}
			}

			// Send the event to the events channel
			operation := "unknown"
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				operation = "create"
			case event.Op&fsnotify.Write == fsnotify.Write:
				operation = "write"
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				operation = "remove"
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				operation = "rename"
			case event.Op&fsnotify.Chmod == fsnotify.Chmod:
				operation = "chmod"
			}

			fw.events <- FileEvent{
				Path:      event.Name,
				Operation: operation,
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.errors <- err
		}
	}
}
