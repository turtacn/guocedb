// Package kvd provides a placeholder for the KVD storage engine.
package kvd

import (
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/interfaces"
)

// Storage is a placeholder for the KVD storage engine.
type Storage struct{}

// NewStorage returns a new instance of the KVD storage engine.
func NewStorage(cfg interface{}) (*Storage, error) {
	return nil, errors.ErrNotImplemented
}

func (s *Storage) Get(ctx *sql.Context, db, table string, key []byte) ([]byte, error) {
	return nil, errors.ErrNotImplemented
}

func (s *Storage) Set(ctx *sql.Context, db, table string, key, value []byte) error {
	return errors.ErrNotImplemented
}

func (s *Storage) Delete(ctx *sql.Context, db, table string, key []byte) error {
	return errors.ErrNotImplemented
}

func (s *Storage) Iterator(ctx *sql.Context, db, table string, prefix []byte) (interfaces.Iterator, error) {
	return nil, errors.ErrNotImplemented
}

func (s *Storage) NewTransaction(ctx *sql.Context, readOnly bool) (interfaces.Transaction, error) {
	return nil, errors.ErrNotImplemented
}

func (s *Storage) CreateDatabase(ctx *sql.Context, name string) error {
	return errors.ErrNotImplemented
}

func (s *Storage) DropDatabase(ctx *sql.Context, name string) error {
	return errors.ErrNotImplemented
}

func (s *Storage) ListDatabases(ctx *sql.Context) ([]string, error) {
	return nil, errors.ErrNotImplemented
}

func (s *Storage) CreateTable(ctx *sql.Context, dbName string, table sql.Table) error {
	return errors.ErrNotImplemented
}

func (s *Storage) DropTable(ctx *sql.Context, dbName, tableName string) error {
	return errors.ErrNotImplemented
}

func (s *Storage) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	return nil, errors.ErrNotImplemented
}

func (s *Storage) ListTables(ctx *sql.Context, dbName string) ([]string, error) {
	return nil, errors.ErrNotImplemented
}

func (s *Storage) Close() error {
	return errors.ErrNotImplemented
}
