package health

import (
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecker_Healthy(t *testing.T) {
	// Create temporary Badger database
	opts := badger.DefaultOptions(t.TempDir())
	opts.Logger = nil
	db, err := badger.Open(opts)
	require.NoError(t, err)
	defer db.Close()
	
	checker := NewChecker(db)
	report := checker.Check()
	
	assert.Equal(t, StatusHealthy, report.Status)
	assert.Len(t, report.Checks, 3) // storage, memory, disk
	
	for _, check := range report.Checks {
		assert.Equal(t, StatusHealthy, check.Status)
		assert.Equal(t, "OK", check.Message)
		assert.Greater(t, check.Duration.Nanoseconds(), int64(0))
	}
}

func TestChecker_StorageFailed(t *testing.T) {
	// Use closed database to simulate storage failure
	opts := badger.DefaultOptions(t.TempDir())
	opts.Logger = nil
	db, _ := badger.Open(opts)
	db.Close() // Close immediately
	
	checker := NewChecker(db)
	report := checker.Check()
	
	assert.Equal(t, StatusUnhealthy, report.Status)
	
	// Find storage check result
	var storageCheck *CheckResult
	for i := range report.Checks {
		if report.Checks[i].Name == "storage" {
			storageCheck = &report.Checks[i]
			break
		}
	}
	
	require.NotNil(t, storageCheck)
	assert.Equal(t, StatusUnhealthy, storageCheck.Status)
	assert.Contains(t, storageCheck.Message, "storage read/write test failed")
}

func TestChecker_Degraded(t *testing.T) {
	opts := badger.DefaultOptions(t.TempDir())
	opts.Logger = nil
	db, _ := badger.Open(opts)
	defer db.Close()
	
	// Create checker with custom checks
	checker := &Checker{
		db: db,
		checks: []Check{
			{Name: "storage", Fn: func() error { return nil }, Critical: true},
			{Name: "optional", Fn: func() error { return assert.AnError }, Critical: false},
		},
	}
	
	report := checker.Check()
	assert.Equal(t, StatusDegraded, report.Status)
	
	// Storage should be healthy
	var storageCheck *CheckResult
	for i := range report.Checks {
		if report.Checks[i].Name == "storage" {
			storageCheck = &report.Checks[i]
			break
		}
	}
	require.NotNil(t, storageCheck)
	assert.Equal(t, StatusHealthy, storageCheck.Status)
	
	// Optional check should be unhealthy
	var optionalCheck *CheckResult
	for i := range report.Checks {
		if report.Checks[i].Name == "optional" {
			optionalCheck = &report.Checks[i]
			break
		}
	}
	require.NotNil(t, optionalCheck)
	assert.Equal(t, StatusUnhealthy, optionalCheck.Status)
}

func TestChecker_AddCheck(t *testing.T) {
	opts := badger.DefaultOptions(t.TempDir())
	opts.Logger = nil
	db, _ := badger.Open(opts)
	defer db.Close()
	
	checker := NewChecker(db)
	initialChecks := len(checker.checks)
	
	// Add custom check
	checker.AddCheck("custom", func() error { return nil }, false)
	
	assert.Len(t, checker.checks, initialChecks+1)
	
	report := checker.Check()
	assert.Len(t, report.Checks, initialChecks+1)
	
	// Find custom check
	var customCheck *CheckResult
	for i := range report.Checks {
		if report.Checks[i].Name == "custom" {
			customCheck = &report.Checks[i]
			break
		}
	}
	
	require.NotNil(t, customCheck)
	assert.Equal(t, StatusHealthy, customCheck.Status)
}

func TestChecker_NilDatabase(t *testing.T) {
	checker := NewChecker(nil)
	report := checker.Check()
	
	assert.Equal(t, StatusUnhealthy, report.Status)
	
	// Storage check should fail
	var storageCheck *CheckResult
	for i := range report.Checks {
		if report.Checks[i].Name == "storage" {
			storageCheck = &report.Checks[i]
			break
		}
	}
	
	require.NotNil(t, storageCheck)
	assert.Equal(t, StatusUnhealthy, storageCheck.Status)
	assert.Contains(t, storageCheck.Message, "database is nil")
}