// File: internal/storage/io/file_ops_test.go

package io

import (
	"errors"
	"testing"
)

func TestEnsureDirectoryExistsWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Test creating a new directory
	helper.UseWithinTest(func() {
		// Test creating a directory that doesn't exist
		err := EnsureDirectoryExists("/test-dir")
		if err != nil {
			t.Errorf("Failed to create directory: %v", err)
		}
		helper.AssertDirExists("/test-dir")

		// Test with existing directory (should not error)
		err = EnsureDirectoryExists("/test-dir")
		if err != nil {
			t.Errorf("Failed on existing directory: %v", err)
		}

		// Test nested directories
		err = EnsureDirectoryExists("/test-dir/nested1/nested2")
		if err != nil {
			t.Errorf("Failed to create nested directories: %v", err)
		}
		helper.AssertDirExists("/test-dir/nested1/nested2")

		// Test error injection
		expectedErr := errors.New("permission denied")
		helper.InjectError("mkdir", "/error-dir", expectedErr)

		err = EnsureDirectoryExists("/error-dir")
		if err == nil {
			t.Errorf("Expected error but got none")
		}
	})
}

func TestFileExistsWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Setup test files
	helper.SetupDir("/test-dir")
	helper.SetupFile("/test-dir/test.txt", "content")

	helper.UseWithinTest(func() {
		// Test existing file
		if !FileExists("/test-dir/test.txt") {
			t.Errorf("FileExists failed to detect existing file")
		}

		// Test non-existent file
		if FileExists("/test-dir/non-existent.txt") {
			t.Errorf("FileExists incorrectly detected non-existent file")
		}

		// Test directory (should return false as it's not a file)
		if FileExists("/test-dir") {
			t.Errorf("FileExists incorrectly identified directory as file")
		}
	})
}

func TestDirectoryExistsWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Setup test directory
	helper.SetupDir("/test-dir")
	helper.SetupFile("/test-dir/test.txt", "content")

	helper.UseWithinTest(func() {
		// Test existing directory
		exists, err := DirectoryExists("/test-dir")
		if err != nil {
			t.Errorf("DirectoryExists error: %v", err)
		}
		if !exists {
			t.Errorf("DirectoryExists failed to detect existing directory")
		}

		// Test non-existent directory
		exists, err = DirectoryExists("/non-existent-dir")
		if err != nil {
			t.Errorf("DirectoryExists error on non-existent dir: %v", err)
		}
		if exists {
			t.Errorf("DirectoryExists incorrectly detected non-existent directory")
		}

		// Test file (should return false as it's not a directory)
		exists, err = DirectoryExists("/test-dir/test.txt")
		if err != nil {
			t.Errorf("DirectoryExists error on file: %v", err)
		}
		if exists {
			t.Errorf("DirectoryExists incorrectly identified file as directory")
		}

		// Test error injection
		expectedErr := errors.New("permission denied")
		helper.InjectError("stat", "/error-dir", expectedErr)

		_, err = DirectoryExists("/error-dir")
		if err == nil {
			t.Errorf("Expected error but got none")
		}
	})
}

func TestReadFileContentWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Create a test file
	content := "Hello, World!"
	helper.SetupFile("/test-dir/test.txt", content)

	helper.UseWithinTest(func() {
		// Read the file
		readContent, err := ReadFileContent("/test-dir/test.txt")
		if err != nil {
			t.Errorf("ReadFileContent error: %v", err)
		}

		if string(readContent) != content {
			t.Errorf("Expected content %q, got %q", content, string(readContent))
		}

		// Test non-existent file
		_, err = ReadFileContent("/test-dir/non-existent.txt")
		if err == nil {
			t.Errorf("Expected error reading non-existent file, got nil")
		}

		// Test error injection
		expectedErr := errors.New("permission denied")
		helper.InjectError("read", "/test-dir/test.txt", expectedErr)

		_, err = ReadFileContent("/test-dir/test.txt")
		if err == nil {
			t.Errorf("Expected injected error but got none")
		}
	})
}

func TestWriteFileContentWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Setup test directory
	helper.SetupDir("/test-dir")

	helper.UseWithinTest(func() {
		// Write to a new file
		content := "Hello, World!"
		err := WriteFileContent("/test-dir/new.txt", []byte(content))
		if err != nil {
			t.Errorf("WriteFileContent error: %v", err)
		}

		helper.AssertFileExists("/test-dir/new.txt")
		helper.AssertFileContent("/test-dir/new.txt", content)

		// Write to an existing file (should overwrite)
		newContent := "New content"
		err = WriteFileContent("/test-dir/new.txt", []byte(newContent))
		if err != nil {
			t.Errorf("WriteFileContent error on existing file: %v", err)
		}

		helper.AssertFileContent("/test-dir/new.txt", newContent)

		// Write to a file in a non-existent directory (should create the directory)
		nestedContent := "Nested content"
		err = WriteFileContent("/test-dir/nested/deep.txt", []byte(nestedContent))
		if err != nil {
			t.Errorf("WriteFileContent error with nested directories: %v", err)
		}

		helper.AssertFileExists("/test-dir/nested/deep.txt")
		helper.AssertFileContent("/test-dir/nested/deep.txt", nestedContent)

		// Test error injection
		expectedErr := errors.New("disk full")
		helper.InjectError("write", "/test-dir/error.txt", expectedErr)

		err = WriteFileContent("/test-dir/error.txt", []byte("error content"))
		if err == nil {
			t.Errorf("Expected injected error but got none")
		}
	})
}

func TestDeleteFileWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Create a test file
	helper.SetupFile("/test-dir/test.txt", "content")

	helper.UseWithinTest(func() {
		// Delete the file
		err := DeleteFile("/test-dir/test.txt")
		if err != nil {
			t.Errorf("DeleteFile error: %v", err)
		}

		helper.AssertFileDoesNotExist("/test-dir/test.txt")

		// Test deleting a non-existent file
		err = DeleteFile("/test-dir/non-existent.txt")
		if err == nil {
			t.Errorf("Expected error deleting non-existent file, got nil")
		}

		// Test error injection
		helper.SetupFile("/test-dir/error.txt", "content")
		expectedErr := errors.New("permission denied")
		helper.InjectError("remove", "/test-dir/error.txt", expectedErr)

		err = DeleteFile("/test-dir/error.txt")
		if err == nil {
			t.Errorf("Expected injected error but got none")
		}
	})
}

func TestMoveFileWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Create a test file
	content := "Test content"
	helper.SetupFile("/test-dir/source.txt", content)

	helper.UseWithinTest(func() {
		// Move to a new location
		err := MoveFile("/test-dir/source.txt", "/test-dir/target.txt")
		if err != nil {
			t.Errorf("MoveFile error: %v", err)
		}

		helper.AssertFileDoesNotExist("/test-dir/source.txt")
		helper.AssertFileExists("/test-dir/target.txt")
		helper.AssertFileContent("/test-dir/target.txt", content)

		// Move to a nested directory that doesn't exist yet
		err = MoveFile("/test-dir/target.txt", "/test-dir/nested/moved.txt")
		if err != nil {
			t.Errorf("MoveFile error to nested directory: %v", err)
		}

		helper.AssertFileDoesNotExist("/test-dir/target.txt")
		helper.AssertFileExists("/test-dir/nested/moved.txt")
		helper.AssertFileContent("/test-dir/nested/moved.txt", content)

		// Test error handling - source doesn't exist
		err = MoveFile("/test-dir/non-existent.txt", "/test-dir/error.txt")
		if err == nil {
			t.Errorf("Expected error moving non-existent file, got nil")
		}

		// Test error injection
		helper.SetupFile("/test-dir/error-source.txt", "error content")
		expectedErr := errors.New("permission denied")
		helper.InjectError("rename", "/test-dir/error-source.txt", expectedErr)

		err = MoveFile("/test-dir/error-source.txt", "/test-dir/error-target.txt")
		if err == nil {
			t.Errorf("Expected injected error but got none")
		}
	})
}

func TestRenameDirectoryWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Create a test directory structure
	helper.SetupDir("/source-dir")
	helper.SetupDir("/source-dir/subdir")
	helper.SetupFile("/source-dir/test.txt", "content")
	helper.SetupFile("/source-dir/subdir/nested.txt", "nested content")

	helper.UseWithinTest(func() {
		// Rename the directory
		err := RenameDirectory("/source-dir", "/target-dir")
		if err != nil {
			t.Errorf("RenameDirectory error: %v", err)
		}

		helper.AssertFileDoesNotExist("/source-dir")
		helper.AssertDirExists("/target-dir")
		helper.AssertDirExists("/target-dir/subdir")
		helper.AssertFileExists("/target-dir/test.txt")
		helper.AssertFileExists("/target-dir/subdir/nested.txt")
		helper.AssertFileContent("/target-dir/test.txt", "content")
		helper.AssertFileContent("/target-dir/subdir/nested.txt", "nested content")

		// Test renaming to a nested directory that doesn't exist
		err = RenameDirectory("/target-dir", "/new-parent/renamed-dir")
		if err != nil {
			t.Errorf("RenameDirectory error to nested directory: %v", err)
		}

		helper.AssertFileDoesNotExist("/target-dir")
		helper.AssertDirExists("/new-parent/renamed-dir")
		helper.AssertFileExists("/new-parent/renamed-dir/test.txt")

		// Test error injection
		helper.SetupDir("/error-dir")
		expectedErr := errors.New("permission denied")
		helper.InjectError("rename", "/error-dir", expectedErr)

		err = RenameDirectory("/error-dir", "/error-target")
		if err == nil {
			t.Errorf("Expected injected error but got none")
		}
	})
}

func TestDeleteDirectoryWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	// Create a test directory with files and subdirectories
	helper.SetupDir("/test-dir")
	helper.SetupDir("/test-dir/subdir")
	helper.SetupFile("/test-dir/file1.txt", "content1")
	helper.SetupFile("/test-dir/subdir/file2.txt", "content2")

	helper.UseWithinTest(func() {
		// Delete the directory
		err := DeleteDirectory("/test-dir")
		if err != nil {
			t.Errorf("DeleteDirectory error: %v", err)
		}

		helper.AssertFileDoesNotExist("/test-dir")
		helper.AssertFileDoesNotExist("/test-dir/file1.txt")
		helper.AssertFileDoesNotExist("/test-dir/subdir/file2.txt")

		// Test deleting a non-existent directory
		err = DeleteDirectory("/non-existent-dir")
		if err != nil {
			t.Errorf("Expected no error when deleting non-existent directory, got: %v", err)
		}

		// Test error injection
		helper.SetupDir("/error-dir")
		expectedErr := errors.New("permission denied")
		helper.InjectError("remove", "/error-dir", expectedErr)

		err = DeleteDirectory("/error-dir")
		if err == nil {
			t.Errorf("Expected injected error but got none")
		}
	})
}

// TestFileSystemSwapping tests proper swapping of the default filesystem
func TestFileSystemSwapping(t *testing.T) {
	helper := NewTestHelper(t)

	// Set up mock filesystem
	helper.SetupFile("/test.txt", "test content")

	// Verify global functions use the real filesystem by default
	beforeSwap := FileExists("/test.txt")
	if beforeSwap {
		t.Errorf("File shouldn't exist in real filesystem before swap")
	}

	// Use helper to swap filesystems
	helper.UseWithinTest(func() {
		// Verify the mock is now used
		duringSwap := FileExists("/test.txt")
		if !duringSwap {
			t.Errorf("File should exist in mock filesystem during swap")
		}

		// Check file content
		content, err := ReadFileContent("/test.txt")
		if err != nil {
			t.Errorf("ReadFileContent error: %v", err)
		}
		if string(content) != "test content" {
			t.Errorf("Expected 'test content', got '%s'", string(content))
		}
	})

	// Verify we're back to the real filesystem
	afterSwap := FileExists("/test.txt")
	if afterSwap {
		t.Errorf("File shouldn't exist in real filesystem after swap")
	}
}

// Integration test demonstrating how all operations work together
func TestIntegratedOperationsWithMockFS(t *testing.T) {
	helper := NewTestHelper(t)

	helper.UseWithinTest(func() {
		// 1. Create directories
		err := EnsureDirectoryExists("/project/src")
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// 2. Write files
		err = WriteFileContent("/project/src/main.go", []byte("package main\n\nfunc main() {}"))
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		err = WriteFileContent("/project/README.md", []byte("# Project"))
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// 3. Verify files exist
		if !FileExists("/project/src/main.go") {
			t.Errorf("File should exist")
		}

		dirExists, err := DirectoryExists("/project/src")
		if err != nil || !dirExists {
			t.Errorf("Directory should exist")
		}

		// 4. Move a file
		err = MoveFile("/project/README.md", "/project/docs/README.md")
		if err != nil {
			t.Fatalf("Failed to move file: %v", err)
		}

		helper.AssertFileExists("/project/docs/README.md")
		helper.AssertFileDoesNotExist("/project/README.md")

		// 5. Rename a directory
		err = RenameDirectory("/project/src", "/project/code")
		if err != nil {
			t.Fatalf("Failed to rename directory: %v", err)
		}

		helper.AssertDirExists("/project/code")
		helper.AssertFileDoesNotExist("/project/src")
		helper.AssertFileExists("/project/code/main.go")

		// 6. Delete a file
		err = DeleteFile("/project/code/main.go")
		if err != nil {
			t.Fatalf("Failed to delete file: %v", err)
		}

		helper.AssertFileDoesNotExist("/project/code/main.go")

		// 7. Delete a directory
		err = DeleteDirectory("/project")
		if err != nil {
			t.Fatalf("Failed to delete directory: %v", err)
		}

		helper.AssertFileDoesNotExist("/project")
	})
}
