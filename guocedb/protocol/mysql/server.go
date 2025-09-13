package mysql

import (
	"fmt"
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/compute/executor"
)

// Server is a wrapper around the go-mysql-server.
// It manages the lifecycle of the MySQL protocol listener.
type Server struct {
	config server.Config
	gms    *server.Server
}

// NewServer creates a new MySQL protocol server.
func NewServer(host string, port int, engine *executor.Engine) (*Server, error) {
	// Create our custom handler and authentication provider.
	handler := NewHandler(engine)
	authProvider := NewAuthProvider()

	// Configure the go-mysql-server instance.
	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", host, port),
		Auth:     authProvider,
		// Add other options like TLS, connection read timeouts, etc.
		ConnReadTimeout: 30 * time.Second,
	}

	s, err := server.NewServer(config, nil, handler)
	if err != nil {
		return nil, err
	}

	return &Server{
		config: config,
		gms:    s,
	}, nil
}

// Start begins listening for client connections.
// This is a blocking call, so it should be run in a goroutine.
func (s *Server) Start() error {
	log.Infof("MySQL server listening on %s", s.config.Address)
	return s.gms.Start()
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() error {
	log.Infof("Shutting down MySQL server...")
	return s.gms.Close()
}

// Address returns the address the server is listening on.
func (s *Server) Address() string {
	return s.gms.Listener.Addr().String()
}
