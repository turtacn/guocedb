package utils

import (
	"os"
	"time"

	"github.com/turtacn/guocedb/common/log"
)

// This package provides internal helper functions that are not part of the public API
// but are used across different components of the guocedb project.

// FileExists checks if a file or directory exists at the given path.
// It returns true if the path exists and is accessible, false otherwise.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	// os.IsNotExist is the most reliable way to check for non-existence.
	// A nil error means it exists. Any other error (e.g., permission denied)
	// is treated as if it doesn't exist for simplicity here.
	return !os.IsNotExist(err)
}

// TimeTrack is a utility function to measure and log the execution time of a function.
// Usage: defer TimeTrack(time.Now(), "myFunction")
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	// Use the structured logger for consistent logging format.
	log.WithField("duration", elapsed).Debugf("Execution of '%s' took %s", name, elapsed)
}

// TODO: Add more utility functions as needed:
// - Concurrency helpers (e.g., a WaitGroup that collects errors).
// - Type conversion helpers with robust error checking.
// - Resource cleanup functions (e.g., a multi-closer).
