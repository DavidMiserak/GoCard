// This file should be created as internal/service/storage/date_parsing_test.go

package storage

import (
	"fmt"
	"testing"
	"time"
)

func TestParseYAMLDate(t *testing.T) {
	// Setup test cases with various date formats
	testCases := []struct {
		name        string
		dateStr     string
		expected    time.Time
		expectError bool
	}{
		{
			name:        "standard ISO date",
			dateStr:     "2023-05-15",
			expected:    time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "ISO datetime",
			dateStr:     "2023-05-15T14:30:45Z",
			expected:    time.Date(2023, 5, 15, 14, 30, 45, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "common datetime format",
			dateStr:     "2023-05-15 14:30:45",
			expected:    time.Date(2023, 5, 15, 14, 30, 45, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "RFC3339 format",
			dateStr:     "2023-05-15T14:30:45+02:00",
			expected:    time.Date(2023, 5, 15, 14, 30, 45, 0, time.FixedZone("", 2*60*60)),
			expectError: false,
		},
		{
			name:        "human readable date",
			dateStr:     "January 15, 2023",
			expected:    time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "short month format",
			dateStr:     "Jan 15, 2023",
			expected:    time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "invalid date",
			dateStr:     "not-a-date",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "empty string",
			dateStr:     "",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "partial date (month and year only)",
			dateStr:     "May 2023",
			expected:    time.Time{},
			expectError: true,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseYAMLDate(tc.dateStr)

			// Check error
			if tc.expectError && err == nil {
				t.Error("expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("did not expect an error but got: %v", err)
			}

			// Skip further checks if we expected an error
			if tc.expectError {
				return
			}

			// For valid dates, check if the result matches the expected value
			// Note: Depending on the time parsing implementation, time zones might differ
			// so we compare year, month, day, hour, minute, second
			expectedYear, expectedMonth, expectedDay := tc.expected.Date()
			expectedHour, expectedMin, expectedSec := tc.expected.Clock()

			resultYear, resultMonth, resultDay := result.Date()
			resultHour, resultMin, resultSec := result.Clock()

			if resultYear != expectedYear || resultMonth != expectedMonth || resultDay != expectedDay {
				t.Errorf("expected date %v, got %v", tc.expected.Format("2006-01-02"), result.Format("2006-01-02"))
			}

			// For time-inclusive formats, check the time components
			if tc.dateStr != "2023-05-15" && tc.dateStr != "January 15, 2023" && tc.dateStr != "Jan 15, 2023" {
				if resultHour != expectedHour || resultMin != expectedMin || resultSec != expectedSec {
					t.Errorf("expected time %v, got %v", tc.expected.Format("15:04:05"), result.Format("15:04:05"))
				}
			}
		})
	}
}

// This function should be added to enhance TestLoadCard in filesystem_test.go
func TestLoadCardWithVariousDateFormats(t *testing.T) {
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Create test cases with different date formats
	testCases := []struct {
		name          string
		frontmatter   string
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{
			name: "ISO date format",
			frontmatter: `---
title: ISO Date Test
last_reviewed: 2023-05-15
review_interval: 7
---
`,
			expectedYear:  2023,
			expectedMonth: time.May,
			expectedDay:   15,
		},
		{
			name: "ISO datetime format",
			frontmatter: `---
title: ISO Datetime Test
last_reviewed: 2023-05-15T14:30:45Z
review_interval: 7
---
`,
			expectedYear:  2023,
			expectedMonth: time.May,
			expectedDay:   15,
		},
		{
			name: "common datetime format",
			frontmatter: `---
title: Common Datetime Test
last_reviewed: 2023-05-15 14:30:45
review_interval: 7
---
`,
			expectedYear:  2023,
			expectedMonth: time.May,
			expectedDay:   15,
		},
		{
			name: "human readable date",
			frontmatter: `---
title: Human Readable Date Test
last_reviewed: January 15, 2023
review_interval: 7
---
`,
			expectedYear:  2023,
			expectedMonth: time.January,
			expectedDay:   15,
		},
		{
			name: "short month format",
			frontmatter: `---
title: Short Month Test
last_reviewed: Jan 15, 2023
review_interval: 7
---
`,
			expectedYear:  2023,
			expectedMonth: time.January,
			expectedDay:   15,
		},
	}

	// Test each date format
	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create card content
			content := tc.frontmatter + "# Question\n\nTest?\n\n---\n\nAnswer.\n"

			// Create the card file
			filename := fmt.Sprintf("date-format-test-%d.md", i)
			cardPath, err := createSampleCardFile(tempDir, filename, content)
			if err != nil {
				t.Fatalf("failed to create sample card file: %v", err)
			}

			// Load the card
			card, err := fs.LoadCard(cardPath)
			if err != nil {
				t.Fatalf("LoadCard() error = %v", err)
			}

			// Verify the date was parsed correctly
			year, month, day := card.LastReviewed.Date()

			if year != tc.expectedYear {
				t.Errorf("expected year %d, got %d", tc.expectedYear, year)
			}
			if month != tc.expectedMonth {
				t.Errorf("expected month %s, got %s", tc.expectedMonth, month)
			}
			if day != tc.expectedDay {
				t.Errorf("expected day %d, got %d", tc.expectedDay, day)
			}
		})
	}
}

