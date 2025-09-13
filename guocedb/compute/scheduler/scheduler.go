package scheduler

import (
	"github.com/dolthub/go-mysql-server/sql"
)

// Task represents a unit of work that can be scheduled for execution.
// In a distributed system, this could be a fragment of a query plan.
type Task struct {
	ID   string
	Node sql.Node
	// Add other metadata like priority, resource requirements, etc.
}

// Result holds the outcome of a task execution.
type Result struct {
	TaskID  string
	RowIter sql.RowIter
	Err     error
}

// Scheduler is the interface for a query scheduler.
// It is responsible for distributing tasks and collecting results.
type Scheduler interface {
	// Schedule submits a task for execution and returns a channel for the result.
	Schedule(ctx *sql.Context, task *Task) (<-chan *Result, error)
	// Close shuts down the scheduler.
	Close() error
}

// LocalScheduler is a simple scheduler that executes tasks locally and synchronously.
type LocalScheduler struct{}

// NewLocalScheduler creates a new local scheduler.
func NewLocalScheduler() Scheduler {
	return &LocalScheduler{}
}

// Schedule executes the task immediately in the same goroutine.
func (s *LocalScheduler) Schedule(ctx *sql.Context, task *Task) (<-chan *Result, error) {
	resultChan := make(chan *Result, 1)
	go func() {
		defer close(resultChan)
		iter, err := task.Node.RowIter(ctx, nil)
		resultChan <- &Result{
			TaskID:  task.ID,
			RowIter: iter,
			Err:     err,
		}
	}()
	return resultChan, nil
}

// Close does nothing for the local scheduler.
func (s *LocalScheduler) Close() error {
	return nil
}

// TODO: Implement a distributed scheduler that can send tasks to remote worker nodes.
// This would involve:
// - A worker discovery mechanism (e.g., using a service mesh or consensus store).
// - A task serialization format (e.g., Protobuf).
// - Network communication for task distribution and result collection (e.g., gRPC).
// - Fault tolerance mechanisms for handling worker failures.
