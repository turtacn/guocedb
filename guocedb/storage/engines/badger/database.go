package badger

import (
	"github.com/dolthub/go-mysql-server/sql"
)

// BadgerDatabase represents a single database stored in BadgerDB.
// It implements the sql.Database interface.
type BadgerDatabase struct {
	db   *BadgerStorage
	name string
}

var _ sql.Database = (*BadgerDatabase)(nil)
var _ sql.TableCreator = (*BadgerDatabase)(nil)
var _ sql.TableDropper = (*BadgerDatabase)(nil)

// NewBadgerDatabase creates a new BadgerDatabase instance.
func NewBadgerDatabase(db *BadgerStorage, name string) *BadgerDatabase {
	return &BadgerDatabase{
		db:   db,
		name: name,
	}
}

// Name returns the name of the database.
func (d *BadgerDatabase) Name() string {
	return d.name
}

// GetTableInsensitive retrieves a table by name, case-insensitive.
func (d *BadgerDatabase) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	// In a real implementation, this would handle case-insensitivity.
	// For now, we just do a case-sensitive lookup.
	table, err := d.db.GetTable(ctx, d.name, tblName)
	if err != nil {
		return nil, false, err
	}
	if table == nil {
		return nil, false, nil
	}
	return table, true, nil
}

// GetTableNames returns a list of all table names in the database.
func (d *BadgerDatabase) GetTableNames(ctx *sql.Context) ([]string, error) {
	// TODO: Implement logic to scan metadata and return all table names for this database.
	return []string{}, nil
}

// CreateTable creates a new table in the database.
func (d *BadgerDatabase) CreateTable(ctx *sql.Context, name string, schema sql.Schema, collation sql.CollationID) error {
	return d.db.CreateTable(ctx, d.name, name, schema)
}

// DropTable drops a table from the database.
func (d *BadgerDatabase) DropTable(ctx *sql.Context, name string) error {
	return d.db.DropTable(ctx, d.name, name)
}
