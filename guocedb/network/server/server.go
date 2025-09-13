package server

import (
	"context"
	"net"
	"sync"

	"github.com/turtacn/guocedb/common/log"
)

// Handler is an interface for handling raw network connections.
type Handler interface {
	Serve(ctx context.Context, conn net.Conn)
}

// Server is a generic network server.
type Server struct {
	addr    string
	handler Handler
	wg      sync.WaitGroup
	lis     net.Listener
}

// NewServer creates a new generic network server.
func NewServer(addr string, handler Handler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

// Start begins listening for and handling connections.
// This is a blocking call.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.lis = lis
	log.Infof("Generic server listening on %s", s.addr)

	for {
		conn, err := lis.Accept()
		if err != nil {
			// Check if the listener was closed.
			select {
			case <-s.lis.(*net.TCPListener).File():
				return nil
			default:
				log.Errorf("Failed to accept connection: %v", err)
				return err
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer conn.Close()
			s.handler.Serve(context.Background(), conn)
		}()
	}
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	if s.lis != nil {
		s.lis.Close()
	}
	s.wg.Wait()
	log.Infof("Generic server on %s stopped", s.addr)
}

// Address returns the address the server is listening on.
func (s *Server) Address() string {
	if s.lis == nil {
		return ""
	}
	return s.lis.Addr().String()
}

// TODO: Add support for different protocols (e.g., Unix sockets).
// TODO: Implement connection limiting and timeout management.
// TODO: Add TLS support for secure communication.
