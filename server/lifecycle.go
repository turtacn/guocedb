package server

import "sync"

// HookFunc is a lifecycle hook function.
type HookFunc func(s *Server)

// LifecycleHooks manages server lifecycle hooks.
type LifecycleHooks struct {
	preStart  []HookFunc
	postStart []HookFunc
	preStop   []HookFunc
	postStop  []HookFunc
	mu        sync.Mutex
}

// NewLifecycleHooks creates a new lifecycle hooks manager.
func NewLifecycleHooks() *LifecycleHooks {
	return &LifecycleHooks{
		preStart:  make([]HookFunc, 0),
		postStart: make([]HookFunc, 0),
		preStop:   make([]HookFunc, 0),
		postStop:  make([]HookFunc, 0),
	}
}

// OnPreStart registers a pre-start hook.
func (h *LifecycleHooks) OnPreStart(fn HookFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.preStart = append(h.preStart, fn)
}

// OnPostStart registers a post-start hook.
func (h *LifecycleHooks) OnPostStart(fn HookFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.postStart = append(h.postStart, fn)
}

// OnPreStop registers a pre-stop hook.
func (h *LifecycleHooks) OnPreStop(fn HookFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.preStop = append(h.preStop, fn)
}

// OnPostStop registers a post-stop hook.
func (h *LifecycleHooks) OnPostStop(fn HookFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.postStop = append(h.postStop, fn)
}

// RunPreStart executes pre-start hooks.
func (h *LifecycleHooks) RunPreStart(s *Server) {
	h.runHooks(h.preStart, s)
}

// RunPostStart executes post-start hooks.
func (h *LifecycleHooks) RunPostStart(s *Server) {
	h.runHooks(h.postStart, s)
}

// RunPreStop executes pre-stop hooks.
func (h *LifecycleHooks) RunPreStop(s *Server) {
	h.runHooks(h.preStop, s)
}

// RunPostStop executes post-stop hooks.
func (h *LifecycleHooks) RunPostStop(s *Server) {
	h.runHooks(h.postStop, s)
}

// runHooks executes a list of hooks, catching panics.
func (h *LifecycleHooks) runHooks(hooks []HookFunc, s *Server) {
	for _, fn := range hooks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Log hook panic without crashing
					if s.logger != nil {
						s.logger.Error("hook panic", "error", r)
					}
				}
			}()
			fn(s)
		}()
	}
}
