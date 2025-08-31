// Package mysql provides MySQL protocol handling for guocedb.
package mysql

import (
	"fmt"

	gms "github.com/dolthub/go-mysql-server/server"
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/compute/executor"
)

// Server is a MySQL protocol server for guocedb.
type Server struct {
	gmsServer *gms.Server
	connMgr   *ConnectionManager
}

// NewServer creates a new MySQL protocol server.
func NewServer(cfg *config.Config, engine *executor.Engine) (*Server, error) {
	// Create the connection manager
	connMgr := NewConnectionManager(cfg.Server.MaxConnections)

	// Create the handler
	handler := NewHandler(engine)

	// Create the underlying go-mysql-server
	server, err := gms.NewServer(
		gms.Config{
			Protocol: "tcp",
			Address:  fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Auth:     &gms.InsecureAuth{}, // Using insecure for now, a real impl would use protocol/mysql/auth.go
		},
		engine, // The GMS engine is now our executor
		handler,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Server{
		gmsServer: server,
		connMgr:   connMgr,
	}, nil
}

// Start begins listening for client connections.
func (s *Server) Start() {
	log.GetLogger().Infof("MySQL server started, listening on %s", s.gmsServer.Addr())
	go func() {
		err := s.gmsServer.Start()
		if err != nil {
			log.GetLogger().Fatalf("Failed to start MySQL server: %v", err)
		}
	}()
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	log.GetLogger().Info("Shutting down MySQL server...")
	s.connMgr.CloseAllConnections()
	return s.gmsServer.Close()
}
