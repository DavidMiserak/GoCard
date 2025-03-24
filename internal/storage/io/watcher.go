// File: internal/storage/io/watcher.go

package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	logger        *Logger
	mu            sync.RWMutex
	closed        bool // Explicitly track if watcher is closed
}

// NewFileWatcher creates a new FileWatcher for the given root directory
func NewFileWatcher(rootDir string) (*FileWatcher, error) {
	// Check if the directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", rootDir)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	logger := NewLogger(os.Stdout, INFO)

	return &FileWatcher{
		watcher:       watcher,
		events:        make(chan FileEvent),
		errors:        make(chan error),
		watchedDirs:   make(map[string]bool),
		rootDir:       rootDir,
		isRunning:     false,
		ignoredFiles:  []string{".git", ".DS_Store", "Thumbs.db", "desktop.ini", "~$"},
		debounceDelay: 100 * time.Millisecond,
		logger:        logger,
		mu:            sync.RWMutex{},
		closed:        false,
	}, nil
}

// SetLogger sets the logger for this watcher
func (fw *FileWatcher) SetLogger(logger *Logger) {
	fw.logger = logger
}

// Start begins watching the root directory and its subdirectories
func (fw *FileWatcher) Start() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.isRunning {
		return fmt.Errorf("watcher is already running")
	}

	if fw.closed || fw.watcher == nil {
		return fmt.Errorf("watcher is closed or nil")
	}

	// Set running state before adding directories
	fw.isRunning = true

	// Add root directory to watcher (outside of lock)
	fw.mu.Unlock()
	err := fw.addDirectory(fw.rootDir)
	fw.mu.Lock()

	if err != nil {
		fw.isRunning = false
		return err
	}

	// Start the event processing goroutine
	go fw.processEvents()

	fw.logger.Debug("File watcher started for %s", fw.rootDir)
	return nil
}

// Stop stops watching and closes channels
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.isRunning || fw.closed || fw.watcher == nil {
		return nil
	}

	// Mark as not running and closed
	fw.isRunning = false
	fw.closed = true

	fw.logger.Debug("File watcher stopped")

	// Close the watcher (this will close the event channels)
	err := fw.watcher.Close()

	// Clear watched directories
	fw.watchedDirs = make(map[string]bool)

	return err
}

// Events returns the channel of file events
func (fw *FileWatcher) Events() <-chan FileEvent {
	return fw.events
}

// Errors returns the channel of errors
func (fw *FileWatcher) Errors() <-chan error {
	return fw.errors
}

// isClosed safely checks if the watcher is closed
func (fw *FileWatcher) isClosed() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.closed || fw.watcher == nil
}

// isWatcherRunning safely checks if the watcher is running
func (fw *FileWatcher) isWatcherRunning() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.isRunning && !fw.closed && fw.watcher != nil
}

// addDirectory adds a directory and its subdirectories to the watcher
func (fw *FileWatcher) addDirectory(dirPath string) error {
	// Skip if closed
	if fw.isClosed() {
		return fmt.Errorf("watcher is closed, cannot add directory: %s", dirPath)
	}

	// Add the directory itself
	err := fw.watcher.Add(dirPath)
	if err != nil {
		// Defensive coding: check for specific signs that the watcher is closed
		if strings.Contains(err.Error(), "closed") {
			fw.mu.Lock()
			fw.closed = true
			fw.mu.Unlock()
			return fmt.Errorf("watcher is closed, cannot add directory: %s", dirPath)
		}
		return fmt.Errorf("failed to watch directory %s: %w", dirPath, err)
	}

	fw.mu.Lock()
	fw.watchedDirs[dirPath] = true
	fw.mu.Unlock()

	fw.logger.Debug("Watching directory: %s", dirPath)

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

			// Skip if watcher is closed
			if fw.isClosed() {
				return fmt.Errorf("watcher is closed during walk")
			}

			err := fw.watcher.Add(path)
			if err != nil {
				// Check for closed watcher
				if strings.Contains(err.Error(), "closed") {
					fw.mu.Lock()
					fw.closed = true
					fw.mu.Unlock()
					return fmt.Errorf("watcher is closed, cannot add directory: %s", path)
				}
				return fmt.Errorf("failed to watch directory %s: %w", path, err)
			}

			fw.mu.Lock()
			fw.watchedDirs[path] = true
			fw.mu.Unlock()

			fw.logger.Debug("Watching directory: %s", path)
		}
		return nil
	})
}

