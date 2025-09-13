package mysql

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/maintenance/metrics"
)

var nextConnectionID uint32

// SessionBuilder creates a new sql.Session for each client connection.
func SessionBuilder(addr string, user, password string) (sql.Session, error) {
	client := sql.Client{Address: addr, User: user, Capabilities: 0}

	// Increment connection metrics
	metrics.ActiveConnections.Inc()

	id := atomic.AddUint32(&nextConnectionID, 1)

	log.WithFields(map[string]interface{}{
		"connectionID": id,
		"clientAddr":   addr,
		"user":         user,
	}).Infof("New client connection")

	// Create a new session with our custom builder logic.
	// We can initialize session variables or other state here.
	return sql.NewSession(addr, client, id), nil
}

// ConnectionManager can be used to track and manage active connections.
// This is a placeholder for more advanced connection management features.
type ConnectionManager struct {
	mu          sync.Mutex
	connections map[uint32]sql.Session
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[uint32]sql.Session),
	}
}

func (m *ConnectionManager) Add(sess sql.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[sess.ID()] = sess
}

func (m *ConnectionManager) Remove(sess sql.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, sess.ID())
	metrics.ActiveConnections.Dec() // Decrement connection metrics
}

func (m *ConnectionManager) Kill(connID uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	sess, ok := m.connections[connID]
	if !ok {
		return fmt.Errorf("connection %d not found", connID)
	}
	// This kills the query, not necessarily the underlying TCP connection.
	sess.Kill()
	return nil
}

// TODO: Implement more robust session management, including session timeouts.
// TODO: Expose connection information through a management API.
