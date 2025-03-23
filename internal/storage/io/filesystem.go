// internal/storage/io/filesystem.go
package io

import (
	"os"
)

// FileSystem defines an interface for filesystem operations
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(path string) error
	Remove(path string) error
	Rename(oldpath, newpath string) error
	Stat(path string) (os.FileInfo, error)
}

// RealFileSystem implements FileSystem using actual OS operations
type RealFileSystem struct{}

func NewRealFileSystem() *RealFileSystem {
	return &RealFileSystem{}
}

func (fs *RealFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (fs *RealFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (fs *RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *RealFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (fs *RealFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (fs *RealFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (fs *RealFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
