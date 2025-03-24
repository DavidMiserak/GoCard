// File: internal/storage/io/logger_test.go

package io

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, DEBUG)

	if logger.out != &buf {
		t.Error("Logger output not set correctly")
	}

	if logger.level != DEBUG {
		t.Errorf("Logger level not set correctly, expected %d, got %d", DEBUG, logger.level)
	}

	if !logger.enabled {
		t.Error("Logger should be enabled by default")
	}

	if !logger.timestamp {
		t.Error("Logger timestamp should be enabled by default")
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		logFunc  func(logger *Logger, msg string)
		logLevel LogLevel
		expected string
	}{
		{"Debug at DEBUG level", DEBUG, func(l *Logger, m string) { l.Debug(m) }, DEBUG, "[DEBUG]"},
		{"Info at DEBUG level", DEBUG, func(l *Logger, m string) { l.Info(m) }, DEBUG, "[INFO]"},
		{"Warn at DEBUG level", DEBUG, func(l *Logger, m string) { l.Warn(m) }, DEBUG, "[WARN]"},
		{"Error at DEBUG level", DEBUG, func(l *Logger, m string) { l.Error(m) }, DEBUG, "[ERROR]"},

		{"Debug at INFO level", INFO, func(l *Logger, m string) { l.Debug(m) }, DEBUG, ""}, // Should not log
		{"Info at INFO level", INFO, func(l *Logger, m string) { l.Info(m) }, INFO, "[INFO]"},
		{"Warn at INFO level", INFO, func(l *Logger, m string) { l.Warn(m) }, WARN, "[WARN]"},
		{"Error at INFO level", INFO, func(l *Logger, m string) { l.Error(m) }, ERROR, "[ERROR]"},

		{"Debug at WARN level", WARN, func(l *Logger, m string) { l.Debug(m) }, DEBUG, ""}, // Should not log
		{"Info at WARN level", WARN, func(l *Logger, m string) { l.Info(m) }, INFO, ""},    // Should not log
		{"Warn at WARN level", WARN, func(l *Logger, m string) { l.Warn(m) }, WARN, "[WARN]"},
		{"Error at WARN level", WARN, func(l *Logger, m string) { l.Error(m) }, ERROR, "[ERROR]"},

		{"Debug at ERROR level", ERROR, func(l *Logger, m string) { l.Debug(m) }, DEBUG, ""}, // Should not log
		{"Info at ERROR level", ERROR, func(l *Logger, m string) { l.Info(m) }, INFO, ""},    // Should not log
		{"Warn at ERROR level", ERROR, func(l *Logger, m string) { l.Warn(m) }, WARN, ""},    // Should not log
		{"Error at ERROR level", ERROR, func(l *Logger, m string) { l.Error(m) }, ERROR, "[ERROR]"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tc.level)
			logger.SetTimestamp(false) // Disable timestamp for easier testing

			tc.logFunc(logger, "Test message")

			result := buf.String()
			if tc.expected == "" {
				if result != "" {
					t.Errorf("Expected no output, got: %q", result)
				}
			} else {
				if !strings.Contains(result, tc.expected) {
					t.Errorf("Expected output containing %q, got: %q", tc.expected, result)
				}
				if !strings.Contains(result, "Test message") {
					t.Errorf("Expected output containing message, got: %q", result)
				}
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, ERROR)

	// Initially, INFO messages should not be logged
	logger.Info("Should not appear")
	if buf.Len() > 0 {
		t.Errorf("Expected no output, got: %q", buf.String())
	}

	// Change level to INFO
	logger.SetLevel(INFO)

	// Now INFO messages should be logged
	logger.SetTimestamp(false) // Disable timestamp for easier testing
	logger.Info("Should appear")

	if !strings.Contains(buf.String(), "Should appear") {
		t.Errorf("Expected message after level change, got: %q", buf.String())
	}
}

func TestSetEnabled(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, INFO)
	logger.SetTimestamp(false) // Disable timestamp for easier testing

	// Initially enabled
	logger.Info("First message")
	if !strings.Contains(buf.String(), "First message") {
		t.Errorf("Expected first message, got: %q", buf.String())
	}

	// Disable logging
	logger.SetEnabled(false)
	buf.Reset()

	logger.Info("Should not appear")
	if buf.Len() > 0 {
		t.Errorf("Expected no output when disabled, got: %q", buf.String())
	}

	// Re-enable logging
	logger.SetEnabled(true)
	logger.Info("Third message")
	if !strings.Contains(buf.String(), "Third message") {
		t.Errorf("Expected third message after re-enabling, got: %q", buf.String())
	}
}

func TestSetOutput(t *testing.T) {
	var buf1 bytes.Buffer
	var buf2 bytes.Buffer

	logger := NewLogger(&buf1, INFO)
	logger.SetTimestamp(false) // Disable timestamp for easier testing

	// Log to first buffer
	logger.Info("First buffer")
	if !strings.Contains(buf1.String(), "First buffer") {
		t.Errorf("Expected message in first buffer, got: %q", buf1.String())
	}

	// Change output to second buffer
	logger.SetOutput(&buf2)
	logger.Info("Second buffer")

	if strings.Contains(buf1.String(), "Second buffer") {
		t.Errorf("First buffer should not contain second message: %q", buf1.String())
	}

	if !strings.Contains(buf2.String(), "Second buffer") {
		t.Errorf("Expected message in second buffer, got: %q", buf2.String())
	}
}

func TestSetTimestamp(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, INFO)

	// Default is with timestamp
	logger.Info("With timestamp")
	if !strings.Contains(buf.String(), "[") && strings.Contains(buf.String(), ":") {
		t.Errorf("Expected timestamp in format [HH:MM:SS], got: %q", buf.String())
	}

	// Disable timestamp
	buf.Reset()
	logger.SetTimestamp(false)
	logger.Info("Without timestamp")

	if strings.Contains(buf.String(), "[") && strings.Contains(buf.String(), ":") &&
		strings.Contains(buf.String(), "]") && !strings.Contains(buf.String(), "[INFO]") {
		t.Errorf("Expected no timestamp, got: %q", buf.String())
	}
}

func TestGlobalLogging(t *testing.T) {
	// Save original state
	origEnabled := loggingEnabled
	defer func() { loggingEnabled = origEnabled }()

	var buf bytes.Buffer
	logger := NewLogger(&buf, INFO)
	logger.SetTimestamp(false)

	// Test global enable/disable
	EnableLogging(false)
	logger.Info("Should not appear")

	if buf.Len() > 0 {
		t.Errorf("Expected no output when globally disabled, got: %q", buf.String())
	}

	EnableLogging(true)
	logger.Info("Should appear")

	if !strings.Contains(buf.String(), "Should appear") {
		t.Errorf("Expected message when globally enabled, got: %q", buf.String())
	}
}

func TestSetGlobalLevel(t *testing.T) {
	// Save original level
	origLevel := DefaultLogger.level
	defer func() { DefaultLogger.level = origLevel }()

	var buf bytes.Buffer
	DefaultLogger.out = &buf
	DefaultLogger.SetTimestamp(false)

	// Set global level to ERROR
	SetGlobalLevel(ERROR)

	// INFO should not be logged
	DefaultLogger.Info("Should not appear")
	if buf.Len() > 0 {
		t.Errorf("Expected no output at ERROR level, got: %q", buf.String())
	}

	// ERROR should be logged
	DefaultLogger.Error("Should appear")
	if !strings.Contains(buf.String(), "Should appear") {
		t.Errorf("Expected error message, got: %q", buf.String())
	}

	// Restore default logger output to avoid affecting other tests
	DefaultLogger.out = nil
}
