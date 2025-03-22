// Package io provides file system operations for the GoCard storage system.
package io

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel defines the severity of a log message
type LogLevel int

const (
	// DEBUG level for detailed diagnostic information
	DEBUG LogLevel = iota
	// INFO level for general operational information
	INFO
	// WARN level for potentially harmful situations
	WARN
	// ERROR level for error events that might still allow the app to continue
	ERROR
)

// Logger provides a simple logging facility
type Logger struct {
	mu        sync.Mutex
	out       io.Writer
	level     LogLevel
	enabled   bool
	timestamp bool
}

var (
	// DefaultLogger is the default logger instance
	DefaultLogger = NewLogger(os.Stdout, INFO)
	// Global flag to enable/disable all logging
	loggingEnabled = true
)

// NewLogger creates a new logger with the specified output and level
func NewLogger(out io.Writer, level LogLevel) *Logger {
	return &Logger{
		out:       out,
		level:     level,
		enabled:   true,
		timestamp: true,
	}
}

// SetLevel sets the minimum log level to output
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetEnabled enables or disables logging for this logger
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = out
}

// SetTimestamp enables or disables timestamps in log messages
func (l *Logger) SetTimestamp(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = enabled
}

// log formats and writes a log message
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if !loggingEnabled || !l.enabled || level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	var prefix string
	switch level {
	case DEBUG:
		prefix = "DEBUG"
	case INFO:
		prefix = "INFO"
	case WARN:
		prefix = "WARN"
	case ERROR:
		prefix = "ERROR"
	}

	var msg string
	if l.timestamp {
		msg = fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), prefix, fmt.Sprintf(format, args...))
	} else {
		msg = fmt.Sprintf("[%s] %s\n", prefix, fmt.Sprintf(format, args...))
	}

	fmt.Fprint(l.out, msg)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// EnableLogging enables or disables all logging globally
func EnableLogging(enabled bool) {
	loggingEnabled = enabled
}

// SetGlobalLevel sets the minimum log level for the default logger
func SetGlobalLevel(level LogLevel) {
	DefaultLogger.SetLevel(level)
}
