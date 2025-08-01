// File: internal/data/editor_test.go

package data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
)

func TestGetShellEditor(t *testing.T) {
	// Test assumes vi and vim exist in env where test is run
	// and "nonexistentEditor123" doesn't exist

	originalEditor := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", originalEditor)

	testCases := []struct {
		name                  string
		editorEnv             string // to set
		expectError           bool
		lookPathShouldContain string // e.g. LookPath has "vi" in path
	}{
		{
			name:                  "$EDITOR not set, fallback to vi",
			editorEnv:             "",
			expectError:           false,
			lookPathShouldContain: "vi",
		},
		{
			name:                  "$EDITOR set to vim",
			editorEnv:             "vim",
			expectError:           false,
			lookPathShouldContain: "vim",
		},
		{
			name:                  "$EDITOR set to nonexistent editor",
			editorEnv:             "nonexistentEditor123",
			expectError:           true,
			lookPathShouldContain: "",
		},
	}

	// Run tests
	for _, testCase := range testCases {
		os.Setenv("EDITOR", testCase.editorEnv)
		result, err := getShellEditor()

		switch {
		case testCase.expectError && err != nil:
			// test passes
		case testCase.expectError && err == nil:
			t.Errorf("Expected error but got none")
		case !testCase.expectError && err != nil:
			t.Errorf("Unexpected error: %v", err)
		case !testCase.expectError && err == nil:
			pathIncludesEditor := strings.Contains(result, testCase.lookPathShouldContain)
			if !pathIncludesEditor {
				t.Errorf("Expected path to include %q, got %q", testCase.lookPathShouldContain, result)
			}
		}
	}
}

func TestCreateTmpFileWithText(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{name: "Empty text", text: ""},
		{name: "Single line text", text: "Hello World!"},
		{name: "Multi line text", text: "Hello\nWorld\n!"},
	}

	expectedFilenamePrefix := "GoCard-tmp-"
	expectedFilenameSuffix := ".md"

	for _, testCase := range testCases {
		// Run function
		filePath, err := createTmpFileWithText(testCase.text)
		defer os.Remove(filePath)

		if err != nil {
			t.Errorf("%q returned unexpected %v", testCase.name, err)
		}

		// Verify existence of file
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("createTmpFileWithText didnt create file at %s", filePath)
		}

		// Filename checks
		fileName := filepath.Base(filePath)
		fileNameHasExpectedPrefix := strings.HasPrefix(fileName, expectedFilenamePrefix)
		fileNameHasExpectedSuffix := strings.HasSuffix(fileName, expectedFilenameSuffix)

		if !fileNameHasExpectedPrefix {
			t.Errorf("Expected filename prefix %q in file %q", expectedFilenamePrefix, fileName)
		}
		if !fileNameHasExpectedSuffix {
			t.Errorf("Expected filename suffix %q in file %q", expectedFilenameSuffix, fileName)
		}

		// Check correct content exists
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Test %q failed to read file: %v", testCase.name, err)
			continue
		}
		if string(content) != testCase.text {
			t.Errorf("Test %q expected content %q but got %q", testCase.name, testCase.text, string(content))
		}

	}
}

func TestCreateTmpFileWithTemplate(t *testing.T) {
	expectedFilenamePrefix := "GoCard-tmp-"
	expectedFilenameSuffix := ".md"

	// Run function
	filePath, err := CreateTmpFileWithTemplate()
	defer os.Remove(filePath)
	if err != nil {
		t.Errorf("CreateTmpFileWithTemplate returned unexpected %v", err)
	}

	// Verify existence of file
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("CreateTmpFileWithTemplate didnt create file at %s", filePath)
	}

	// Filename checks
	fileName := filepath.Base(filePath)
	fileNameHasExpectedPrefix := strings.HasPrefix(fileName, expectedFilenamePrefix)
	fileNameHasExpectedSuffix := strings.HasSuffix(fileName, expectedFilenameSuffix)
	if !fileNameHasExpectedPrefix {
		t.Errorf("Expected filename prefix %q in file %q", expectedFilenamePrefix, fileName)
	}
	if !fileNameHasExpectedSuffix {
		t.Errorf("Expected filename suffix %q in file %q", expectedFilenameSuffix, fileName)
	}

	// Check file contents exist
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("CreateTmpFileWithTemplate failed to read file: %v", err)
	}
	actualLines := strings.Split(string(content), "\n")

	// Check parts w/ ordering
	expectedLines := []string{
		"---",
		"tags: []",
		"created: " + time.Now().Format("2006-01-02"),
		"review_interval: 0",
		"---",
		"",
		"# Title",
		"",
		"## Question",
		"",
		"## Answer",
		"",
	}
	for index, writtenLine := range expectedLines {
		actualLine := strings.TrimSpace(actualLines[index])
		writtenLine = strings.TrimSpace(writtenLine)

		if actualLine != writtenLine {
			t.Errorf("Line %d: expected %q, got %q", index, writtenLine, actualLine)
		}
	}
}

func TestCreateTmpFileWithCard(t *testing.T) {
	// Create Fake File
	tempDir, err := os.MkdirTemp("", "card-test")
	if err != nil {
		t.Fatalf("Cant cretae temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	fakeCardPath := filepath.Join(tempDir, "test-card.md")
	fakeCard := `---
tags: [test,go]
created: 2025-01-01
review_interval: 5
difficulty: 2.3
---

# Fake Title

## Question

Fake question

## Answer

Fake answer
`

	testCard := model.Card{
		ID: fakeCardPath,
	}

	err = os.WriteFile(fakeCardPath, []byte(fakeCard), 0644)
	if err != nil {
		t.Fatalf("Failed to write fake card %v", err)
	}

	expectedFilenamePrefix := "GoCard-tmp-"
	expectedFilenameSuffix := ".md"

	// Run function
	filePath, err := CreateTmpFileWithCard(testCard)
	defer os.Remove(filePath)
	if err != nil {
		t.Errorf("CreateTmpFileWithCard returned unexpected %v", err)
	}

	// Verify existence of file
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("CreateTmpFileWithCard didnt create file at %s", filePath)
	}

	// Filename checks
	fileName := filepath.Base(filePath)
	fileNameHasExpectedPrefix := strings.HasPrefix(fileName, expectedFilenamePrefix)
	fileNameHasExpectedSuffix := strings.HasSuffix(fileName, expectedFilenameSuffix)
	if !fileNameHasExpectedPrefix {
		t.Errorf("Expected filename prefix %q in file %q", expectedFilenamePrefix, fileName)
	}
	if !fileNameHasExpectedSuffix {
		t.Errorf("Expected filename suffix %q in file %q", expectedFilenameSuffix, fileName)
	}

	// Check file contents exist
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("CreateTmpFileWithCard failed to read file: %v", err)
	}

	if string(content) != fakeCard {
		t.Errorf("Temp file doesnt match card content")
	}
}
