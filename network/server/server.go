// Package server provides a generic framework for network services.
package server

import (
	"net"

	"github.com/turtacn/guocedb/common/log"
)

// ManagedServer is an interface for a network service that can be started and stopped.
type ManagedServer interface {
	// Start begins the server's execution.
	Start()
	// Close gracefully shuts down the server.
	Close() error
	// Addr returns the network address the server is listening on.
	Addr() string
}

// Manager holds and manages multiple network services.
type Manager struct {
	servers []ManagedServer
}

// NewManager creates a new server manager.
func NewManager() *Manager {
	return &Manager{
		servers: make([]ManagedServer, 0),
	}
}

// Register adds a new server to be managed.
func (m *Manager) Register(s ManagedServer) {
	m.servers = append(m.servers, s)
}

// StartAll starts all registered servers.
func (m *Manager) StartAll() {
	log.GetLogger().Info("Starting all registered network services...")
	for _, s := range m.servers {
		s.Start()
	}
}

// CloseAll stops all registered servers.
func (m *Manager) CloseAll() {
	log.GetLogger().Info("Stopping all registered network services...")
	for _, s := range m.servers {
		if err := s.Close(); err != nil {
			log.GetLogger().Errorf("Failed to close server %s: %v", s.Addr(), err)
		}
	}
}

// TCPServer is a basic TCP server placeholder.
type TCPServer struct {
	listener net.Listener
}

// NewTCPServer creates a new TCP server.
func NewTCPServer(addr string) (*TCPServer, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPServer{listener: l}, nil
}

func (s *TCPServer) Start() {
	// Placeholder for accept loop
}

func (s *TCPServer) Close() error {
	return s.listener.Close()
}

func (s *TCPServer) Addr() string {
	return s.listener.Addr().String()
}
