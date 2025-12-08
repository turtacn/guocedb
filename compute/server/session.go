package server

import (
	"context"
	"sync"

	"github.com/turtacn/guocedb/compute/sql"
)

// Session represents a database session with connection state
type Session struct {
	id          uint32
	currentDB   string
	user        string
	client      string
	vars        map[string]interface{}
	transaction sql.Transaction
	autoCommit  bool
	mu          sync.RWMutex
}

// NewSession creates a new session with the given parameters
func NewSession(id uint32, user, client string) *Session {
	return &Session{
		id:         id,
		user:       user,
		client:     client,
		vars:       make(map[string]interface{}),
		autoCommit: true, // Default to autocommit mode
	}
}

// ID returns the session ID
func (s *Session) ID() uint32 {
	return s.id
}

// SetCurrentDB sets the current database for this session
func (s *Session) SetCurrentDB(db string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentDB = db
}

// GetCurrentDB returns the current database for this session
func (s *Session) GetCurrentDB() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentDB
}

// Context creates a new SQL context with session information
func (s *Session) Context(baseCtx context.Context) *sql.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	ctx := sql.NewContext(baseCtx)
	if s.currentDB != "" {
		ctx.SetCurrentDatabase(s.currentDB)
	}
	if s.transaction != nil {
		ctx.SetTransaction(s.transaction)
	}
	return ctx
}

// SetVar sets a session variable
func (s *Session) SetVar(name string, val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars[name] = val
}

// GetVar gets a session variable
func (s *Session) GetVar(name string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vars[name]
}

// User returns the session user
func (s *Session) User() string {
	return s.user
}

// Client returns the client address
func (s *Session) Client() string {
	return s.client
}

// GetTransaction returns the current transaction
func (s *Session) GetTransaction() sql.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.transaction
}

// SetTransaction sets the current transaction
func (s *Session) SetTransaction(txn sql.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transaction = txn
}

// GetAutoCommit returns the autocommit setting
func (s *Session) GetAutoCommit() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.autoCommit
}

// SetAutoCommit sets the autocommit setting
func (s *Session) SetAutoCommit(autoCommit bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoCommit = autoCommit
}

// EnhancedSessionManager manages database sessions with enhanced functionality
type EnhancedSessionManager struct {
	sessions map[uint32]*Session
	mu       sync.RWMutex
	nextID   uint32
}

// NewEnhancedSessionManager creates a new enhanced session manager
func NewEnhancedSessionManager() *EnhancedSessionManager {
	return &EnhancedSessionManager{
		sessions: make(map[uint32]*Session),
	}
}

// NewSession creates a new session for the given connection
func (m *EnhancedSessionManager) NewSession(user, client string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	sess := NewSession(m.nextID, user, client)
	m.sessions[m.nextID] = sess
	return sess
}

// GetSession retrieves a session by ID
func (m *EnhancedSessionManager) GetSession(id uint32) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// RemoveSession removes a session by ID
func (m *EnhancedSessionManager) RemoveSession(id uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}

