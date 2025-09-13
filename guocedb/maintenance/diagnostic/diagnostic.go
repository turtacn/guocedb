package diagnostic

import (
	"io"
	"runtime/pprof"
	"time"
)

// Manager provides functions for system diagnostics and problem investigation.
type Manager struct {
	// Configuration for diagnostics, e.g., output directory.
	OutputDir string
}

// NewManager creates a new diagnostic manager.
func NewManager(outputDir string) *Manager {
	return &Manager{OutputDir: outputDir}
}

// GoroutineDump writes a dump of all current goroutine stacks to the provided writer.
// This is useful for debugging deadlocks.
func (m *Manager) GoroutineDump(w io.Writer) error {
	return pprof.Lookup("goroutine").WriteTo(w, 1)
}

// HeapDump writes a snapshot of the memory heap to the provided writer.
// This is useful for analyzing memory usage.
func (m *Manager) HeapDump(w io.Writer) error {
	return pprof.Lookup("heap").WriteTo(w, 1)
}

// CpuProfile collects a CPU profile for the specified duration and writes it to the writer.
// This is useful for identifying performance bottlenecks.
func (m *Manager) CpuProfile(w io.Writer, duration time.Duration) error {
	if err := pprof.StartCPUProfile(w); err != nil {
		return err
	}
	time.Sleep(duration)
	pprof.StopCPUProfile()
	return nil
}

// TODO: Create HTTP handlers to expose these diagnostic tools securely.
// TODO: Implement a function to generate a comprehensive diagnostic bundle (a zip file with multiple reports).
