// Package mysql provides MySQL protocol handling for guocedb.
package mysql

import (
	"sync"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/turtacn/guocedb/maintenance/metrics"
)

// ConnectionManager manages all active client connections.
type ConnectionManager struct {
	mu          sync.Mutex
	connections map[uint32]*server.Conn
	maxConns    int
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(maxConns int) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[uint32]*server.Conn),
		maxConns:    maxConns,
	}
}

// AddConnection adds a new connection to the manager.
func (cm *ConnectionManager) AddConnection(conn *server.Conn) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.connections) >= cm.maxConns {
		return false // Connection limit reached
	}

	cm.connections[conn.ConnectionID] = conn
	metrics.ActiveConnections.Inc()
	return true
}

// RemoveConnection removes a connection from the manager.
func (cm *ConnectionManager) RemoveConnection(conn *server.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.connections, conn.ConnectionID)
	metrics.ActiveConnections.Dec()
}

// CloseAllConnections gracefully closes all active connections.
func (cm *ConnectionManager) CloseAllConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, conn := range cm.connections {
		conn.Close()
	}
}

// newConnectionCallback is a function used by the GMS server to notify of new connections.
func (cm *ConnectionManager) newConnectionCallback(conn *server.Conn) {
	if !cm.AddConnection(conn) {
		// Refuse connection
		conn.Close()
	}
}

// closeConnectionCallback is a function used by the GMS server to notify of closed connections.
func (cm *ConnectionManager) closeConnectionCallback(conn *server.Conn) {
	cm.RemoveConnection(conn)
}
