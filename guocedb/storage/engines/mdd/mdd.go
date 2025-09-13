package mdd

import (
	"errors"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/interfaces"
	// "github.com/turtacn/guocedb/common/constants"
	// "github.com/turtacn/guocedb/storage/sal"
)

// MDDStorage is a placeholder struct for the MDD storage engine.
type MDDStorage struct{}

var ErrNotImplemented = errors.New("MDD storage engine is not implemented")

// init would register the MDD storage engine. It is commented out until
// the engine is implemented.
//
// func init() {
// 	sal.RegisterStorageEngine(constants.StorageEngineMDD, NewMDDStorage)
// }

// NewMDDStorage creates a new, unimplemented instance of the MDD storage engine.
func NewMDDStorage(cfg *config.StorageConfig) (interfaces.Storage, error) {
	return &MDDStorage{}, ErrNotImplemented
}

func (m *MDDStorage) NewTransaction(ctx *sql.Context, readOnly bool) (interfaces.Transaction, error) {
	return nil, ErrNotImplemented
}

func (m *MDDStorage) CreateDatabase(ctx *sql.Context, name string) error {
	return ErrNotImplemented
}

func (m *MDDStorage) DropDatabase(ctx *sql.Context, name string) error {
	return ErrNotImplemented
}

func (m *MDDStorage) CreateTable(ctx *sql.Context, dbName, tableName string, schema sql.Schema) error {
	return ErrNotImplemented
}

func (m *MDDStorage) DropTable(ctx *sql.Context, dbName, tableName string) error {
	return ErrNotImplemented
}

func (m *MDDStorage) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	return nil, ErrNotImplemented
}

func (m *MDDStorage) Close() error {
	return nil // Nothing to close
}
