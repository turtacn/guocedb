package server

import (
	"io"
	"github.com/turtacn/guocedb/compute/sql"
)

// mockDatabase implements a simple database for testing
type mockDatabase struct {
	name   string
	tables map[string]sql.Table
}

func newMockDatabase(name string) *mockDatabase {
	return &mockDatabase{
		name:   name,
		tables: make(map[string]sql.Table),
	}
}

func (m *mockDatabase) Name() string {
	return m.name
}

func (m *mockDatabase) Tables() map[string]sql.Table {
	return m.tables
}

func (m *mockDatabase) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	table, ok := m.tables[tblName]
	return table, ok, nil
}

func (m *mockDatabase) GetTableNames(ctx *sql.Context) ([]string, error) {
	names := make([]string, 0, len(m.tables))
	for name := range m.tables {
		names = append(names, name)
	}
	return names, nil
}

// mockTable implements a simple table for testing
type mockTable struct {
	name   string
	schema sql.Schema
	rows   []sql.Row
}

func newMockTable(name string, schema sql.Schema, rows []sql.Row) *mockTable {
	return &mockTable{
		name:   name,
		schema: schema,
		rows:   rows,
	}
}

func (m *mockTable) Name() string {
	return m.name
}

func (m *mockTable) String() string {
	return m.name
}

func (m *mockTable) Schema() sql.Schema {
	return m.schema
}

func (m *mockTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return &mockPartitionIter{partitions: []sql.Partition{&mockPartition{}}}, nil
}

func (m *mockTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &mockRowIter{rows: m.rows}, nil
}

// mockPartition implements sql.Partition
type mockPartition struct{}

func (m *mockPartition) Key() []byte {
	return []byte("mock")
}

// mockPartitionIter implements sql.PartitionIter
type mockPartitionIter struct {
	partitions []sql.Partition
	pos        int
}

func (m *mockPartitionIter) Next() (sql.Partition, error) {
	if m.pos >= len(m.partitions) {
		return nil, io.EOF
	}
	partition := m.partitions[m.pos]
	m.pos++
	return partition, nil
}

func (m *mockPartitionIter) Close() error {
	return nil
}

// mockRowIter implements sql.RowIter
type mockRowIter struct {
	rows []sql.Row
	pos  int
}

func (m *mockRowIter) Next() (sql.Row, error) {
	if m.pos >= len(m.rows) {
		return nil, io.EOF
	}
	row := m.rows[m.pos]
	m.pos++
	return row, nil
}

func (m *mockRowIter) Close() error {
	return nil
}