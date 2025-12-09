package plan

import (
	"github.com/turtacn/guocedb/compute/sql"
)

// CreateDatabase is a node describing database creation
type CreateDatabase struct {
	name        string
	ifNotExists bool
	catalog     *sql.Catalog
}

// NewCreateDatabase creates a new CreateDatabase node
func NewCreateDatabase(name string, ifNotExists bool) *CreateDatabase {
	return &CreateDatabase{
		name:        name,
		ifNotExists: ifNotExists,
	}
}

// SetCatalog sets the catalog for this node
func (c *CreateDatabase) SetCatalog(cat *sql.Catalog) {
	c.catalog = cat
}

// Resolved implements the Resolvable interface
func (c *CreateDatabase) Resolved() bool {
	return true
}

// RowIter implements the Node interface
func (c *CreateDatabase) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	if c.catalog == nil {
		return nil, sql.ErrDatabaseNotFound.New("catalog not set")
	}

	err := c.catalog.CreateDatabase(ctx, c.name)
	if err != nil && !c.ifNotExists {
		return nil, err
	}
	
	return sql.RowsToRowIter(), nil
}

// Schema implements the Node interface
func (c *CreateDatabase) Schema() sql.Schema { return nil }

// Children implements the Node interface
func (c *CreateDatabase) Children() []sql.Node { return nil }

// TransformUp implements the Transformable interface
func (c *CreateDatabase) TransformUp(f sql.TransformNodeFunc) (sql.Node, error) {
	node := NewCreateDatabase(c.name, c.ifNotExists)
	node.catalog = c.catalog
	return f(node)
}

// TransformExpressionsUp implements the Transformable interface
func (c *CreateDatabase) TransformExpressionsUp(f sql.TransformExprFunc) (sql.Node, error) {
	return c, nil
}

func (c *CreateDatabase) String() string {
	return "CreateDatabase(" + c.name + ")"
}

// Name returns the database name
func (c *CreateDatabase) Name() string {
	return c.name
}

// IfNotExists returns whether IF NOT EXISTS was specified
func (c *CreateDatabase) IfNotExists() bool {
	return c.ifNotExists
}

// DropDatabase is a node describing database deletion
type DropDatabase struct {
	name     string
	ifExists bool
	catalog  *sql.Catalog
}

// NewDropDatabase creates a new DropDatabase node
func NewDropDatabase(name string, ifExists bool) *DropDatabase {
	return &DropDatabase{
		name:     name,
		ifExists: ifExists,
	}
}

// SetCatalog sets the catalog for this node
func (d *DropDatabase) SetCatalog(cat *sql.Catalog) {
	d.catalog = cat
}

// Resolved implements the Resolvable interface
func (d *DropDatabase) Resolved() bool {
	return true
}

// RowIter implements the Node interface
func (d *DropDatabase) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	if d.catalog == nil {
		return nil, sql.ErrDatabaseNotFound.New("catalog not set")
	}

	err := d.catalog.DropDatabase(ctx, d.name)
	if err != nil && !d.ifExists {
		return nil, err
	}
	
	return sql.RowsToRowIter(), nil
}

// Schema implements the Node interface
func (d *DropDatabase) Schema() sql.Schema { return nil }

// Children implements the Node interface
func (d *DropDatabase) Children() []sql.Node { return nil }

// TransformUp implements the Transformable interface
func (d *DropDatabase) TransformUp(f sql.TransformNodeFunc) (sql.Node, error) {
	node := NewDropDatabase(d.name, d.ifExists)
	node.catalog = d.catalog
	return f(node)
}

// TransformExpressionsUp implements the Transformable interface
func (d *DropDatabase) TransformExpressionsUp(f sql.TransformExprFunc) (sql.Node, error) {
	return d, nil
}

func (d *DropDatabase) String() string {
	return "DropDatabase(" + d.name + ")"
}

// Name returns the database name
func (d *DropDatabase) Name() string {
	return d.name
}

// IfExists returns whether IF EXISTS was specified
func (d *DropDatabase) IfExists() bool {
	return d.ifExists
}

// DropTable is a node describing table deletion
type DropTable struct {
	Database sql.Database
	name     string
	ifExists bool
}

// NewDropTable creates a new DropTable node
func NewDropTable(db sql.Database, name string, ifExists bool) *DropTable {
	return &DropTable{
		Database: db,
		name:     name,
		ifExists: ifExists,
	}
}

// Resolved implements the Resolvable interface
func (d *DropTable) Resolved() bool {
	_, ok := d.Database.(sql.UnresolvedDatabase)
	return !ok
}

// RowIter implements the Node interface
func (d *DropTable) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	return sql.RowsToRowIter(), nil
}

// Schema implements the Node interface
func (d *DropTable) Schema() sql.Schema { return nil }

// Children implements the Node interface
func (d *DropTable) Children() []sql.Node { return nil }

// TransformUp implements the Transformable interface
func (d *DropTable) TransformUp(f sql.TransformNodeFunc) (sql.Node, error) {
	return f(NewDropTable(d.Database, d.name, d.ifExists))
}

// TransformExpressionsUp implements the Transformable interface
func (d *DropTable) TransformExpressionsUp(f sql.TransformExprFunc) (sql.Node, error) {
	return d, nil
}

func (d *DropTable) String() string {
	return "DropTable(" + d.name + ")"
}

// Name returns the table name
func (d *DropTable) Name() string {
	return d.name
}

// IfExists returns whether IF EXISTS was specified
func (d *DropTable) IfExists() bool {
	return d.ifExists
}
