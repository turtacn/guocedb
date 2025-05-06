// Package value defines basic data types used within the database.
// This package might wrap or extend types from go-mysql-server/sql/types.
// value 包定义了数据库内部使用的基本数据类型。
// 此包可能会包装或扩展 go-mysql-server/sql/types 中的类型。
package value

// Importing necessary types from go-mysql-server/sql
// 导入 go-mysql-server/sql 中的必要类型
import "github.com/dolthub/go-mysql-server/sql"

// Value represents a single data value in a row.
// It is typically an interface{} but specific types might be used for clarity or optimization.
// Value 表示行中的一个单一数据值。
// 通常是 interface{} 类型，但为了清晰或优化，可能会使用特定类型。
// We will primarily use sql.Row and the types within sql package.
// 我们主要使用 sql.Row 以及 sql 包中的类型。

// Row represents a row of data, using sql.Row.
// Row 表示一行数据，使用 sql.Row。
type Row = sql.Row

// Schema represents the schema of a table, using sql.Schema.
// Schema 表示表的模式，使用 sql.Schema。
type Schema = sql.Schema

// Column represents a column definition, using sql.Column.
// Column 表示列定义，使用 sql.Column。
type Column = sql.Column

// Type represents a SQL data type, using sql.Type.
// Type 表示 SQL 数据类型，使用 sql.Type。
type Type = sql.Type

// Note: Instead of defining custom types here, we directly use the types provided
// by go-mysql-server/sql. This ensures compatibility with the query engine.
//
// 注意：我们不在此处定义自定义类型，而是直接使用 go-mysql-server/sql
// 提供的类型。这确保了与查询引擎的兼容性。

// Add any custom types or helper functions related to value manipulation if necessary,
// but align with sql package types.
// 如果需要，可以在此处添加与值操作相关的任何自定义类型或辅助函数，
// 但要与 sql 包类型保持一致。