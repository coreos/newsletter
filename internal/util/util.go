package util

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger wraps standard logging
type Logger struct {
	Verbose bool
}

// NewLogger creates a new logger
func NewLogger(verbose bool) *Logger {
	return &Logger{Verbose: verbose}
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	fmt.Printf("[INFO] %s\n", msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	fmt.Printf("[WARN] %s\n", msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
}

// Debug logs a debug message if verbose is enabled
func (l *Logger) Debug(msg string) {
	if l.Verbose {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

// Fatal logs an error and exits
func (l *Logger) Fatal(msg string) {
	log.Fatal("[FATAL] " + msg)
}


// ParseDate parses an ISO8601 date string
func ParseDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// Handle Z suffix (replace with +00:00)
	if len(dateStr) > 0 && dateStr[len(dateStr)-1] == 'Z' {
		dateStr = dateStr[:len(dateStr)-1] + "+00:00"
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Contains checks if a string slice contains a string
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// TimePtr returns a pointer to a time
func TimePtr(t time.Time) *time.Time {
	return &t
}
