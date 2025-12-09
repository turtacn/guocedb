package server

import (
	"log/slog"

	"github.com/turtacn/guocedb/config"
)

// Option is a function that modifies server configuration.
type Option func(*Server)

// WithLogger sets a custom logger for the server.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithHook adds a lifecycle hook to the server.
func WithHook(phase string, fn HookFunc) Option {
	return func(s *Server) {
		switch phase {
		case "preStart":
			s.hooks.OnPreStart(fn)
		case "postStart":
			s.hooks.OnPostStart(fn)
		case "preStop":
			s.hooks.OnPreStop(fn)
		case "postStop":
			s.hooks.OnPostStop(fn)
		}
	}
}

// NewWithOptions creates a server with additional options.
func NewWithOptions(cfg *config.Config, opts ...Option) (*Server, error) {
	srv, err := New(cfg)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}
