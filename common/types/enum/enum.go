// Package enum 定义了 GuoceDB 系统中使用的枚举类型
// Package enum defines enumeration types used in GuoceDB system
package enum

import (
	"fmt"
	"strings"
)

// ===== 数据类型枚举 Data Type Enums =====

// DataType 数据类型
// DataType data type
type DataType int

const (
	// TypeUnknown 未知类型
	// TypeUnknown unknown type
	TypeUnknown DataType = iota

	// ===== 数值类型 Numeric Types =====

	// TypeTinyInt TINYINT 类型 (1字节)
	// TypeTinyInt TINYINT type (1 byte)
	TypeTinyInt
	// TypeSmallInt SMALLINT 类型 (2字节)
	// TypeSmallInt SMALLINT type (2 bytes)
	TypeSmallInt
	// TypeMediumInt MEDIUMINT 类型 (3字节)
	// TypeMediumInt MEDIUMINT type (3 bytes)
	TypeMediumInt
	// TypeInt INT 类型 (4字节)
	// TypeInt INT type (4 bytes)
	TypeInt
	// TypeBigInt BIGINT 类型 (8字节)
	// TypeBigInt BIGINT type (8 bytes)
	TypeBigInt

	// TypeFloat FLOAT 类型 (4字节)
	// TypeFloat FLOAT type (4 bytes)
	TypeFloat
	// TypeDouble DOUBLE 类型 (8字节)
	// TypeDouble DOUBLE type (8 bytes)
	TypeDouble
	// TypeDecimal DECIMAL 类型 (精确数值)
	// TypeDecimal DECIMAL type (exact numeric)
	TypeDecimal

	// ===== 字符串类型 String Types =====

	// TypeChar CHAR 类型 (定长)
	// TypeChar CHAR type (fixed length)
	TypeChar
	// TypeVarchar VARCHAR 类型 (变长)
	// TypeVarchar VARCHAR type (variable length)
	TypeVarchar
	// TypeText TEXT 类型
	// TypeText TEXT type
	TypeText
	// TypeTinyText TINYTEXT 类型
	// TypeTinyText TINYTEXT type
	TypeTinyText
	// TypeMediumText MEDIUMTEXT 类型
	// TypeMediumText MEDIUMTEXT type
	TypeMediumText
	// TypeLongText LONGTEXT 类型
	// TypeLongText LONGTEXT type
	TypeLongText

	// ===== 二进制类型 Binary Types =====

	// TypeBinary BINARY 类型 (定长)
	// TypeBinary BINARY type (fixed length)
	TypeBinary
	// TypeVarBinary VARBINARY 类型 (变长)
	// TypeVarBinary VARBINARY type (variable length)
	TypeVarBinary
	// TypeBlob BLOB 类型
	// TypeBlob BLOB type
	TypeBlob
	// TypeTinyBlob TINYBLOB 类型
	// TypeTinyBlob TINYBLOB type
	TypeTinyBlob
	// TypeMediumBlob MEDIUMBLOB 类型
	// TypeMediumBlob MEDIUMBLOB type
	TypeMediumBlob
	// TypeLongBlob LONGBLOB 类型
	// TypeLongBlob LONGBLOB type
	TypeLongBlob

	// ===== 日期时间类型 Date/Time Types =====

	// TypeDate DATE 类型
	// TypeDate DATE type
	TypeDate
	// TypeTime TIME 类型
	// TypeTime TIME type
	TypeTime
	// TypeDateTime DATETIME 类型
	// TypeDateTime DATETIME type
	TypeDateTime
	// TypeTimestamp TIMESTAMP 类型
	// TypeTimestamp TIMESTAMP type
	TypeTimestamp
	// TypeYear YEAR 类型
	// TypeYear YEAR type
	TypeYear

	// ===== 其他类型 Other Types =====

	// TypeBool BOOL 类型
	// TypeBool BOOL type
	TypeBool
	// TypeJSON JSON 类型
	// TypeJSON JSON type
	TypeJSON
	// TypeEnum ENUM 类型
	// TypeEnum ENUM type
	TypeEnum
	// TypeSet SET 类型
	// TypeSet SET type
	TypeSet
	// TypeBit BIT 类型
	// TypeBit BIT type
	TypeBit
	// TypeGeometry GEOMETRY 类型
	// TypeGeometry GEOMETRY type
	TypeGeometry
)

