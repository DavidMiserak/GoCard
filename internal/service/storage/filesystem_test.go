// internal/service/storage/filesystem_test.go
package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gocard-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	testCases := []struct {
		name        string
		rootDir     string
		shouldError bool
	}{
		{
			name:        "existing directory",
			rootDir:     tempDir,
			shouldError: false,
		},
		{
			name:        "non-existent directory that can be created",
			rootDir:     filepath.Join(tempDir, "new-dir"),
			shouldError: false,
		},
		{
			name:        "file instead of directory",
			rootDir:     filepath.Join(tempDir, "file.txt"),
			shouldError: true, // Our implementation returns an error if path exists but isn't a directory
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// If we're testing with a file, create it
			if tc.name == "file instead of directory" {
				file, err := os.Create(tc.rootDir)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				file.Close()
			}

			fs := NewFileSystemStorage()
			err := fs.Initialize(tc.rootDir)

			if tc.shouldError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tc.shouldError && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}

			// Verify the directory exists if we don't expect an error
			if !tc.shouldError {
				info, err := os.Stat(tc.rootDir)
				if err != nil {
					t.Errorf("failed to stat directory after Initialize: %v", err)
				} else if !info.IsDir() {
					t.Errorf("expected %s to be a directory", tc.rootDir)
				}
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
	fs := NewFileSystemStorage()

	// Test cases
	testCases := []struct {
		name              string
		content           string
		expectFrontmatter bool
		expectError       bool
	}{
		{
			name: "valid frontmatter",
			content: `---
title: Test Card
tags:
  - test
  - example
difficulty: 3
---
# Question

What is this test for?

---

# Answer

To test frontmatter parsing.
`,
			expectFrontmatter: true,
			expectError:       false,
		},
		{
			name: "no frontmatter",
			content: `# Question

What is this test for?

---

# Answer

To test frontmatter parsing.
`,
			expectFrontmatter: false,
			expectError:       false,
		},
		{
			name: "invalid YAML in frontmatter",
			content: `---
title: Test Card
tags: [broken,
difficulty: 3
---
# Question

What is this test for?
`,
			expectFrontmatter: false,
			expectError:       true,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frontmatter, markdown, err := fs.ParseFrontmatter([]byte(tc.content))

			if tc.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}

			if tc.expectFrontmatter && len(frontmatter) == 0 {
				t.Errorf("expected frontmatter but got none")
			}

			if !tc.expectFrontmatter && len(frontmatter) > 0 {
				t.Errorf("did not expect frontmatter but got: %v", frontmatter)
			}

			if !tc.expectError && len(markdown) == 0 {
				t.Errorf("expected markdown content but got none")
			}
		})
	}
}

func TestUpdateFrontmatter(t *testing.T) {
	fs := NewFileSystemStorage()

	// Create test content
	content := `---
title: Test Card
tags:
  - test
difficulty: 3
---
# Question

What is this test for?

---

# Answer

To test frontmatter updating.
`

	// Test updating various fields
	updates := map[string]interface{}{
		"last_reviewed":   "2025-03-25",
		"review_interval": 5,
		"difficulty":      4,
	}

	// Update frontmatter
	updatedContent, err := fs.UpdateFrontmatter([]byte(content), updates)
	if err != nil {
		t.Fatalf("failed to update frontmatter: %v", err)
	}

	// Parse the updated content to verify changes
	frontmatter, _, err := fs.ParseFrontmatter(updatedContent)
	if err != nil {
		t.Fatalf("failed to parse updated frontmatter: %v", err)
	}

	// Verify updates were applied
	for key, expectedValue := range updates {
		if value, ok := frontmatter[key]; !ok {
			t.Errorf("expected key %s to be present in frontmatter", key)
		} else if value != expectedValue {
			t.Errorf("expected %s to be %v, got %v", key, expectedValue, value)
		}
	}

	// Verify original fields were preserved
	if title, ok := frontmatter["title"].(string); !ok || title != "Test Card" {
		t.Errorf("expected title 'Test Card', got %v", frontmatter["title"])
	}
}
