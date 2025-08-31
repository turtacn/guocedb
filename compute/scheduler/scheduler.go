// Package scheduler provides distributed task scheduling for guocedb.
package scheduler

import (
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/compute/plan"
)

// Task represents a unit of work that can be scheduled on a node.
type Task struct {
	ID   string
	Plan plan.Node // The sub-plan to be executed by the node.
}

// Node represents a worker node in the distributed cluster.
type Node struct {
	ID      string
	Address string
	// Other metadata like capacity, load, etc.
}

// Scheduler is the interface for a distributed task scheduler.
type Scheduler interface {
	// Schedule takes a query plan and distributes it as tasks to worker nodes.
	Schedule(p plan.Node) (chan Result, error)
	// AddNode adds a new worker node to the scheduler.
	AddNode(n Node) error
	// RemoveNode removes a worker node from the scheduler.
	RemoveNode(nodeID string) error
}

// Result is the result of a task execution from a worker node.
type Result struct {
	TaskID string
	// In a real implementation, this would be a stream of rows or a batch.
	Data   []byte
	Error  error
}

// DefaultScheduler is a basic scheduler for a single-node setup.
// It doesn't actually distribute any tasks.
type DefaultScheduler struct {
	// In a distributed setup, this would hold the state of the cluster.
}

// NewDefaultScheduler creates a new default scheduler.
func NewDefaultScheduler() *DefaultScheduler {
	return &DefaultScheduler{}
}

// Schedule "schedules" a plan by simply indicating it's not implemented for distribution.
func (s *DefaultScheduler) Schedule(p plan.Node) (chan Result, error) {
	// In a single-node setup, the plan is executed directly, not scheduled.
	return nil, errors.ErrNotImplemented
}

func (s *DefaultScheduler) AddNode(n Node) error {
	return errors.ErrNotImplemented
}

func (s *DefaultScheduler) RemoveNode(nodeID string) error {
	return errors.ErrNotImplemented
}

// Enforce interface compliance
var _ Scheduler = (*DefaultScheduler)(nil)
