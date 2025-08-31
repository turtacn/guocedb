// Package mesh provides service mesh integration capabilities for guocedb.
package mesh

import "github.com/turtacn/guocedb/common/errors"

// Service represents a service discovered via the service mesh.
type Service struct {
	Name      string
	Address   string
	Port      int
	Tags      []string
	Namespace string
}

// Discovery is the interface for a service discovery client.
type Discovery interface {
	// GetService returns the details for a specific service.
	GetService(name string) (*Service, error)
	// ListServices returns all services with a given tag.
	ListServices(tag string) ([]*Service, error)
	// Register registers the current node as a service.
	Register() error
	// Deregister removes the current node from service discovery.
	Deregister() error
}

// TrafficManager is the interface for managing traffic policies.
type TrafficManager interface {
	// ApplyCanary sets up a canary release policy.
	ApplyCanary(serviceName string, percentage int) error
	// ApplyFailover configures automatic failover.
	ApplyFailover(serviceName string) error
	// InjectFault injects faults for chaos testing.
	InjectFault(serviceName string, faultType string, percentage int) error
}

// IstioAdapter is a placeholder for an Istio integration adapter.
type IstioAdapter struct{}

func (i *IstioAdapter) GetService(name string) (*Service, error) {
	return nil, errors.ErrNotImplemented
}

func (i *IstioAdapter) ListServices(tag string) ([]*Service, error) {
	return nil, errors.ErrNotImplemented
}

func (i *IstioAdapter) Register() error {
	return errors.ErrNotImplemented
}

func (i *IstioAdapter) Deregister() error {
	return errors.ErrNotImplemented
}

// Enforce interface compliance
var _ Discovery = (*IstioAdapter)(nil)