// This function tests various data types in YAML frontmatter
func TestLoadCardWithVariousDateTypes(t *testing.T) {
	fs, tempDir, cleanup := setupFileSystemTest(t)
	defer cleanup()

	// Create test cases with different frontmatter formats
	testCases := []struct {
		name           string
		frontmatter    string
		expectNonZero  bool
		expectInterval int
	}{
		{
			name: "string date",
			frontmatter: `---
title: String Date Test
last_reviewed: 2023-05-15
review_interval: 7
---`,
			expectNonZero:  true,
			expectInterval: 7,
		},
		{
			name: "null/nil date",
			frontmatter: `---
title: Null Date Test
last_reviewed: null
review_interval: 7
---`,
			expectNonZero:  false,
			expectInterval: 7,
		},
		{
			name: "int as review interval",
			frontmatter: `---
title: Int Interval Test
last_reviewed: 2023-05-15
review_interval: 10
---`,
			expectNonZero:  true,
			expectInterval: 10,
		},
		{
			name: "float as review interval",
			frontmatter: `---
title: Float Interval Test
last_reviewed: 2023-05-15
review_interval: 10.5
---`,
			expectNonZero:  true,
			expectInterval: 10, // Should be truncated to int
		},
		{
			name: "string as review interval",
			frontmatter: `---
title: String Interval Test
last_reviewed: 2023-05-15
review_interval: "15"
---`,
			expectNonZero:  true,
			expectInterval: 15, // Should be parsed from string
		},
		{
			name: "unix timestamp as date",
			frontmatter: `---
title: Timestamp Date Test
last_reviewed: 1684108800  # 2023-05-15 00:00:00 UTC
review_interval: 7
---`,
			expectNonZero:  true,
			expectInterval: 7,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create card content with complete structure
			content := tc.frontmatter + `
# Question

Test question?

---

Test answer.
`

			// Create the card file
			filename := fmt.Sprintf("date-type-test-%d.md", i)
			cardPath, err := createSampleCardFile(tempDir, filename, content)
			if err != nil {
				t.Fatalf("failed to create sample card file: %v", err)
			}

			// Load the card
			card, err := fs.LoadCard(cardPath)
			if err != nil {
				t.Fatalf("LoadCard() error = %v", err)
			}

			// Check if LastReviewed is correctly set
			if tc.expectNonZero {
				if card.LastReviewed.IsZero() {
					t.Error("expected LastReviewed to be non-zero")
				}
			} else {
				if !card.LastReviewed.IsZero() {
					t.Error("expected LastReviewed to be zero")
				}
			}

			// Check if ReviewInterval is correctly set
			if card.ReviewInterval != tc.expectInterval {
				t.Errorf("expected ReviewInterval to be %d, got %d",
					tc.expectInterval, card.ReviewInterval)
			}
		})
	}
}
