package sal

import (
	"fmt"
	"sync"

	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/interfaces"
)

// StorageEngineFactory is a function that creates an instance of a storage engine.
type StorageEngineFactory func(cfg *config.StorageConfig) (interfaces.Storage, error)

var (
	factories = make(map[string]StorageEngineFactory)
	mu        sync.RWMutex
)

// RegisterStorageEngine registers a new storage engine factory.
// This function should be called by engine implementations in their init() function.
func RegisterStorageEngine(name string, factory StorageEngineFactory) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := factories[name]; exists {
		panic(fmt.Sprintf("storage engine factory named '%s' already registered", name))
	}
	factories[name] = factory
}

// NewStorageEngine creates a new storage engine instance based on the provided configuration.
// It looks up the appropriate factory from the registry and invokes it.
func NewStorageEngine(cfg *config.StorageConfig) (interfaces.Storage, error) {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := factories[cfg.Engine]
	if !ok {
		return nil, fmt.Errorf("unknown storage engine: %s", cfg.Engine)
	}
	return factory(cfg)
}
