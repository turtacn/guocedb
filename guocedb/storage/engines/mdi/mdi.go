package mdi

import (
	"errors"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/interfaces"
	// "github.com/turtacn/guocedb/common/constants"
	// "github.com/turtacn/guocedb/storage/sal"
)

// MDIStorage is a placeholder struct for the MDI storage engine.
type MDIStorage struct{}

var ErrNotImplemented = errors.New("MDI storage engine is not implemented")

// init would register the MDI storage engine. It is commented out until
// the engine is implemented.
//
// func init() {
// 	sal.RegisterStorageEngine(constants.StorageEngineMDI, NewMDIStorage)
// }

// NewMDIStorage creates a new, unimplemented instance of the MDI storage engine.
func NewMDIStorage(cfg *config.StorageConfig) (interfaces.Storage, error) {
	return &MDIStorage{}, ErrNotImplemented
}

func (m *MDIStorage) NewTransaction(ctx *sql.Context, readOnly bool) (interfaces.Transaction, error) {
	return nil, ErrNotImplemented
}

func (m *MDIStorage) CreateDatabase(ctx *sql.Context, name string) error {
	return ErrNotImplemented
}

func (m *MDIStorage) DropDatabase(ctx *sql.Context, name string) error {
	return ErrNotImplemented
}

func (m *MDIStorage) CreateTable(ctx *sql.Context, dbName, tableName string, schema sql.Schema) error {
	return ErrNotImplemented
}

func (m *MDIStorage) DropTable(ctx *sql.Context, dbName, tableName string) error {
	return ErrNotImplemented
}

func (m *MDIStorage) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	return nil, ErrNotImplemented
}

func (m *MDIStorage) Close() error {
	return nil // Nothing to close
}