// removeDirectory removes a directory from the watched list - with extra safeguards
func (fw *FileWatcher) removeDirectory(dirPath string) {
	// Skip the entire operation if the watcher is closed
	if fw.isClosed() {
		fw.logger.Debug("Watcher is closed, skipping removal: %s", dirPath)
		return
	}

	// Check if the watcher itself is nil
	if fw.watcher == nil {
		fw.logger.Debug("Watcher is nil, skipping removal: %s", dirPath)
		return
	}

	// Check if the directory is actually being watched
	fw.mu.Lock()
	_, exists := fw.watchedDirs[dirPath]
	if !exists {
		fw.mu.Unlock()
		fw.logger.Debug("Directory not in watch list, skipping removal: %s", dirPath)
		return
	}

	// Remove from our map first - this prevents other goroutines from
	// trying to remove the same directory
	delete(fw.watchedDirs, dirPath)

	// Collect subdirectories to remove while we have the lock
	var subdirsToRemove []string
	for watchedDir := range fw.watchedDirs {
		if strings.HasPrefix(watchedDir, dirPath+string(filepath.Separator)) {
			subdirsToRemove = append(subdirsToRemove, watchedDir)
		}
	}
	fw.mu.Unlock()

	// Now try to remove from the watcher - we do this after removing from our map
	// to ensure we don't try to access a deleted map entry if something fails
	if !fw.isClosed() && fw.watcher != nil {
		// Use a defer/recover to catch any panics from the fsnotify library
		func() {
			defer func() {
				if r := recover(); r != nil {
					fw.logger.Error("Panic when removing directory from watcher: %v", r)
				}
			}()

			// Try to remove - but don't panic if it fails
			err := fw.watcher.Remove(dirPath)
			if err != nil {
				fw.logger.Debug("Error removing directory from watcher: %s - %v", dirPath, err)

				// If we get a specific error about closed watcher, mark as closed
				if strings.Contains(err.Error(), "closed") {
					fw.mu.Lock()
					fw.closed = true
					fw.mu.Unlock()
				}
			} else {
				fw.logger.Debug("Stopped watching directory: %s", dirPath)
			}
		}()
	}

	// Remove subdirectories
	for _, subdir := range subdirsToRemove {
		fw.removeDirectory(subdir)
	}
}

// processEvents processes events from the watcher
func (fw *FileWatcher) processEvents() {
	// Create a map to debounce events
	eventDebounce := make(map[string]time.Time)

	for {
		// Exit early if watcher is not running
		if !fw.isWatcherRunning() {
			return
		}

		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				// Channel was closed, exit the goroutine
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

			// Skip further processing if watcher is closed
			if fw.isClosed() {
				return
			}

			// Handle directory creation
			if event.Op&fsnotify.Create == fsnotify.Create {
				// Check if it's a directory
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					// Skip if watcher is closed
					if fw.isClosed() {
						return
					}

					err := fw.addDirectory(event.Name)
					if err != nil {
						// Only send error if still running
						if fw.isWatcherRunning() {
							fw.errors <- err
						}
					}
				}
			}

			// Handle directory removal with extra care
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				fw.mu.RLock()
				isWatched := fw.watchedDirs[event.Name]
				fw.mu.RUnlock()

				if isWatched && !fw.isClosed() {
					fw.removeDirectory(event.Name)
				}
			}

			// Skip if closed before sending event
			if fw.isClosed() {
				return
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
				// Channel was closed, exit the goroutine
				return
			}

			// Only forward errors if still running
			if fw.isWatcherRunning() {
				fw.errors <- err
			} else {
				return
			}
		}
	}
}
