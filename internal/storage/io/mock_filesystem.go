// internal/storage/io/mock_filesystem.go
package io

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

// MockFileSystem implements FileSystem for testing
type MockFileSystem struct {
	mu       sync.RWMutex
	files    map[string][]byte
	dirs     map[string]bool
	fileInfo map[string]mockFileInfo

	// Error maps for simulating failures
	readErr   map[string]error
	writeErr  map[string]error
	mkdirErr  map[string]error
	removeErr map[string]error
	renameErr map[string]error
	statErr   map[string]error
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem() *MockFileSystem {
	fs := &MockFileSystem{
		files:     make(map[string][]byte),
		dirs:      make(map[string]bool),
		fileInfo:  make(map[string]mockFileInfo),
		readErr:   make(map[string]error),
		writeErr:  make(map[string]error),
		mkdirErr:  make(map[string]error),
		removeErr: make(map[string]error),
		renameErr: make(map[string]error),
		statErr:   make(map[string]error),
	}

	// Set up the root directory which always exists
	fs.dirs["/"] = true
	fs.fileInfo["/"] = mockFileInfo{
		name:    "/",
		size:    0,
		mode:    0755 | os.ModeDir,
		modTime: time.Now(),
		isDir:   true,
	}

	return fs
}

func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if err, ok := fs.readErr[path]; ok && err != nil {
		return nil, err
	}

	data, ok := fs.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return data, nil
}

func (fs *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err, ok := fs.writeErr[path]; ok && err != nil {
		return err
	}

	// Create parent directories automatically
	dir := filepath.Dir(path)
	fs.createParentDirs(dir)

	fs.files[path] = data
	fs.fileInfo[path] = mockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(data)),
		mode:    perm,
		modTime: time.Now(),
		isDir:   false,
	}

	return nil
}

// Helper method to create parent directories
func (fs *MockFileSystem) createParentDirs(path string) {
	if path == "/" || path == "." {
		return
	}

	// Recursively create parent directories
	parent := filepath.Dir(path)
	if parent != "/" && parent != "." {
		fs.createParentDirs(parent)
	}

	// Create this directory if it doesn't exist
	fs.dirs[path] = true
	if _, ok := fs.fileInfo[path]; !ok {
		fs.fileInfo[path] = mockFileInfo{
			name:    filepath.Base(path),
			size:    0,
			mode:    0755 | os.ModeDir,
			modTime: time.Now(),
			isDir:   true,
		}
	}
}

func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err, ok := fs.mkdirErr[path]; ok && err != nil {
		return err
	}

	// Create the directory and all parent directories
	fs.createParentDirs(path)

	return nil
}

func (fs *MockFileSystem) RemoveAll(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err, ok := fs.removeErr[path]; ok && err != nil {
		return err
	}

	// Don't allow removing the root directory
	if path == "/" {
		return errors.New("cannot remove root directory")
	}

	// Remove the directory and all its contents
	delete(fs.dirs, path)
	delete(fs.fileInfo, path)

	// Remove all files and directories that start with this path
	for filePath := range fs.files {
		if filePath == path || strings.HasPrefix(filePath, path+"/") {
			delete(fs.files, filePath)
			delete(fs.fileInfo, filePath)
		}
	}

	for dirPath := range fs.dirs {
		if dirPath == path || strings.HasPrefix(dirPath, path+"/") {
			delete(fs.dirs, dirPath)
			delete(fs.fileInfo, dirPath)
		}
	}

	return nil
}

func (fs *MockFileSystem) Remove(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err, ok := fs.removeErr[path]; ok && err != nil {
		return err
	}

	// Don't allow removing the root directory
	if path == "/" {
		return errors.New("cannot remove root directory")
	}

	// Check if it's a file
	if _, ok := fs.files[path]; ok {
		delete(fs.files, path)
		delete(fs.fileInfo, path)
		return nil
	}

	// Check if it's a directory
	if _, ok := fs.dirs[path]; ok {
		// Check if the directory is empty
		for filePath := range fs.files {
			if strings.HasPrefix(filePath, path+"/") {
				return errors.New("directory not empty")
			}
		}

		for dirPath := range fs.dirs {
			if dirPath != path && strings.HasPrefix(dirPath, path+"/") {
				return errors.New("directory not empty")
			}
		}

		delete(fs.dirs, path)
		delete(fs.fileInfo, path)
		return nil
	}

	return os.ErrNotExist
}

func (fs *MockFileSystem) Rename(oldpath, newpath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err, ok := fs.renameErr[oldpath]; ok && err != nil {
		return err
	}

	// Create parent directories for the new path
	dir := filepath.Dir(newpath)
	fs.createParentDirs(dir)

	// Check if the source exists
	if fileData, ok := fs.files[oldpath]; ok {
		// It's a file
		fs.files[newpath] = fileData
		fs.fileInfo[newpath] = fs.fileInfo[oldpath]
		delete(fs.files, oldpath)
		delete(fs.fileInfo, oldpath)
		return nil
	}

	if _, ok := fs.dirs[oldpath]; ok {
		// It's a directory
		fs.dirs[newpath] = true
		fs.fileInfo[newpath] = fs.fileInfo[oldpath]
		delete(fs.dirs, oldpath)
		delete(fs.fileInfo, oldpath)

		// Rename all nested files and directories
		for filePath, fileData := range fs.files {
			if strings.HasPrefix(filePath, oldpath+"/") {
				newFilePath := newpath + filePath[len(oldpath):]
				fs.files[newFilePath] = fileData
				fs.fileInfo[newFilePath] = fs.fileInfo[filePath]
				delete(fs.files, filePath)
				delete(fs.fileInfo, filePath)
			}
		}

		for dirPath := range fs.dirs {
			if dirPath != oldpath && strings.HasPrefix(dirPath, oldpath+"/") {
				newDirPath := newpath + dirPath[len(oldpath):]
				fs.dirs[newDirPath] = true
				fs.fileInfo[newDirPath] = fs.fileInfo[dirPath]
				delete(fs.dirs, dirPath)
				delete(fs.fileInfo, dirPath)
			}
		}

		return nil
	}

	return os.ErrNotExist
}

func (fs *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if err, ok := fs.statErr[path]; ok && err != nil {
		return nil, err
	}

	// Check if it's a file
	if _, ok := fs.files[path]; ok {
		info := fs.fileInfo[path]
		return &info, nil
	}

	// Check if it's a directory
	if _, ok := fs.dirs[path]; ok {
		info := fs.fileInfo[path]
		return &info, nil
	}

	return nil, os.ErrNotExist
}

// Helper methods for testing

// SetupFile adds a file to the mock filesystem for testing
func (fs *MockFileSystem) SetupFile(path string, content []byte) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Create parent directories
	dir := filepath.Dir(path)
	fs.createParentDirs(dir)

	fs.files[path] = content
	fs.fileInfo[path] = mockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(content)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}
}

// SetupDir adds a directory to the mock filesystem for testing
func (fs *MockFileSystem) SetupDir(path string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.createParentDirs(path)
}

// InjectError sets up an error for a specific operation and path
func (fs *MockFileSystem) InjectError(op string, path string, err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	switch op {
	case "read":
		fs.readErr[path] = err
	case "write":
		fs.writeErr[path] = err
	case "mkdir":
		fs.mkdirErr[path] = err
	case "remove":
		fs.removeErr[path] = err
	case "rename":
		fs.renameErr[path] = err
	case "stat":
		fs.statErr[path] = err
	}
}
