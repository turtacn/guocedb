package kvd

import (
	"errors"

	"github.com/dolthub/go-mysql-server/sql"
	"github.comcom/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/interfaces"
	// "github.com/turtacn/guocedb/common/constants"
	// "github.com/turtacn/guocedb/storage/sal"
)

// KVDStorage is a placeholder struct for the KVD storage engine.
type KVDStorage struct{}

var ErrNotImplemented = errors.New("KVD storage engine is not implemented")

// init would register the KVD storage engine. It is commented out until
// the engine is implemented.
//
// func init() {
// 	sal.RegisterStorageEngine(constants.StorageEngineKVD, NewKVDStorage)
// }

// NewKVDStorage creates a new, unimplemented instance of the KVD storage engine.
func NewKVDStorage(cfg *config.StorageConfig) (interfaces.Storage, error) {
	return &KVDStorage{}, ErrNotImplemented
}

func (k *KVDStorage) NewTransaction(ctx *sql.Context, readOnly bool) (interfaces.Transaction, error) {
	return nil, ErrNotImplemented
}

func (k *KVDStorage) CreateDatabase(ctx *sql.Context, name string) error {
	return ErrNotImplemented
}

func (k *KVDStorage) DropDatabase(ctx *sql.Context, name string) error {
	return ErrNotImplemented
}

func (k *KVDStorage) CreateTable(ctx *sql.Context, dbName, tableName string, schema sql.Schema) error {
	return ErrNotImplemented
}

func (k *KVDStorage) DropTable(ctx *sql.Context, dbName, tableName string) error {
	return ErrNotImplemented
}

func (k *KVDStorage) GetTable(ctx *sql.Context, dbName, tableName string) (sql.Table, error) {
	return nil, ErrNotImplemented
}

func (k *KVDStorage) Close() error {
	return nil // Nothing to close
}