// String 返回数据类型的字符串表示
// String returns string representation of data type
func (t DataType) String() string {
	switch t {
	case TypeUnknown:
		return "UNKNOWN"
	case TypeTinyInt:
		return "TINYINT"
	case TypeSmallInt:
		return "SMALLINT"
	case TypeMediumInt:
		return "MEDIUMINT"
	case TypeInt:
		return "INT"
	case TypeBigInt:
		return "BIGINT"
	case TypeFloat:
		return "FLOAT"
	case TypeDouble:
		return "DOUBLE"
	case TypeDecimal:
		return "DECIMAL"
	case TypeChar:
		return "CHAR"
	case TypeVarchar:
		return "VARCHAR"
	case TypeText:
		return "TEXT"
	case TypeTinyText:
		return "TINYTEXT"
	case TypeMediumText:
		return "MEDIUMTEXT"
	case TypeLongText:
		return "LONGTEXT"
	case TypeBinary:
		return "BINARY"
	case TypeVarBinary:
		return "VARBINARY"
	case TypeBlob:
		return "BLOB"
	case TypeTinyBlob:
		return "TINYBLOB"
	case TypeMediumBlob:
		return "MEDIUMBLOB"
	case TypeLongBlob:
		return "LONGBLOB"
	case TypeDate:
		return "DATE"
	case TypeTime:
		return "TIME"
	case TypeDateTime:
		return "DATETIME"
	case TypeTimestamp:
		return "TIMESTAMP"
	case TypeYear:
		return "YEAR"
	case TypeBool:
		return "BOOL"
	case TypeJSON:
		return "JSON"
	case TypeEnum:
		return "ENUM"
	case TypeSet:
		return "SET"
	case TypeBit:
		return "BIT"
	case TypeGeometry:
		return "GEOMETRY"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// IsNumeric 判断是否为数值类型
// IsNumeric checks if type is numeric
func (t DataType) IsNumeric() bool {
	switch t {
	case TypeTinyInt, TypeSmallInt, TypeMediumInt, TypeInt, TypeBigInt,
		TypeFloat, TypeDouble, TypeDecimal:
		return true
	default:
		return false
	}
}

// IsString 判断是否为字符串类型
// IsString checks if type is string
func (t DataType) IsString() bool {
	switch t {
	case TypeChar, TypeVarchar, TypeText, TypeTinyText, TypeMediumText, TypeLongText:
		return true
	default:
		return false
	}
}

// IsBinary 判断是否为二进制类型
// IsBinary checks if type is binary
func (t DataType) IsBinary() bool {
	switch t {
	case TypeBinary, TypeVarBinary, TypeBlob, TypeTinyBlob, TypeMediumBlob, TypeLongBlob:
		return true
	default:
		return false
	}
}

// IsDateTime 判断是否为日期时间类型
// IsDateTime checks if type is date/time
func (t DataType) IsDateTime() bool {
	switch t {
	case TypeDate, TypeTime, TypeDateTime, TypeTimestamp, TypeYear:
		return true
	default:
		return false
	}
}

// ===== 索引类型枚举 Index Type Enums =====

// IndexType 索引类型
// IndexType index type
type IndexType int

const (
	// IndexTypeBTree B树索引
	// IndexTypeBTree B-tree index
	IndexTypeBTree IndexType = iota
	// IndexTypeHash 哈希索引
	// IndexTypeHash hash index
	IndexTypeHash
	// IndexTypeFullText 全文索引
	// IndexTypeFullText full-text index
	IndexTypeFullText
	// IndexTypeSpatial 空间索引
	// IndexTypeSpatial spatial index
	IndexTypeSpatial
)

// String 返回索引类型的字符串表示
// String returns string representation of index type
func (t IndexType) String() string {
	switch t {
	case IndexTypeBTree:
		return "BTREE"
	case IndexTypeHash:
		return "HASH"
	case IndexTypeFullText:
		return "FULLTEXT"
	case IndexTypeSpatial:
		return "SPATIAL"
	default:
		return "UNKNOWN"
	}
}

// ===== 约束类型枚举 Constraint Type Enums =====

// ConstraintType 约束类型
// ConstraintType constraint type
type ConstraintType int

const (
	// ConstraintPrimaryKey 主键约束
	// ConstraintPrimaryKey primary key constraint
	ConstraintPrimaryKey ConstraintType = iota
	// ConstraintUnique 唯一约束
	// ConstraintUnique unique constraint
	ConstraintUnique
	// ConstraintForeignKey 外键约束
	// ConstraintForeignKey foreign key constraint
	ConstraintForeignKey
	// ConstraintCheck 检查约束
	// ConstraintCheck check constraint
	ConstraintCheck
	// ConstraintNotNull 非空约束
	// ConstraintNotNull not null constraint
	ConstraintNotNull
	// ConstraintDefault 默认值约束
	// ConstraintDefault default constraint
	ConstraintDefault
)

// String 返回约束类型的字符串表示
// String returns string representation of constraint type
func (t ConstraintType) String() string {
	switch t {
	case ConstraintPrimaryKey:
		return "PRIMARY KEY"
	case ConstraintUnique:
		return "UNIQUE"
	case ConstraintForeignKey:
		return "FOREIGN KEY"
	case ConstraintCheck:
		return "CHECK"
	case ConstraintNotNull:
		return "NOT NULL"
	case ConstraintDefault:
		return "DEFAULT"
	default:
		return "UNKNOWN"
	}
}

// ===== 操作类型枚举 Operation Type Enums =====

// OperationType 操作类型
// OperationType operation type
type OperationType int

const (
	// OpSelect SELECT 操作
	// OpSelect SELECT operation
	OpSelect OperationType = iota
	// OpInsert INSERT 操作
	// OpInsert INSERT operation
	OpInsert
	// OpUpdate UPDATE 操作
	// OpUpdate UPDATE operation
	OpUpdate
	// OpDelete DELETE 操作
	// OpDelete DELETE operation
	OpDelete
	// OpCreate CREATE 操作
	// OpCreate CREATE operation
	OpCreate
	// OpDrop DROP 操作
	// OpDrop DROP operation
	OpDrop
	// OpAlter ALTER 操作
	// OpAlter ALTER operation
	OpAlter
	// OpTruncate TRUNCATE 操作
	// OpTruncate TRUNCATE operation
	OpTruncate
	// OpRename RENAME 操作
	// OpRename RENAME operation
	OpRename
	// OpGrant GRANT 操作
	// OpGrant GRANT operation
	OpGrant
	// OpRevoke REVOKE 操作
	// OpRevoke REVOKE operation
	OpRevoke
	// OpBegin BEGIN 操作
	// OpBegin BEGIN operation
	OpBegin
	// OpCommit COMMIT 操作
	// OpCommit COMMIT operation
	OpCommit
	// OpRollback ROLLBACK 操作
	// OpRollback ROLLBACK operation
	OpRollback
)

// String 返回操作类型的字符串表示
// String returns string representation of operation type
func (t OperationType) String() string {
	switch t {
	case OpSelect:
		return "SELECT"
	case OpInsert:
		return "INSERT"
	case OpUpdate:
		return "UPDATE"
	case OpDelete:
		return "DELETE"
	case OpCreate:
		return "CREATE"
	case OpDrop:
		return "DROP"
	case OpAlter:
		return "ALTER"
	case OpTruncate:
		return "TRUNCATE"
	case OpRename:
		return "RENAME"
	case OpGrant:
		return "GRANT"
	case OpRevoke:
		return "REVOKE"
	case OpBegin:
		return "BEGIN"
	case OpCommit:
		return "COMMIT"
	case OpRollback:
		return "ROLLBACK"
	default:
		return "UNKNOWN"
	}
}

// IsDML 判断是否为DML操作
// IsDML checks if operation is DML
func (t OperationType) IsDML() bool {
	switch t {
	case OpSelect, OpInsert, OpUpdate, OpDelete:
		return true
	default:
		return false
	}
}

// IsDDL 判断是否为DDL操作
// IsDDL checks if operation is DDL
func (t OperationType) IsDDL() bool {
	switch t {
	case OpCreate, OpDrop, OpAlter, OpTruncate, OpRename:
		return true
	default:
		return false
	}
}

// IsDCL 判断是否为DCL操作
// IsDCL checks if operation is DCL
func (t OperationType) IsDCL() bool {
	switch t {
	case OpGrant, OpRevoke:
		return true
	default:
		return false
	}
}

// IsTCL 判断是否为TCL操作
// IsTCL checks if operation is TCL
func (t OperationType) IsTCL() bool {
	switch t {
	case OpBegin, OpCommit, OpRollback:
		return true
	default:
		return false
	}
}

// ===== 连接类型枚举 Join Type Enums =====

// JoinType 连接类型
// JoinType join type
type JoinType int

const (
	// JoinInner 内连接
	// JoinInner inner join
	JoinInner JoinType = iota
	// JoinLeft 左连接
	// JoinLeft left join
	JoinLeft
	// JoinRight 右连接
	// JoinRight right join
	JoinRight
	// JoinFull 全连接
	// JoinFull full join
	JoinFull
	// JoinCross 交叉连接
	// JoinCross cross join
	JoinCross
	// JoinNatural 自然连接
	// JoinNatural natural join
	JoinNatural
	// JoinSemi 半连接
	// JoinSemi semi join
	JoinSemi
	// JoinAnti 反连接
	// JoinAnti anti join
	JoinAnti
)

// String 返回连接类型的字符串表示
// String returns string representation of join type
func (t JoinType) String() string {
	switch t {
	case JoinInner:
		return "INNER JOIN"
	case JoinLeft:
		return "LEFT JOIN"
	case JoinRight:
		return "RIGHT JOIN"
	case JoinFull:
		return "FULL JOIN"
	case JoinCross:
		return "CROSS JOIN"
	case JoinNatural:
		return "NATURAL JOIN"
	case JoinSemi:
		return "SEMI JOIN"
	case JoinAnti:
		return "ANTI JOIN"
	default:
		return "UNKNOWN JOIN"
	}
}

// ===== 排序方向枚举 Sort Direction Enums =====

// SortDirection 排序方向
// SortDirection sort direction
type SortDirection int

const (
	// SortAsc 升序
	// SortAsc ascending order
	SortAsc SortDirection = iota
	// SortDesc 降序
	// SortDesc descending order
	SortDesc
)

// String 返回排序方向的字符串表示
// String returns string representation of sort direction
func (d SortDirection) String() string {
	switch d {
	case SortAsc:
		return "ASC"
	case SortDesc:
		return "DESC"
	default:
		return "UNKNOWN"
	}
}

// ===== 事务状态枚举 Transaction State Enums =====

// TransactionState 事务状态
// TransactionState transaction state
type TransactionState int

const (
	// TxStateNone 无事务
	// TxStateNone no transaction
	TxStateNone TransactionState = iota
	// TxStateActive 活跃事务
	// TxStateActive active transaction
	TxStateActive
	// TxStateCommitting 提交中
	// TxStateCommitting committing
	TxStateCommitting
	// TxStateRollingBack 回滚中
	// TxStateRollingBack rolling back
	TxStateRollingBack
	// TxStateCommitted 已提交
	// TxStateCommitted committed
	TxStateCommitted
	// TxStateAborted 已中止
	// TxStateAborted aborted
	TxStateAborted
)

// String 返回事务状态的字符串表示
// String returns string representation of transaction state
func (s TransactionState) String() string {
	switch s {
	case TxStateNone:
		return "NONE"
	case TxStateActive:
		return "ACTIVE"
	case TxStateCommitting:
		return "COMMITTING"
	case TxStateRollingBack:
		return "ROLLING_BACK"
	case TxStateCommitted:
		return "COMMITTED"
	case TxStateAborted:
		return "ABORTED"
	default:
		return "UNKNOWN"
	}
}

// ===== 锁类型枚举 Lock Type Enums =====

// LockType 锁类型
// LockType lock type
type LockType int

const (
	// LockNone 无锁
	// LockNone no lock
	LockNone LockType = iota
	// LockShared 共享锁
	// LockShared shared lock
	LockShared
	// LockExclusive 排他锁
	// LockExclusive exclusive lock
	LockExclusive
	// LockIntentShared 意向共享锁
	// LockIntentShared intent shared lock
	LockIntentShared
	// LockIntentExclusive 意向排他锁
	// LockIntentExclusive intent exclusive lock
	LockIntentExclusive
	// LockSharedIntentExclusive 共享意向排他锁
	// LockSharedIntentExclusive shared intent exclusive lock
	LockSharedIntentExclusive
)

// String 返回锁类型的字符串表示
// String returns string representation of lock type
func (t LockType) String() string {
	switch t {
	case LockNone:
		return "NONE"
	case LockShared:
		return "S"
	case LockExclusive:
		return "X"
	case LockIntentShared:
		return "IS"
	case LockIntentExclusive:
		return "IX"
	case LockSharedIntentExclusive:
		return "SIX"
	default:
		return "UNKNOWN"
	}
}

// IsCompatible 判断两个锁是否兼容
// IsCompatible checks if two locks are compatible
func (t LockType) IsCompatible(other LockType) bool {
	// 锁兼容性矩阵
	// Lock compatibility matrix
	compatibility := map[LockType]map[LockType]bool{
		LockNone: {
			LockNone: true, LockShared: true, LockExclusive: true,
			LockIntentShared: true, LockIntentExclusive: true, LockSharedIntentExclusive: true,
		},
		LockShared: {
			LockNone: true, LockShared: true, LockExclusive: false,
			LockIntentShared: true, LockIntentExclusive: false, LockSharedIntentExclusive: false,
		},
		LockExclusive: {
			LockNone: true, LockShared: false, LockExclusive: false,
			LockIntentShared: false, LockIntentExclusive: false, LockSharedIntentExclusive: false,
		},
		LockIntentShared: {
			LockNone: true, LockShared: true, LockExclusive: false,
			LockIntentShared: true, LockIntentExclusive: true, LockSharedIntentExclusive: true,
		},
		LockIntentExclusive: {
			LockNone: true, LockShared: false, LockExclusive: false,
			LockIntentShared: true, LockIntentExclusive: true, LockSharedIntentExclusive: false,
		},
		LockSharedIntentExclusive: {
			LockNone: true, LockShared: false, LockExclusive: false,
			LockIntentShared: true, LockIntentExclusive: false, LockSharedIntentExclusive: false,
		},
	}

	if compat, ok := compatibility[t]; ok {
		if result, ok := compat[other]; ok {
			return result
		}
	}
	return false
}

// ===== 比较操作符枚举 Comparison Operator Enums =====

// ComparisonOp 比较操作符
// ComparisonOp comparison operator
type ComparisonOp int

const (
	// OpEqual 等于
	// OpEqual equal
	OpEqual ComparisonOp = iota
	// OpNotEqual 不等于
	// OpNotEqual not equal
	OpNotEqual
	// OpLess 小于
	// OpLess less than
	OpLess
	// OpLessOrEqual 小于等于
	// OpLessOrEqual less than or equal
	OpLessOrEqual
	// OpGreater 大于
	// OpGreater greater than
	OpGreater
	// OpGreaterOrEqual 大于等于
	// OpGreaterOrEqual greater than or equal
	OpGreaterOrEqual
	// OpIn IN 操作
	// OpIn IN operation
	OpIn
	// OpNotIn NOT IN 操作
	// OpNotIn NOT IN operation
	OpNotIn
	// OpLike LIKE 操作
	// OpLike LIKE operation
	OpLike
	// OpNotLike NOT LIKE 操作
	// OpNotLike NOT LIKE operation
	OpNotLike
	// OpBetween BETWEEN 操作
	// OpBetween BETWEEN operation
	OpBetween
	// OpNotBetween NOT BETWEEN 操作
	// OpNotBetween NOT BETWEEN operation
	OpNotBetween
	// OpIsNull IS NULL 操作
	// OpIsNull IS NULL operation
	OpIsNull
	// OpIsNotNull IS NOT NULL 操作
	// OpIsNotNull IS NOT NULL operation
	OpIsNotNull
	// OpExists EXISTS 操作
	// OpExists EXISTS operation
	OpExists
	// OpNotExists NOT EXISTS 操作
	// OpNotExists NOT EXISTS operation
	OpNotExists
)

// String 返回比较操作符的字符串表示
// String returns string representation of comparison operator
func (op ComparisonOp) String() string {
	switch op {
	case OpEqual:
		return "="
	case OpNotEqual:
		return "<>"
	case OpLess:
		return "<"
	case OpLessOrEqual:
		return "<="
	case OpGreater:
		return ">"
	case OpGreaterOrEqual:
		return ">="
	case OpIn:
		return "IN"
	case OpNotIn:
		return "NOT IN"
	case OpLike:
		return "LIKE"
	case OpNotLike:
		return "NOT LIKE"
	case OpBetween:
		return "BETWEEN"
	case OpNotBetween:
		return "NOT BETWEEN"
	case OpIsNull:
		return "IS NULL"
	case OpIsNotNull:
		return "IS NOT NULL"
	case OpExists:
		return "EXISTS"
	case OpNotExists:
		return "NOT EXISTS"
	default:
		return "UNKNOWN"
	}
}

// ===== 聚合函数枚举 Aggregate Function Enums =====

// AggregateFunc 聚合函数
// AggregateFunc aggregate function
type AggregateFunc int

const (
	// AggCount COUNT 函数
	// AggCount COUNT function
	AggCount AggregateFunc = iota
	// AggSum SUM 函数
	// AggSum SUM function
	AggSum
	// AggAvg AVG 函数
	// AggAvg AVG function
	AggAvg
	// AggMin MIN 函数
	// AggMin MIN function
	AggMin
	// AggMax MAX 函数
	// AggMax MAX function
	AggMax
	// AggGroupConcat GROUP_CONCAT 函数
	// AggGroupConcat GROUP_CONCAT function
	AggGroupConcat
	// AggStdDev 标准差函数
	// AggStdDev standard deviation function
	AggStdDev
	// AggVariance 方差函数
	// AggVariance variance function
	AggVariance
)

// String 返回聚合函数的字符串表示
// String returns string representation of aggregate function
func (f AggregateFunc) String() string {
	switch f {
	case AggCount:
		return "COUNT"
	case AggSum:
		return "SUM"
	case AggAvg:
		return "AVG"
	case AggMin:
		return "MIN"
	case AggMax:
		return "MAX"
	case AggGroupConcat:
		return "GROUP_CONCAT"
	case AggStdDev:
		return "STDDEV"
	case AggVariance:
		return "VARIANCE"
	default:
		return "UNKNOWN"
	}
}

// ===== 权限级别枚举 Privilege Level Enums =====

// PrivilegeLevel 权限级别
// PrivilegeLevel privilege level
type PrivilegeLevel int

const (
	// PrivLevelGlobal 全局级别
	// PrivLevelGlobal global level
	PrivLevelGlobal PrivilegeLevel = iota
	// PrivLevelDatabase 数据库级别
	// PrivLevelDatabase database level
	PrivLevelDatabase
	// PrivLevelTable 表级别
	// PrivLevelTable table level
	PrivLevelTable
	// PrivLevelColumn 列级别
	// PrivLevelColumn column level
	PrivLevelColumn
	// PrivLevelRoutine 存储过程级别
	// PrivLevelRoutine routine level
	PrivLevelRoutine
)

// String 返回权限级别的字符串表示
// String returns string representation of privilege level
func (l PrivilegeLevel) String() string {
	switch l {
	case PrivLevelGlobal:
		return "GLOBAL"
	case PrivLevelDatabase:
		return "DATABASE"
	case PrivLevelTable:
		return "TABLE"
	case PrivLevelColumn:
		return "COLUMN"
	case PrivLevelRoutine:
		return "ROUTINE"
	default:
		return "UNKNOWN"
	}
}

// ===== 节点类型枚举 Node Type Enums =====

// NodeType 执行计划节点类型
// NodeType execution plan node type
type NodeType int

const (
	// NodeScan 扫描节点
	// NodeScan scan node
	NodeScan NodeType = iota
	// NodeFilter 过滤节点
	// NodeFilter filter node
	NodeFilter
	// NodeProject 投影节点
	// NodeProject projection node
	NodeProject
	// NodeJoin 连接节点
	// NodeJoin join node
	NodeJoin
	// NodeAggregate 聚合节点
	// NodeAggregate aggregate node
	NodeAggregate
	// NodeSort 排序节点
	// NodeSort sort node
	NodeSort
	// NodeLimit 限制节点
	// NodeLimit limit node
	NodeLimit
	// NodeUnion 联合节点
	// NodeUnion union node
	NodeUnion
	// NodeIntersect 交集节点
	// NodeIntersect intersect node
	NodeIntersect
	// NodeExcept 差集节点
	// NodeExcept except node
	NodeExcept
)

// String 返回节点类型的字符串表示
// String returns string representation of node type
func (t NodeType) String() string {
	switch t {
	case NodeScan:
		return "SCAN"
	case NodeFilter:
		return "FILTER"
	case NodeProject:
		return "PROJECT"
	case NodeJoin:
		return "JOIN"
	case NodeAggregate:
		return "AGGREGATE"
	case NodeSort:
		return "SORT"
	case NodeLimit:
		return "LIMIT"
	case NodeUnion:
		return "UNION"
	case NodeIntersect:
		return "INTERSECT"
	case NodeExcept:
		return "EXCEPT"
	default:
		return "UNKNOWN"
	}
}

// ===== 辅助函数 Helper Functions =====

// ParseDataType 解析数据类型字符串
// ParseDataType parses data type string
func ParseDataType(typeStr string) (DataType, error) {
	typeStr = strings.ToUpper(strings.TrimSpace(typeStr))

	// 移除括号中的长度/精度信息
	// Remove length/precision info in parentheses
	if idx := strings.Index(typeStr, "("); idx != -1 {
		typeStr = typeStr[:idx]
	}

	switch typeStr {
	case "TINYINT":
		return TypeTinyInt, nil
	case "SMALLINT":
		return TypeSmallInt, nil
	case "MEDIUMINT":
		return TypeMediumInt, nil
	case "INT", "INTEGER":
		return TypeInt, nil
	case "BIGINT":
		return TypeBigInt, nil
	case "FLOAT":
		return TypeFloat, nil
	case "DOUBLE", "DOUBLE PRECISION":
		return TypeDouble, nil
	case "DECIMAL", "NUMERIC", "DEC":
		return TypeDecimal, nil
	case "CHAR":
		return TypeChar, nil
	case "VARCHAR":
		return TypeVarchar, nil
	case "TEXT":
		return TypeText, nil
	case "TINYTEXT":
		return TypeTinyText, nil
	case "MEDIUMTEXT":
		return TypeMediumText, nil
	case "LONGTEXT":
		return TypeLongText, nil
	case "BINARY":
		return TypeBinary, nil
	case "VARBINARY":
		return TypeVarBinary, nil
	case "BLOB":
		return TypeBlob, nil
	case "TINYBLOB":
		return TypeTinyBlob, nil
	case "MEDIUMBLOB":
		return TypeMediumBlob, nil
	case "LONGBLOB":
		return TypeLongBlob, nil
	case "DATE":
		return TypeDate, nil
	case "TIME":
		return TypeTime, nil
	case "DATETIME":
		return TypeDateTime, nil
	case "TIMESTAMP":
		return TypeTimestamp, nil
	case "YEAR":
		return TypeYear, nil
	case "BOOL", "BOOLEAN":
		return TypeBool, nil
	case "JSON":
		return TypeJSON, nil
	case "ENUM":
		return TypeEnum, nil
	case "SET":
		return TypeSet, nil
	case "BIT":
		return TypeBit, nil
	case "GEOMETRY":
		return TypeGeometry, nil
	default:
		return TypeUnknown, fmt.Errorf("unknown data type: %s", typeStr)
	}
}

// IsValidIdentifier 检查是否为有效的标识符
// IsValidIdentifier checks if string is valid identifier
func IsValidIdentifier(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}

	// 首字符必须是字母或下划线
	// First character must be letter or underscore
	if !isLetter(rune(name[0])) && name[0] != '_' {
		return false
	}

	// 后续字符可以是字母、数字或下划线
	// Subsequent characters can be letters, digits or underscore
	for i := 1; i < len(name); i++ {
		ch := rune(name[i])
		if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			return false
		}
	}

	return true
}

// isLetter 判断是否为字母
// isLetter checks if rune is letter
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit 判断是否为数字
// isDigit checks if rune is digit
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}
