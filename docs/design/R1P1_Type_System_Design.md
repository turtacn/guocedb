# R1P1: Type System Design

This document outlines the design of the core SQL type system for the compute engine.

## 1. Overview

The type system is a foundational component of the query engine, responsible for representing and manipulating data of different SQL types. It is designed to be independent of `go-mysql-server` while maintaining full compatibility with its type system.

## 2. Core Components

The type system consists of the following key components:

- **`QueryType` (enum):** An enum that defines all the supported SQL types (e.g., `INT64`, `TEXT`, `TIMESTAMP`).
- **`Type` (interface):** An interface that defines the behavior of a SQL type, including comparison, conversion, and serialization.
- **`Value` (struct):** A wrapper for a SQL value, containing both the data and its `Type`.
- **`Schema` and `Row`:** Structures for representing table schemas and rows of data.
- **`convert.go`:** A library of functions for converting between different data types.

## 3. UML Class Diagram

```mermaid
classDiagram
    class QueryType {
        <<enumeration>>
        INT64
        TEXT
        TIMESTAMP
        ...
    }

    class Type {
        <<interface>>
        QueryType() QueryType
        SQL() string
        Compare(a, b interface{}) (int, error)
        Convert(v interface{}) (interface{}, error)
        Zero() interface{}
    }

    class Value {
        -typ Type
        -data interface{}
        +NewValue(typ Type, data interface{}) (*Value, error)
        +IsNull() bool
        +Compare(other *Value) (int, error)
        +ToBytes() ([]byte, error)
        +FromBytes([]byte) error
    }

    class Schema {
        -columns []*Column
        +IndexOf(columnName string) int
        +CheckRow(row Row) error
    }

    class Column {
        +Name string
        +Type Type
        +Nullable bool
        ...
    }

    class Row {
        -values []interface{}
    }

    Type <|-- Int64Type
    Type <|-- TextType
    Type <|-- TimestampType
    Value o-- Type
    Schema o-- Column
    Column o-- Type
    Row "1" -- "many" interface{}
```

## 4. `go-mysql-server` Type Mapping

The `MysqlTypeToType` function in `compute/types/sql_type.go` maps the `query.Type` enum from `go-mysql-server` to the `Type` interface in our type system. This ensures compatibility with the MySQL wire protocol and other components that rely on the `go-mysql-server` types.

| `go-mysql-server` Type | Our `Type`         |
| ---------------------- | ------------------ |
| `sql.Int8`             | `types.Int64`      |
| `sql.Int64`            | `types.Int64`      |
| `sql.Text`             | `types.Text`       |
| `sql.Timestamp`        | `types.Timestamp`  |
| ...                    | ...                |

## 5. Differences from `go-mysql-server`

- **Code Independence:** The type system is completely independent of the `go-mysql-server` codebase. This allows for more flexibility and control over the implementation.
- **Custom Serialization:** The `Value` struct uses `encoding/gob` for serialization, which is a custom binary format. This is different from the serialization format used by `go-mysql-server`.

## 6. Future Work

- **More Types:** The type system will be extended to support more SQL types, such as `DECIMAL`, `JSON`, and `ENUM`.
- **Performance Optimizations:** The performance of the type system will be benchmarked and optimized as needed.
- **Collation Support:** Collation support will be added to the `stringType` to handle different character sets and sorting rules.
