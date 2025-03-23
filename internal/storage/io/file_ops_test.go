// internal/storage/io/file_ops_test.go
package io

import (
	"os"
	"path/filepath"
	"testing"

	testhelp "github.com/DavidMiserak/GoCard/internal/testing"
)

func TestEnsureDirectoryExists(t *testing.T) {
	// Create a temporary testing directory
	tempDir := testhelp.TempDir(t)

	// Test creating a new directory
	testDir := filepath.Join(tempDir, "test-dir")
	err := EnsureDirectoryExists(testDir)
	if err != nil {
		t.Errorf("Failed to create directory: %v", err)
	}

	testhelp.AssertFileExists(t, testDir)

	// Test with existing directory (should not error)
	err = EnsureDirectoryExists(testDir)
	if err != nil {
		t.Errorf("Failed on existing directory: %v", err)
	}

	// Test nested directories
	nestedDir := filepath.Join(testDir, "nested1", "nested2")
	err = EnsureDirectoryExists(nestedDir)
	if err != nil {
		t.Errorf("Failed to create nested directories: %v", err)
	}

	testhelp.AssertFileExists(t, nestedDir)
}

func TestGetAbsolutePath(t *testing.T) {
	// Test relative path
	relPath := "test/path"
	absPath, err := GetAbsolutePath(relPath)
	if err != nil {
		t.Errorf("Failed to get absolute path: %v", err)
	}

	// The absolute path should contain the relative path
	if filepath.Base(absPath) != filepath.Base(relPath) {
		t.Errorf("Base of absolute path %s doesn't match base of relative path %s",
			absPath, relPath)
	}

	// Test already absolute path
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	absResult, err := GetAbsolutePath(currentDir)
	if err != nil {
		t.Errorf("Failed on absolute path: %v", err)
	}

	if absResult != currentDir {
		t.Errorf("Expected %s, got %s", currentDir, absResult)
	}
}

func TestFileExists(t *testing.T) {
	tempDir := testhelp.TempDir(t)

	// Create a test file
	testFile := testhelp.WriteTestFile(t, tempDir, "test.txt", "content")

	// Test existing file
	if !FileExists(testFile) {
		t.Errorf("FileExists failed to detect existing file")
	}

	// Test non-existent file
	nonExistentFile := filepath.Join(tempDir, "non-existent.txt")
	if FileExists(nonExistentFile) {
		t.Errorf("FileExists incorrectly detected non-existent file")
	}

	// Test directory (should return false as it's not a file)
	if FileExists(tempDir) {
		t.Errorf("FileExists incorrectly identified directory as file")
	}
}

func TestDirectoryExists(t *testing.T) {
	tempDir := testhelp.TempDir(t)

	// Test existing directory
	exists, err := DirectoryExists(tempDir)
	if err != nil {
		t.Errorf("DirectoryExists error: %v", err)
	}
	if !exists {
		t.Errorf("DirectoryExists failed to detect existing directory")
	}

	// Test non-existent directory
	nonExistentDir := filepath.Join(tempDir, "non-existent-dir")
	exists, err = DirectoryExists(nonExistentDir)
	if err != nil {
		t.Errorf("DirectoryExists error on non-existent dir: %v", err)
	}
	if exists {
		t.Errorf("DirectoryExists incorrectly detected non-existent directory")
	}

	// Test file (should return false as it's not a directory)
	testFile := testhelp.WriteTestFile(t, tempDir, "test.txt", "content")
	exists, err = DirectoryExists(testFile)
	if err != nil {
		t.Errorf("DirectoryExists error on file: %v", err)
	}
	if exists {
		t.Errorf("DirectoryExists incorrectly identified file as directory")
	}
}

func TestReadFileContent(t *testing.T) {
	tempDir := testhelp.TempDir(t)

	// Create a test file
	content := "Hello, World!"
	testFile := testhelp.WriteTestFile(t, tempDir, "test.txt", content)

	// Read the file
	readContent, err := ReadFileContent(testFile)
	if err != nil {
		t.Errorf("ReadFileContent error: %v", err)
	}

	if string(readContent) != content {
		t.Errorf("Expected content %q, got %q", content, string(readContent))
	}

	// Test non-existent file
	nonExistentFile := filepath.Join(tempDir, "non-existent.txt")
	_, err = ReadFileContent(nonExistentFile)
	if err == nil {
		t.Errorf("Expected error reading non-existent file, got nil")
	}
}

func TestWriteFileContent(t *testing.T) {
	tempDir := testhelp.TempDir(t)

	// Write to a new file
	content := "Hello, World!"
	newFile := filepath.Join(tempDir, "new.txt")

	err := WriteFileContent(newFile, []byte(content))
	if err != nil {
		t.Errorf("WriteFileContent error: %v", err)
	}

	testhelp.AssertFileExists(t, newFile)
	testhelp.AssertFileContent(t, newFile, content)

	// Write to an existing file (should overwrite)
	newContent := "New content"
	err = WriteFileContent(newFile, []byte(newContent))
	if err != nil {
		t.Errorf("WriteFileContent error on existing file: %v", err)
	}

	testhelp.AssertFileContent(t, newFile, newContent)

	// Write to a file in a non-existent directory (should create the directory)
	nestedFile := filepath.Join(tempDir, "nested", "deep.txt")
	err = WriteFileContent(nestedFile, []byte(content))
	if err != nil {
		t.Errorf("WriteFileContent error with nested directories: %v", err)
	}

	testhelp.AssertFileExists(t, nestedFile)
	testhelp.AssertFileContent(t, nestedFile, content)
}

func TestDeleteFile(t *testing.T) {
	tempDir := testhelp.TempDir(t)

	// Create a test file
	testFile := testhelp.WriteTestFile(t, tempDir, "test.txt", "content")

	// Delete the file
	err := DeleteFile(testFile)
	if err != nil {
		t.Errorf("DeleteFile error: %v", err)
	}

	testhelp.AssertFileDoesNotExist(t, testFile)

	// Test deleting a non-existent file
	err = DeleteFile(testFile) // File was already deleted
	if err == nil {
		t.Errorf("Expected error deleting non-existent file, got nil")
	}
}

func TestMoveFile(t *testing.T) {
	tempDir := testhelp.TempDir(t)

	// Create a test file
	content := "Test content"
	sourceFile := testhelp.WriteTestFile(t, tempDir, "source.txt", content)

	// Move to a new location
	targetFile := filepath.Join(tempDir, "target.txt")
	err := MoveFile(sourceFile, targetFile)
	if err != nil {
		t.Errorf("MoveFile error: %v", err)
	}

	testhelp.AssertFileDoesNotExist(t, sourceFile)
	testhelp.AssertFileExists(t, targetFile)
	testhelp.AssertFileContent(t, targetFile, content)

	// Move to a nested directory that doesn't exist yet
	nestedTarget := filepath.Join(tempDir, "nested", "moved.txt")
	err = MoveFile(targetFile, nestedTarget)
	if err != nil {
		t.Errorf("MoveFile error to nested directory: %v", err)
	}

	testhelp.AssertFileDoesNotExist(t, targetFile)
	testhelp.AssertFileExists(t, nestedTarget)
	testhelp.AssertFileContent(t, nestedTarget, content)
}

func TestRenameDirectory(t *testing.T) {
	tempDir := testhelp.TempDir(t)

	// Create a test directory with a file
	testDirs := testhelp.CreateTestSubdirs(t, tempDir, "source-dir")
	sourceDir := testDirs["source-dir"]
	testhelp.WriteTestFile(t, sourceDir, "test.txt", "content")

	// Rename the directory
	targetDir := filepath.Join(tempDir, "target-dir")
	err := RenameDirectory(sourceDir, targetDir)
	if err != nil {
		t.Errorf("RenameDirectory error: %v", err)
	}

	testhelp.AssertFileDoesNotExist(t, sourceDir)
	testhelp.AssertFileExists(t, targetDir)

	// Check that the file was moved too
	targetFile := filepath.Join(targetDir, "test.txt")
	testhelp.AssertFileExists(t, targetFile)

	// Test renaming to a nested directory that doesn't exist
	nestedTargetDir := filepath.Join(tempDir, "nested", "renamed-dir")
	err = RenameDirectory(targetDir, nestedTargetDir)
	if err != nil {
		t.Errorf("RenameDirectory error to nested directory: %v", err)
	}

	testhelp.AssertFileDoesNotExist(t, targetDir)
	testhelp.AssertFileExists(t, nestedTargetDir)
}

func TestDeleteDirectory(t *testing.T) {
	tempDir := testhelp.TempDir(t)

	// Create a test directory with files and subdirectories
	testDirs := testhelp.CreateTestSubdirs(t, tempDir, "test-dir", "test-dir/subdir")
	testDir := testDirs["test-dir"]

	testhelp.WriteTestFile(t, testDir, "file1.txt", "content1")
	testhelp.WriteTestFile(t, testDirs["test-dir/subdir"], "file2.txt", "content2")

	// Delete the directory
	err := DeleteDirectory(testDir)
	if err != nil {
		t.Errorf("DeleteDirectory error: %v", err)
	}

	testhelp.AssertFileDoesNotExist(t, testDir)

	// Test deleting a non-existent directory
	err = DeleteDirectory(testDir) // Directory was already deleted
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent directory, got: %v", err)
	}
}
