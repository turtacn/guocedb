// Package badger implements table-level operations for Badger storage engine
// badger包，实现Badger存储引擎的表级别操作
package badger

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/turtacn/guocedb/common/errors"
	logging "github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/common/types"
	"github.com/turtacn/guocedb/common/types/value"
)

// TableManager 表管理器 Table manager
type TableManager struct {
	database      *Database         // 数据库实例 Database instance
	tables        map[string]*Table // 表映射 Table mapping
	mu            sync.RWMutex      // 读写锁 Read-write lock
	logger        logging.Logger    // 日志记录器 Logger
	autoIncrement map[string]int64  // 自增计数器 Auto increment counters
}

// Table 表实例 Table instance
type Table struct {
	name         string                 // 表名 Table name
	database     *Database              // 所属数据库 Parent database
	schema       *types.Schema          // 表结构 Table schema
	indexes      map[string]*Index      // 索引映射 Index mapping
	constraints  map[string]*Constraint // 约束映射 Constraint mapping
	statistics   *TableStatistics       // 表统计信息 Table statistics
	encoder      *KeyEncoder            // 键编码器 Key encoder
	decoder      *KeyDecoder            // 键解码器 Key decoder
	valueEncoder *ValueEncoder          // 值编码器 Value encoder
	logger       logging.Logger         // 日志记录器 Logger
	mu           sync.RWMutex           // 读写锁 Read-write lock
	sequence     int64                  // 序列号 Sequence number
}

// Index 索引结构 Index structure
type Index struct {
	Name       string            `json:"name"`       // 索引名 Index name
	TableName  string            `json:"table_name"` // 表名 Table name
	Columns    []string          `json:"columns"`    // 索引列 Index columns
	Unique     bool              `json:"unique"`     // 唯一索引 Unique index
	Primary    bool              `json:"primary"`    // 主键索引 Primary key index
	Type       string            `json:"type"`       // 索引类型 Index type
	CreatedAt  time.Time         `json:"created_at"` // 创建时间 Creation time
	UpdatedAt  time.Time         `json:"updated_at"` // 更新时间 Update time
	Properties map[string]string `json:"properties"` // 属性 Properties
	Statistics *IndexStatistics  `json:"statistics"` // 统计信息 Statistics
}

// IndexStatistics 索引统计信息 Index statistics
type IndexStatistics struct {
	KeyCount   int64     `json:"key_count"`    // 键数量 Key count
	UniqueKeys int64     `json:"unique_keys"`  // 唯一键数量 Unique key count
	AvgKeySize float64   `json:"avg_key_size"` // 平均键大小 Average key size
	LastUsed   time.Time `json:"last_used"`    // 最后使用时间 Last used
	UsageCount int64     `json:"usage_count"`  // 使用次数 Usage count
}

// Constraint 约束结构 Constraint structure
type Constraint struct {
	Name       string            `json:"name"`        // 约束名 Constraint name
	TableName  string            `json:"table_name"`  // 表名 Table name
	Type       string            `json:"type"`        // 约束类型 Constraint type
	Columns    []string          `json:"columns"`     // 约束列 Constraint columns
	RefTable   string            `json:"ref_table"`   // 引用表 Reference table
	RefColumns []string          `json:"ref_columns"` // 引用列 Reference columns
	OnUpdate   string            `json:"on_update"`   // 更新操作 On update action
	OnDelete   string            `json:"on_delete"`   // 删除操作 On delete action
	CheckExpr  string            `json:"check_expr"`  // 检查表达式 Check expression
	CreatedAt  time.Time         `json:"created_at"`  // 创建时间 Creation time
	Properties map[string]string `json:"properties"`  // 属性 Properties
}

// TableStatistics 表统计信息 Table statistics
type TableStatistics struct {
	RowCount        int64     `json:"row_count"`        // 行数 Row count
	ColumnCount     int       `json:"column_count"`     // 列数 Column count
	IndexCount      int       `json:"index_count"`      // 索引数 Index count
	ConstraintCount int       `json:"constraint_count"` // 约束数 Constraint count
	AvgRowSize      float64   `json:"avg_row_size"`     // 平均行大小 Average row size
	DataSize        int64     `json:"data_size"`        // 数据大小 Data size
	IndexSize       int64     `json:"index_size"`       // 索引大小 Index size
	LastAnalyzed    time.Time `json:"last_analyzed"`    // 最后分析时间 Last analyzed
	CreatedAt       time.Time `json:"created_at"`       // 创建时间 Creation time
	UpdatedAt       time.Time `json:"updated_at"`       // 更新时间 Update time
}

// QueryOptions 查询选项 Query options
type QueryOptions struct {
	Limit     int             `json:"limit"`      // 限制数量 Limit count
	Offset    int             `json:"offset"`     // 偏移量 Offset
	OrderBy   []OrderByClause `json:"order_by"`   // 排序 Order by
	Where     *WhereClause    `json:"where"`      // 条件 Where clause
	GroupBy   []string        `json:"group_by"`   // 分组 Group by
	Having    *WhereClause    `json:"having"`     // Having条件 Having clause
	Distinct  bool            `json:"distinct"`   // 去重 Distinct
	ForUpdate bool            `json:"for_update"` // 锁定 For update
	Timeout   time.Duration   `json:"timeout"`    // 超时 Timeout
}

// OrderByClause 排序子句 Order by clause
type OrderByClause struct {
	Column    string `json:"column"`    // 列名 Column name
	Direction string `json:"direction"` // 方向 ASC/DESC Direction
}

// WhereClause 条件子句 Where clause
type WhereClause struct {
	Operator string         `json:"operator"`  // 操作符 Operator
	Column   string         `json:"column"`    // 列名 Column name
	Value    *types.Value   `json:"value"`     // 值 Value
	Values   []*types.Value `json:"values"`    // 多个值 Multiple values
	SubQuery *QueryOptions  `json:"sub_query"` // 子查询 Sub query
	Left     *WhereClause   `json:"left"`      // 左条件 Left condition
	Right    *WhereClause   `json:"right"`     // 右条件 Right condition
}

// NewTableManager 创建表管理器 Create table manager
func NewTableManager(database *Database, logger logging.Logger) *TableManager {
	return &TableManager{
		database:      database,
		tables:        make(map[string]*Table),
		logger:        logger,
		autoIncrement: make(map[string]int64),
	}
}

// CreateTable 创建表 Create table
func (tm *TableManager) CreateTable(ctx context.Context, schema *types.Schema, options map[string]interface{}) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	if err := tm.validateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 检查表是否已存在 Check if table already exists
	if _, exists := tm.tables[schema.Name]; exists {
		return errors.NewTableExistsError(schema.Name)
	}

	// 创建表统计信息 Create table statistics
	stats := &TableStatistics{
		RowCount:        0,
		ColumnCount:     len(schema.Columns),
		IndexCount:      0,
		ConstraintCount: 0,
		AvgRowSize:      0,
		DataSize:        0,
		IndexSize:       0,
		LastAnalyzed:    time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 创建表实例 Create table instance
	table := &Table{
		name:         schema.Name,
		database:     tm.database,
		schema:       schema,
		indexes:      make(map[string]*Index),
		constraints:  make(map[string]*Constraint),
		statistics:   stats,
		encoder:      tm.database.encoder,
		decoder:      tm.database.decoder,
		valueEncoder: tm.database.valueEncoder,
		logger:       tm.logger,
		sequence:     0,
	}

	// 保存表结构 Save table schema
	if err := table.saveSchema(); err != nil {
		return fmt.Errorf("failed to save table schema: %w", err)
	}

	// 保存表统计信息 Save table statistics
	if err := table.saveStatistics(); err != nil {
		return fmt.Errorf("failed to save table statistics: %w", err)
	}

	// 创建主键索引 Create primary key index
	if len(schema.PrimaryKey) > 0 {
		pkIndex := &Index{
			Name:       fmt.Sprintf("pk_%s", schema.Name),
			TableName:  schema.Name,
			Columns:    schema.PrimaryKey,
			Unique:     true,
			Primary:    true,
			Type:       "btree",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Properties: make(map[string]string),
			Statistics: &IndexStatistics{},
		}

		if err := table.createIndex(pkIndex); err != nil {
			return fmt.Errorf("failed to create primary key index: %w", err)
		}
	}

	// 创建唯一约束索引 Create unique constraint indexes
	for _, col := range schema.Columns {
		if col.Unique && !col.PrimaryKey {
			uniqueIndex := &Index{
				Name:       fmt.Sprintf("uk_%s_%s", schema.Name, col.Name),
				TableName:  schema.Name,
				Columns:    []string{col.Name},
				Unique:     true,
				Primary:    false,
				Type:       "btree",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				Properties: make(map[string]string),
				Statistics: &IndexStatistics{},
			}

			if err := table.createIndex(uniqueIndex); err != nil {
				return fmt.Errorf("failed to create unique index for column %s: %w", col.Name, err)
			}
		}
	}

	// 添加到管理器 Add to manager
	tm.tables[schema.Name] = table

	// 增加数据库表计数 Increment database table count
	if err := tm.database.IncrementTableCount(); err != nil {
		tm.logger.Warn("Failed to increment table count",
			"table", schema.Name,
			"error", err)
	}

	tm.logger.Info("Table created successfully",
		"table", schema.Name,
		"columns", len(schema.Columns))

	return nil
}

// DropTable 删除表 Drop table
func (tm *TableManager) DropTable(ctx context.Context, tableName string, cascade bool) error {
	if err := tm.validateTableName(tableName); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 检查表是否存在 Check if table exists
	table, exists := tm.tables[tableName]
	if !exists {
		return errors.NewTableNotFoundError(tableName)
	}

	// 删除所有数据 Delete all data
	if err := table.truncateData(ctx); err != nil {
		return fmt.Errorf("failed to delete table data: %w", err)
	}

	// 删除所有索引 Delete all indexes
	for _, index := range table.indexes {
		if err := table.dropIndex(index.Name); err != nil {
			tm.logger.Warn("Failed to drop index during table deletion",
				"table", tableName,
				"index", index.Name,
				"error", err)
		}
	}

	// 删除表结构和统计信息 Delete table schema and statistics
	if err := table.deleteMetadata(); err != nil {
		tm.logger.Warn("Failed to delete table metadata",
			"table", tableName,
			"error", err)
	}

	// 从管理器中移除 Remove from manager
	delete(tm.tables, tableName)
	delete(tm.autoIncrement, tableName)

	// 减少数据库表计数 Decrement database table count
	if err := tm.database.DecrementTableCount(); err != nil {
		tm.logger.Warn("Failed to decrement table count",
			"table", tableName,
			"error", err)
	}

	tm.logger.Info("Table dropped successfully",
		"table", tableName,
		"cascade", cascade)

	return nil
}

// GetTable 获取表 Get table
func (tm *TableManager) GetTable(tableName string) (*Table, error) {
	if err := tm.validateTableName(tableName); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	table, exists := tm.tables[tableName]
	if !exists {
		return nil, errors.NewTableNotFoundError(tableName)
	}

	return table, nil
}

// ListTables 列出表 List tables
func (tm *TableManager) ListTables(ctx context.Context) ([]*types.Schema, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	schemas := make([]*types.Schema, 0, len(tm.tables))
	for _, table := range tm.tables {
		table.mu.RLock()
		// 复制结构 Copy schema
		schema := *table.schema
		table.mu.RUnlock()
		schemas = append(schemas, &schema)
	}

	// 按名称排序 Sort by name
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Name < schemas[j].Name
	})

	return schemas, nil
}

// TableExists 检查表是否存在 Check if table exists
func (tm *TableManager) TableExists(tableName string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, exists := tm.tables[tableName]
	return exists
}

// AlterTable 修改表结构 Alter table
func (tm *TableManager) AlterTable(ctx context.Context, tableName string, changes *types.AlterTableSpec) error {
	table, err := tm.GetTable(tableName)
	if err != nil {
		return err
	}

	table.mu.Lock()
	defer table.mu.Unlock()

	// 处理添加列 Handle add columns
	for _, addCol := range changes.AddColumns {
		if err := table.addColumn(addCol); err != nil {
			return fmt.Errorf("failed to add column %s: %w", addCol.Name, err)
		}
	}

	// 处理删除列 Handle drop columns
	for _, dropCol := range changes.DropColumns {
		if err := table.dropColumn(dropCol); err != nil {
			return fmt.Errorf("failed to drop column %s: %w", dropCol, err)
		}
	}

	// 处理修改列 Handle modify columns
	for _, modifyCol := range changes.ModifyColumns {
		if err := table.modifyColumn(modifyCol); err != nil {
			return fmt.Errorf("failed to modify column %s: %w", modifyCol.Name, err)
		}
	}

	// 保存更新的结构 Save updated schema
	table.schema.Version++
	table.schema.UpdatedAt = time.Now()

	if err := table.saveSchema(); err != nil {
		return fmt.Errorf("failed to save updated schema: %w", err)
	}

	tm.logger.Info("Table altered successfully",
		"table", tableName,
		"changes", fmt.Sprintf("%+v", changes))

	return nil
}

// Insert 插入数据 Insert data
func (t *Table) Insert(ctx context.Context, row *types.Row) error {
	if row == nil {
		return fmt.Errorf("row cannot be nil")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 验证行数据 Validate row data
	if err := t.validateRow(row); err != nil {
		return fmt.Errorf("invalid row data: %w", err)
	}

	// 处理自增列 Handle auto increment columns
	if err := t.handleAutoIncrement(row); err != nil {
		return fmt.Errorf("failed to handle auto increment: %w", err)
	}

	// 生成行键 Generate row key
	rowKey, err := t.generateRowKey(row)
	if err != nil {
		return fmt.Errorf("failed to generate row key: %w", err)
	}

	// 检查主键冲突 Check primary key conflict
	if err := t.checkPrimaryKeyConflict(rowKey); err != nil {
		return err
	}

	// 检查唯一约束 Check unique constraints
	if err := t.checkUniqueConstraints(row); err != nil {
		return err
	}

	// 编码行数据 Encode row data
	rowData, err := t.valueEncoder.EncodeRow(row)
	if err != nil {
		return fmt.Errorf("failed to encode row data: %w", err)
	}

	// 开始事务 Start transaction
	return t.database.db.Update(func(txn *badger.Txn) error {
		// 插入行数据 Insert row data
		if err := txn.Set(rowKey, rowData); err != nil {
			return fmt.Errorf("failed to insert row data: %w", err)
		}

		// 更新索引 Update indexes
		for _, index := range t.indexes {
			if err := t.updateIndexForInsert(txn, index, row, rowKey); err != nil {
				return fmt.Errorf("failed to update index %s: %w", index.Name, err)
			}
		}

		// 更新统计信息 Update statistics
		t.statistics.RowCount++
		t.statistics.UpdatedAt = time.Now()

		return nil
	})
}

// Update 更新数据 Update data
func (t *Table) Update(ctx context.Context, rowKey []byte, updates map[string]*types.Value, conditions *WhereClause) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.database.db.Update(func(txn *badger.Txn) error {
		// 获取原始行数据 Get original row data
		item, err := txn.Get(rowKey)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return errors.NewRowNotFoundError()
			}
			return fmt.Errorf("failed to get row: %w", err)
		}

		var oldRowData []byte
		if err := item.Value(func(data []byte) error {
			oldRowData = append([]byte(nil), data...)
			return nil
		}); err != nil {
			return fmt.Errorf("failed to read row data: %w", err)
		}

		// 解码原始行 Decode original row
		oldRow, err := t.valueEncoder.DecodeRow(oldRowData)
		if err != nil {
			return fmt.Errorf("failed to decode row: %w", err)
		}

		// 检查条件 Check conditions
		if conditions != nil {
			if match, err := t.evaluateWhereClause(conditions, oldRow); err != nil {
				return fmt.Errorf("failed to evaluate conditions: %w", err)
			} else if !match {
				return nil // 条件不匹配，跳过更新 Condition not matched, skip update
			}
		}

		// 创建新行 Create new row
		newRow := &types.Row{
			Values:    make([]*types.Value, len(oldRow.Values)),
			Version:   oldRow.Version + 1,
			CreatedAt: oldRow.CreatedAt,
			UpdatedAt: time.Now(),
			Deleted:   false,
			Metadata:  oldRow.Metadata,
		}

		// 复制原始值 Copy original values
		copy(newRow.Values, oldRow.Values)

		// 应用更新 Apply updates
		for colName, newValue := range updates {
			colIndex := t.getColumnIndex(colName)
			if colIndex < 0 {
				return fmt.Errorf("column %s not found", colName)
			}
			newRow.Values[colIndex] = newValue
		}

		// 验证新行 Validate new row
		if err := t.validateRow(newRow); err != nil {
			return fmt.Errorf("invalid updated row: %w", err)
		}

		// 检查唯一约束 Check unique constraints
		if err := t.checkUniqueConstraintsForUpdate(newRow, rowKey); err != nil {
			return err
		}

		// 编码新行数据 Encode new row data
		newRowData, err := t.valueEncoder.EncodeRow(newRow)
		if err != nil {
			return fmt.Errorf("failed to encode new row: %w", err)
		}

		// 更新行数据 Update row data
		if err := txn.Set(rowKey, newRowData); err != nil {
			return fmt.Errorf("failed to update row data: %w", err)
		}

		// 更新索引 Update indexes
		for _, index := range t.indexes {
			if err := t.updateIndexForUpdate(txn, index, oldRow, newRow, rowKey); err != nil {
				return fmt.Errorf("failed to update index %s: %w", index.Name, err)
			}
		}

		// 更新统计信息 Update statistics
		t.statistics.UpdatedAt = time.Now()

		return nil
	})
}

// Delete 删除数据 Delete data
func (t *Table) Delete(ctx context.Context, conditions *WhereClause) (int64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var deletedCount int64

	return deletedCount, t.database.db.Update(func(txn *badger.Txn) error {
		// 创建迭代器 Create iterator
		prefix := t.encoder.EncodeRowPrefix(t.name)
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		// 收集要删除的键 Collect keys to delete
		keysToDelete := make([][]byte, 0)
		rowsToDelete := make([]*types.Row, 0)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)

			// 获取行数据 Get row data
			var rowData []byte
			if err := item.Value(func(data []byte) error {
				rowData = append([]byte(nil), data...)
				return nil
			}); err != nil {
				continue
			}

			// 解码行 Decode row
			row, err := t.valueEncoder.DecodeRow(rowData)
			if err != nil {
				continue
			}

			// 检查删除条件 Check delete conditions
			if conditions != nil {
				if match, err := t.evaluateWhereClause(conditions, row); err != nil {
					return fmt.Errorf("failed to evaluate conditions: %w", err)
				} else if !match {
					continue
				}
			}

			keysToDelete = append(keysToDelete, key)
			rowsToDelete = append(rowsToDelete, row)
		}

		// 执行删除 Execute deletion
		for i, key := range keysToDelete {
			row := rowsToDelete[i]

			// 删除行数据 Delete row data
			if err := txn.Delete(key); err != nil {
				return fmt.Errorf("failed to delete row: %w", err)
			}

			// 更新索引 Update indexes
			for _, index := range t.indexes {
				if err := t.updateIndexForDelete(txn, index, row, key); err != nil {
					return fmt.Errorf("failed to update index %s: %w", index.Name, err)
				}
			}

			deletedCount++
		}

		// 更新统计信息 Update statistics
		t.statistics.RowCount -= deletedCount
		t.statistics.UpdatedAt = time.Now()

		return nil
	})
}

// Select 查询数据 Select data
func (t *Table) Select(ctx context.Context, columns []string, options *QueryOptions) ([]*types.Row, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var rows []*types.Row

	err := t.database.db.View(func(txn *badger.Txn) error {
		// 选择查询策略 Choose query strategy
		if options != nil && options.Where != nil {
			// 尝试使用索引查询 Try to use index query
			if indexRows, err := t.selectWithIndex(txn, columns, options); err == nil {
				rows = indexRows
				return nil
			}
		}

		// 全表扫描 Full table scan
		return t.selectWithScan(txn, columns, options, &rows)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to select data: %w", err)
	}

	// 应用后处理 Apply post-processing
	if options != nil {
		rows = t.applyPostProcessing(rows, options)
	}

	return rows, nil
}

// selectWithIndex 使用索引查询 Select with index
func (t *Table) selectWithIndex(txn *badger.Txn, columns []string, options *QueryOptions) ([]*types.Row, error) {
	// 分析WHERE子句找到可用索引 Analyze WHERE clause to find usable index
	index, indexKey := t.findBestIndex(options.Where)
	if index == nil {
		return nil, fmt.Errorf("no suitable index found")
	}

	var rows []*types.Row

	// 使用索引键进行查询 Query using index key
	opts := badger.DefaultIteratorOptions
	it := txn.NewIterator(opts)
	defer it.Close()

	if indexKey != nil {
		// 精确匹配 Exact match
		item, err := txn.Get(indexKey)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return rows, nil
			}
			return nil, err
		}

		// 从索引获取行键 Get row key from index
		var rowKey []byte
		if err := item.Value(func(data []byte) error {
			rowKey = append([]byte(nil), data...)
			return nil
		}); err != nil {
			return nil, err
		}

		// 获取行数据 Get row data
		if row, err := t.getRowByKey(txn, rowKey); err == nil {
			rows = append(rows, row)
		}
	} else {
		// 范围查询 Range query
		prefix := t.encoder.EncodeIndexPrefix(t.name, index.Name)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			// 获取行键 Get row key
			var rowKey []byte
			if err := item.Value(func(data []byte) error {
				rowKey = append([]byte(nil), data...)
				return nil
			}); err != nil {
				continue
			}

			// 获取行数据 Get row data
			if row, err := t.getRowByKey(txn, rowKey); err == nil {
				// 检查WHERE条件 Check WHERE conditions
				if options.Where != nil {
					if match, err := t.evaluateWhereClause(options.Where, row); err != nil {
						continue
					} else if !match {
						continue
					}
				}
				rows = append(rows, row)
			}
		}
	}

	return rows, nil
}

// selectWithScan 全表扫描查询 Select with full table scan
func (t *Table) selectWithScan(txn *badger.Txn, columns []string, options *QueryOptions, rows *[]*types.Row) error {
	prefix := t.encoder.EncodeRowPrefix(t.name)
	opts := badger.DefaultIteratorOptions
	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()

		// 获取行数据 Get row data
		var rowData []byte
		if err := item.Value(func(data []byte) error {
			rowData = append([]byte(nil), data...)
			return nil
		}); err != nil {
			continue
		}

		// 解码行 Decode row
		row, err := t.valueEncoder.DecodeRow(rowData)
		if err != nil {
			continue
		}

		// 检查WHERE条件 Check WHERE conditions
		if options != nil && options.Where != nil {
			if match, err := t.evaluateWhereClause(options.Where, row); err != nil {
				continue
			} else if !match {
				continue
			}
		}

		// 选择指定列 Select specified columns
		if len(columns) > 0 && !t.containsAllColumns(columns) {
			row = t.selectColumns(row, columns)
		}

		*rows = append(*rows, row)
	}

	return nil
}

// applyPostProcessing 应用后处理 Apply post-processing
func (t *Table) applyPostProcessing(rows []*types.Row, options *QueryOptions) []*types.Row {
	// 应用排序 Apply ordering
	if len(options.OrderBy) > 0 {
		rows = t.applySorting(rows, options.OrderBy)
	}

	// 应用去重 Apply distinct
	if options.Distinct {
		rows = t.applyDistinct(rows)
	}

	// 应用偏移和限制 Apply offset and limit
	if options.Offset > 0 {
		if options.Offset >= len(rows) {
			return []*types.Row{}
		}
		rows = rows[options.Offset:]
	}

	if options.Limit > 0 && options.Limit < len(rows) {
		rows = rows[:options.Limit]
	}

	return rows
}

// Count 统计行数 Count rows
func (t *Table) Count(ctx context.Context, conditions *WhereClause) (int64, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var count int64

	err := t.database.db.View(func(txn *badger.Txn) error {
		prefix := t.encoder.EncodeRowPrefix(t.name)
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // 只需要键，不需要值 Only need keys, not values
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if conditions == nil {
				count++
				continue
			}

			// 需要检查条件时，获取行数据 When checking conditions, get row data
			item := it.Item()
			var rowData []byte
			if err := item.Value(func(data []byte) error {
				rowData = append([]byte(nil), data...)
				return nil
			}); err != nil {
				continue
			}

			// 解码行 Decode row
			row, err := t.valueEncoder.DecodeRow(rowData)
			if err != nil {
				continue
			}

			// 检查条件 Check conditions
			if match, err := t.evaluateWhereClause(conditions, row); err != nil {
				continue
			} else if match {
				count++
			}
		}

		return nil
	})

	return count, err
}

// Truncate 清空表数据 Truncate table data
func (t *Table) Truncate(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.truncateData(ctx)
}

// truncateData 清空表数据（内部方法） Truncate table data (internal method)
func (t *Table) truncateData(ctx context.Context) error {
	return t.database.db.Update(func(txn *badger.Txn) error {
		// 删除所有行数据 Delete all row data
		prefix := t.encoder.EncodeRowPrefix(t.name)
		if err := t.deleteKeysWithPrefix(txn, prefix); err != nil {
			return fmt.Errorf("failed to delete row data: %w", err)
		}

		// 删除所有索引数据 Delete all index data
		for _, index := range t.indexes {
			indexPrefix := t.encoder.EncodeIndexPrefix(t.name, index.Name)
			if err := t.deleteKeysWithPrefix(txn, indexPrefix); err != nil {
				return fmt.Errorf("failed to delete index data for %s: %w", index.Name, err)
			}
		}

		// 重置统计信息 Reset statistics
		t.statistics.RowCount = 0
		t.statistics.DataSize = 0
		t.statistics.UpdatedAt = time.Now()

		return t.saveStatistics()
	})
}

// CreateIndex 创建索引 Create index
func (t *Table) CreateIndex(ctx context.Context, index *Index) error {
	if index == nil {
		return fmt.Errorf("index cannot be nil")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	return t.createIndex(index)
}

// createIndex 创建索引（内部方法） Create index (internal method)
func (t *Table) createIndex(index *Index) error {
	// 验证索引 Validate index
	if err := t.validateIndex(index); err != nil {
		return fmt.Errorf("invalid index: %w", err)
	}

	// 检查索引是否已存在 Check if index already exists
	if _, exists := t.indexes[index.Name]; exists {
		return fmt.Errorf("index %s already exists", index.Name)
	}

	// 验证列是否存在 Validate columns exist
	for _, colName := range index.Columns {
		if t.getColumnIndex(colName) < 0 {
			return fmt.Errorf("column %s not found", colName)
		}
	}

	// 创建索引统计信息 Create index statistics
	index.Statistics = &IndexStatistics{
		KeyCount:   0,
		UniqueKeys: 0,
		AvgKeySize: 0,
		LastUsed:   time.Now(),
		UsageCount: 0,
	}

	// 为现有数据构建索引 Build index for existing data
	if err := t.buildIndexForExistingData(index); err != nil {
		return fmt.Errorf("failed to build index for existing data: %w", err)
	}

	// 保存索引定义 Save index definition
	if err := t.saveIndexDefinition(index); err != nil {
		return fmt.Errorf("failed to save index definition: %w", err)
	}

	// 添加到索引映射 Add to index mapping
	t.indexes[index.Name] = index
	t.statistics.IndexCount++
	t.statistics.UpdatedAt = time.Now()

	// 增加数据库索引计数 Increment database index count
	if err := t.database.IncrementIndexCount(); err != nil {
		t.logger.Warn("Failed to increment index count",
			"table", t.name,
			"index", index.Name,
			"error", err)
	}

	t.logger.Info("Index created successfully",
		"table", t.name,
		"index", index.Name,
		"columns", index.Columns)

	return nil
}

// DropIndex 删除索引 Drop index
func (t *Table) DropIndex(ctx context.Context, indexName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.dropIndex(indexName)
}

// dropIndex 删除索引（内部方法） Drop index (internal method)
func (t *Table) dropIndex(indexName string) error {
	// 检查索引是否存在 Check if index exists
	index, exists := t.indexes[indexName]
	if !exists {
		return fmt.Errorf("index %s not found", indexName)
	}

	// 不能删除主键索引 Cannot drop primary key index
	if index.Primary {
		return fmt.Errorf("cannot drop primary key index")
	}

	// 删除索引数据 Delete index data
	if err := t.database.db.Update(func(txn *badger.Txn) error {
		prefix := t.encoder.EncodeIndexPrefix(t.name, indexName)
		return t.deleteKeysWithPrefix(txn, prefix)
	}); err != nil {
		return fmt.Errorf("failed to delete index data: %w", err)
	}

	// 删除索引定义 Delete index definition
	if err := t.deleteIndexDefinition(indexName); err != nil {
		t.logger.Warn("Failed to delete index definition",
			"table", t.name,
			"index", indexName,
			"error", err)
	}

	// 从索引映射中移除 Remove from index mapping
	delete(t.indexes, indexName)
	t.statistics.IndexCount--
	t.statistics.UpdatedAt = time.Now()

	// 减少数据库索引计数 Decrement database index count
	if err := t.database.DecrementIndexCount(); err != nil {
		t.logger.Warn("Failed to decrement index count",
			"table", t.name,
			"index", indexName,
			"error", err)
	}

	t.logger.Info("Index dropped successfully",
		"table", t.name,
		"index", indexName)

	return nil
}

// ListIndexes 列出索引 List indexes
func (t *Table) ListIndexes(ctx context.Context) ([]*Index, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	indexes := make([]*Index, 0, len(t.indexes))
	for _, index := range t.indexes {
		// 复制索引 Copy index
		indexCopy := *index
		indexes = append(indexes, &indexCopy)
	}

	// 按名称排序 Sort by name
	sort.Slice(indexes, func(i, j int) bool {
		return indexes[i].Name < indexes[j].Name
	})

	return indexes, nil
}

// GetIndexStatistics 获取索引统计信息 Get index statistics
func (t *Table) GetIndexStatistics(indexName string) (*IndexStatistics, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	index, exists := t.indexes[indexName]
	if !exists {
		return nil, fmt.Errorf("index %s not found", indexName)
	}

	// 返回统计信息副本 Return copy of statistics
	stats := *index.Statistics
	return &stats, nil
}

// RefreshStatistics 刷新表统计信息 Refresh table statistics
func (t *Table) RefreshStatistics(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.database.db.View(func(txn *badger.Txn) error {
		var stats TableStatistics
		var totalRowSize int64
		var rowCount int64

		// 统计行数据 Count row data
		prefix := t.encoder.EncodeRowPrefix(t.name)
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			rowCount++
			totalRowSize += item.EstimatedSize()
		}

		// 计算平均行大小 Calculate average row size
		var avgRowSize float64
		if rowCount > 0 {
			avgRowSize = float64(totalRowSize) / float64(rowCount)
		}

		// 统计索引大小 Count index size
		var indexSize int64
		for _, index := range t.indexes {
			indexPrefix := t.encoder.EncodeIndexPrefix(t.name, index.Name)
			indexIt := txn.NewIterator(opts)

			for indexIt.Seek(indexPrefix); indexIt.ValidForPrefix(indexPrefix); indexIt.Next() {
				item := indexIt.Item()
				indexSize += item.EstimatedSize()
			}
			indexIt.Close()
		}

		// 更新统计信息 Update statistics
		stats.RowCount = rowCount
		stats.ColumnCount = len(t.schema.Columns)
		stats.IndexCount = len(t.indexes)
		stats.ConstraintCount = len(t.constraints)
		stats.AvgRowSize = avgRowSize
		stats.DataSize = totalRowSize
		stats.IndexSize = indexSize
		stats.LastAnalyzed = time.Now()
		stats.CreatedAt = t.statistics.CreatedAt
		stats.UpdatedAt = time.Now()

		t.statistics = &stats
		return t.saveStatistics()
	})
}

// 私有辅助方法 Private helper methods

// validateSchema 验证表结构 Validate schema
func (tm *TableManager) validateSchema(schema *types.Schema) error {
	if schema.Name == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	if len(schema.Name) > 64 {
		return fmt.Errorf("table name too long (max 64 characters)")
	}

	if len(schema.Columns) == 0 {
		return fmt.Errorf("table must have at least one column")
	}

	// 检查列名重复 Check for duplicate column names
	columnNames := make(map[string]bool)
	for _, col := range schema.Columns {
		if col.Name == "" {
			return fmt.Errorf("column name cannot be empty")
		}
		if columnNames[col.Name] {
			return fmt.Errorf("duplicate column name: %s", col.Name)
		}
		columnNames[col.Name] = true
	}

	// 验证主键 Validate primary key
	if len(schema.PrimaryKey) > 0 {
		for _, pkCol := range schema.PrimaryKey {
			if !columnNames[pkCol] {
				return fmt.Errorf("primary key column %s not found", pkCol)
			}
		}
	}

	return nil
}

// validateTableName 验证表名 Validate table name
func (tm *TableManager) validateTableName(tableName string) error {
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	if len(tableName) > 64 {
		return fmt.Errorf("table name too long (max 64 characters)")
	}

	// 检查表名格式 Check table name format
	for _, r := range tableName {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_') {
			return fmt.Errorf("invalid table name format")
		}
	}

	return nil
}

// validateRow 验证行数据 Validate row data
func (t *Table) validateRow(row *types.Row) error {
	if len(row.Values) != len(t.schema.Columns) {
		return fmt.Errorf("column count mismatch: expected %d, got %d",
			len(t.schema.Columns), len(row.Values))
	}

	// 验证每个列的值 Validate each column value
	for i, col := range t.schema.Columns {
		value := row.Values[i]

		// 检查非空约束 Check NOT NULL constraint
		if !col.Nullable && (value == nil || value.IsNull()) {
			return fmt.Errorf("column %s cannot be null", col.Name)
		}

		// 检查数据类型 Check data type
		if value != nil && !value.IsNull() {
			if err := t.validateColumnValue(col, value); err != nil {
				return fmt.Errorf("invalid value for column %s: %w", col.Name, err)
			}
		}
	}

	return nil
}

// validateColumnValue 验证列值 Validate column value
func (t *Table) validateColumnValue(col *types.Column, value *types.Value) error {
	// 检查数据类型兼容性 Check data type compatibility
	if !t.isTypeCompatible(col.Type, value.Type()) {
		return fmt.Errorf("type mismatch: expected %s, got %s",
			col.Type, value.Type())
	}

	// 检查长度限制 Check length constraints
	if col.Length > 0 {
		switch value.Type() {
		case types.StringType:
			if str, _ := value.ToString(); len(str) > col.Length {
				return fmt.Errorf("string too long: max %d characters", col.Length)
			}
		case types.BinaryType:
			if data, _ := value.ToBinary(); len(data) > col.Length {
				return fmt.Errorf("binary data too long: max %d bytes", col.Length)
			}
		}
	}

	return nil
}

// isTypeCompatible 检查类型兼容性 Check type compatibility
func (t *Table) isTypeCompatible(expected string, actual types.ValueType) bool {
	switch expected {
	case "INT", "INTEGER":
		return actual == types.IntType
	case "BIGINT":
		return actual == types.IntType
	case "FLOAT", "DOUBLE":
		return actual == types.FloatType
	case "STRING", "VARCHAR", "TEXT":
		return actual == types.StringType
	case "BOOLEAN", "BOOL":
		return actual == types.BoolType
	case "BINARY", "BLOB":
		return actual == types.BinaryType
	case "TIMESTAMP", "DATETIME":
		return actual == types.TimestampType
	default:
		return false
	}
}

// handleAutoIncrement 处理自增列 Handle auto increment columns
func (t *Table) handleAutoIncrement(row *types.Row) error {
	for i, col := range t.schema.Columns {
		if col.AutoIncrement {
			// 获取下一个自增值 Get next auto increment value
			t.sequence++
			autoValue := types.NewIntValue(t.sequence)
			row.Values[i] = autoValue
		}
	}
	return nil
}

// generateRowKey 生成行键 Generate row key
func (t *Table) generateRowKey(row *types.Row) ([]byte, error) {
	if len(t.schema.PrimaryKey) == 0 {
		// 如果没有主键，使用序列号 If no primary key, use sequence number
		t.sequence++
		return t.encoder.EncodeRowKey(t.name, [][]byte{
			[]byte(fmt.Sprintf("%d", t.sequence)),
		}), nil
	}

	// 使用主键值 Use primary key values
	pkValues := make([][]byte, len(t.schema.PrimaryKey))
	for i, pkCol := range t.schema.PrimaryKey {
		colIndex := t.getColumnIndex(pkCol)
		if colIndex < 0 {
			return nil, fmt.Errorf("primary key column %s not found", pkCol)
		}

		value := row.Values[colIndex]
		if value == nil || value.IsNull() {
			return nil, fmt.Errorf("primary key column %s cannot be null", pkCol)
		}

		// 编码主键值 Encode primary key value
		encoded, err := t.valueEncoder.EncodeValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to encode primary key value: %w", err)
		}
		pkValues[i] = encoded
	}

	return t.encoder.EncodeRowKey(t.name, pkValues), nil
}

// getColumnIndex 获取列索引 Get column index
func (t *Table) getColumnIndex(columnName string) int {
	for i, col := range t.schema.Columns {
		if col.Name == columnName {
			return i
		}
	}
	return -1
}

// checkPrimaryKeyConflict 检查主键冲突 Check primary key conflict
func (t *Table) checkPrimaryKeyConflict(rowKey []byte) error {
	return t.database.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(rowKey)
		if err == badger.ErrKeyNotFound {
			return nil // 没有冲突 No conflict
		}
		if err != nil {
			return fmt.Errorf("failed to check primary key: %w", err)
		}
		return errors.NewPrimaryKeyViolationError()
	})
}

// checkUniqueConstraints 检查唯一约束 Check unique constraints
func (t *Table) checkUniqueConstraints(row *types.Row) error {
	for _, index := range t.indexes {
		if !index.Unique {
			continue
		}

		// 生成索引键 Generate index key
		indexKey, err := t.generateIndexKey(index, row)
		if err != nil {
			return fmt.Errorf("failed to generate index key: %w", err)
		}

		// 检查是否已存在 Check if already exists
		err = t.database.db.View(func(txn *badger.Txn) error {
			_, err := txn.Get(indexKey)
			if err == badger.ErrKeyNotFound {
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to check unique constraint: %w", err)
			}
			return errors.NewUniqueViolationError(index.Name)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// checkUniqueConstraintsForUpdate 检查更新时的唯一约束 Check unique constraints for update
func (t *Table) checkUniqueConstraintsForUpdate(row *types.Row, excludeRowKey []byte) error {
	for _, index := range t.indexes {
		if !index.Unique {
			continue
		}

		// 生成索引键 Generate index key
		indexKey, err := t.generateIndexKey(index, row)
		if err != nil {
			return fmt.Errorf("failed to generate index key: %w", err)
		}

		// 检查是否已存在（排除当前行） Check if exists (exclude current row)
		err = t.database.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(indexKey)
			if err == badger.ErrKeyNotFound {
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to check unique constraint: %w", err)
			}

			// 获取索引指向的行键 Get row key from index
			var existingRowKey []byte
			if err := item.Value(func(data []byte) error {
				existingRowKey = append([]byte(nil), data...)
				return nil
			}); err != nil {
				return err
			}

			// 如果指向的是当前行，则不算冲突 If pointing to current row, no conflict
			if string(existingRowKey) == string(excludeRowKey) {
				return nil
			}

			return errors.NewUniqueViolationError(index.Name)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// generateIndexKey 生成索引键 Generate index key
func (t *Table) generateIndexKey(index *Index, row *types.Row) ([]byte, error) {
	values := make([][]byte, len(index.Columns))

	for i, colName := range index.Columns {
		colIndex := t.getColumnIndex(colName)
		if colIndex < 0 {
			return nil, fmt.Errorf("index column %s not found", colName)
		}

		value := row.Values[colIndex]
		if value == nil || value.IsNull() {
			values[i] = []byte{} // 空值用空字节表示 Empty value represented by empty bytes
		} else {
			encoded, err := t.valueEncoder.EncodeValue(value)
			if err != nil {
				return nil, fmt.Errorf("failed to encode index value: %w", err)
			}
			values[i] = encoded
		}
	}

	return t.encoder.EncodeIndexKey(t.name, index.Name, values), nil
}

// updateIndexForInsert 为插入更新索引 Update index for insert
func (t *Table) updateIndexForInsert(txn *badger.Txn, index *Index, row *types.Row, rowKey []byte) error {
	// 生成索引键 Generate index key
	indexKey, err := t.generateIndexKey(index, row)
	if err != nil {
		return err
	}

	// 插入索引条目 Insert index entry
	if err := txn.Set(indexKey, rowKey); err != nil {
		return err
	}

	// 更新索引统计信息 Update index statistics
	index.Statistics.KeyCount++
	index.Statistics.LastUsed = time.Now()

	return nil
}

// updateIndexForUpdate 为更新更新索引 Update index for update
func (t *Table) updateIndexForUpdate(txn *badger.Txn, index *Index, oldRow, newRow *types.Row, rowKey []byte) error {
	// 删除旧索引条目 Delete old index entry
	oldIndexKey, err := t.generateIndexKey(index, oldRow)
	if err != nil {
		return err
	}
	if err := txn.Delete(oldIndexKey); err != nil {
		return err
	}

	// 插入新索引条目 Insert new index entry
	newIndexKey, err := t.generateIndexKey(index, newRow)
	if err != nil {
		return err
	}
	if err := txn.Set(newIndexKey, rowKey); err != nil {
		return err
	}

	// 更新索引统计信息 Update index statistics
	index.Statistics.LastUsed = time.Now()

	return nil
}

// updateIndexForDelete 为删除更新索引 Update index for delete
func (t *Table) updateIndexForDelete(txn *badger.Txn, index *Index, row *types.Row, rowKey []byte) error {
	// 生成索引键 Generate index key
	indexKey, err := t.generateIndexKey(index, row)
	if err != nil {
		return err
	}

	// 删除索引条目 Delete index entry
	if err := txn.Delete(indexKey); err != nil {
		return err
	}

	// 更新索引统计信息 Update index statistics
	index.Statistics.KeyCount--

	return nil
}

// evaluateWhereClause 评估WHERE条件 Evaluate WHERE clause
func (t *Table) evaluateWhereClause(clause *WhereClause, row *types.Row) (bool, error) {
	if clause == nil {
		return true, nil
	}

	switch clause.Operator {
	case "AND":
		if clause.Left == nil || clause.Right == nil {
			return false, fmt.Errorf("AND operator requires left and right conditions")
		}
		leftResult, err := t.evaluateWhereClause(clause.Left, row)
		if err != nil {
			return false, err
		}
		if !leftResult {
			return false, nil // 短路求值 Short-circuit evaluation
		}
		return t.evaluateWhereClause(clause.Right, row)

	case "OR":
		if clause.Left == nil || clause.Right == nil {
			return false, fmt.Errorf("OR operator requires left and right conditions")
		}
		leftResult, err := t.evaluateWhereClause(clause.Left, row)
		if err != nil {
			return false, err
		}
		if leftResult {
			return true, nil // 短路求值 Short-circuit evaluation
		}
		return t.evaluateWhereClause(clause.Right, row)

	case "=", "EQ":
		return t.evaluateComparison(clause.Column, clause.Value, row, func(cmp int) bool { return cmp == 0 })

	case "!=", "<>", "NE":
		return t.evaluateComparison(clause.Column, clause.Value, row, func(cmp int) bool { return cmp != 0 })

	case "<", "LT":
		return t.evaluateComparison(clause.Column, clause.Value, row, func(cmp int) bool { return cmp < 0 })

	case "<=", "LE":
		return t.evaluateComparison(clause.Column, clause.Value, row, func(cmp int) bool { return cmp <= 0 })

	case ">", "GT":
		return t.evaluateComparison(clause.Column, clause.Value, row, func(cmp int) bool { return cmp > 0 })

	case ">=", "GE":
		return t.evaluateComparison(clause.Column, clause.Value, row, func(cmp int) bool { return cmp >= 0 })

	case "IN":
		return t.evaluateIn(clause.Column, clause.Values, row)

	case "NOT IN":
		result, err := t.evaluateIn(clause.Column, clause.Values, row)
		return !result, err

	case "LIKE":
		return t.evaluateLike(clause.Column, clause.Value, row)

	case "NOT LIKE":
		result, err := t.evaluateLike(clause.Column, clause.Value, row)
		return !result, err

	case "IS NULL":
		return t.evaluateIsNull(clause.Column, row, true)

	case "IS NOT NULL":
		return t.evaluateIsNull(clause.Column, row, false)

	default:
		return false, fmt.Errorf("unsupported operator: %s", clause.Operator)
	}
}

// evaluateComparison 评估比较操作 Evaluate comparison operation
func (t *Table) evaluateComparison(columnName string, value *types.Value, row *types.Row, compareFn func(int) bool) (bool, error) {
	colIndex := t.getColumnIndex(columnName)
	if colIndex < 0 {
		return false, fmt.Errorf("column %s not found", columnName)
	}

	rowValue := row.Values[colIndex]
	if rowValue == nil || rowValue.IsNull() || value == nil || value.IsNull() {
		return false, nil // NULL值比较结果为false NULL comparison results in false
	}

	cmp, err := rowValue.Compare(value)
	if err != nil {
		return false, fmt.Errorf("failed to compare values: %w", err)
	}

	return compareFn(cmp), nil
}

// evaluateIn 评估IN操作 Evaluate IN operation
func (t *Table) evaluateIn(columnName string, values []*types.Value, row *types.Row) (bool, error) {
	colIndex := t.getColumnIndex(columnName)
	if colIndex < 0 {
		return false, fmt.Errorf("column %s not found", columnName)
	}

	rowValue := row.Values[colIndex]
	if rowValue == nil || rowValue.IsNull() {
		return false, nil
	}

	for _, value := range values {
		if value == nil || value.IsNull() {
			continue
		}

		cmp, err := rowValue.Compare(value)
		if err != nil {
			continue
		}

		if cmp == 0 {
			return true, nil
		}
	}

	return false, nil
}

// evaluateLike 评估LIKE操作 Evaluate LIKE operation
func (t *Table) evaluateLike(columnName string, pattern *types.Value, row *types.Row) (bool, error) {
	colIndex := t.getColumnIndex(columnName)
	if colIndex < 0 {
		return false, fmt.Errorf("column %s not found", columnName)
	}

	rowValue := row.Values[colIndex]
	if rowValue == nil || rowValue.IsNull() || pattern == nil || pattern.IsNull() {
		return false, nil
	}

	// 转换为字符串进行匹配 Convert to string for matching
	rowStr, err := rowValue.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to convert row value to string: %w", err)
	}

	patternStr, err := pattern.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to convert pattern to string: %w", err)
	}

	// 简单的LIKE实现（支持%和_通配符） Simple LIKE implementation (supports % and _ wildcards)
	return t.matchPattern(rowStr, patternStr), nil
}

// evaluateIsNull 评估IS NULL操作 Evaluate IS NULL operation
func (t *Table) evaluateIsNull(columnName string, row *types.Row, expectNull bool) (bool, error) {
	colIndex := t.getColumnIndex(columnName)
	if colIndex < 0 {
		return false, fmt.Errorf("column %s not found", columnName)
	}

	rowValue := row.Values[colIndex]
	isNull := rowValue == nil || rowValue.IsNull()

	return isNull == expectNull, nil
}

// matchPattern 模式匹配（LIKE操作） Pattern matching (LIKE operation)
func (t *Table) matchPattern(text, pattern string) bool {
	// 简化版本的LIKE匹配 Simplified LIKE matching
	// %: 匹配任意数量的字符 matches any number of characters
	// _: 匹配单个字符 matches a single character

	textRunes := []rune(text)
	patternRunes := []rune(pattern)

	return t.matchPatternRecursive(textRunes, patternRunes, 0, 0)
}

// matchPatternRecursive 递归模式匹配 Recursive pattern matching
func (t *Table) matchPatternRecursive(text, pattern []rune, textIdx, patternIdx int) bool {
	// 如果模式匹配完成 If pattern is fully matched
	if patternIdx == len(pattern) {
		return textIdx == len(text)
	}

	// 如果文本匹配完成但模式还有非%字符 If text is fully matched but pattern has non-% characters
	if textIdx == len(text) {
		for i := patternIdx; i < len(pattern); i++ {
			if pattern[i] != '%' {
				return false
			}
		}
		return true
	}

	// 处理%通配符 Handle % wildcard
	if pattern[patternIdx] == '%' {
		// 跳过连续的% Skip consecutive %
		for patternIdx < len(pattern) && pattern[patternIdx] == '%' {
			patternIdx++
		}

		// 如果%是最后的字符 If % is the last character
		if patternIdx == len(pattern) {
			return true
		}

		// 尝试从当前位置到文本结尾的每个位置 Try every position from current to end of text
		for i := textIdx; i <= len(text); i++ {
			if t.matchPatternRecursive(text, pattern, i, patternIdx) {
				return true
			}
		}
		return false
	}

	// 处理_通配符或精确匹配 Handle _ wildcard or exact match
	if pattern[patternIdx] == '_' || pattern[patternIdx] == text[textIdx] {
		return t.matchPatternRecursive(text, pattern, textIdx+1, patternIdx+1)
	}

	return false
}

// findBestIndex 找到最佳索引 Find best index
func (t *Table) findBestIndex(clause *WhereClause) (*Index, []byte) {
	if clause == nil {
		return nil, nil
	}

	// 查找等值条件的索引 Look for index on equality conditions
	if clause.Operator == "=" || clause.Operator == "EQ" {
		for _, index := range t.indexes {
			if len(index.Columns) == 1 && index.Columns[0] == clause.Column {
				// 尝试生成索引键 Try to generate index key
				if indexKey := t.tryGenerateIndexKeyForValue(index, clause.Column, clause.Value); indexKey != nil {
					// 更新索引使用统计 Update index usage statistics
					index.Statistics.UsageCount++
					index.Statistics.LastUsed = time.Now()
					return index, indexKey
				}
			}
		}
	}

	// 查找复合条件的索引 Look for index on compound conditions
	if clause.Operator == "AND" {
		// 收集所有等值条件 Collect all equality conditions
		conditions := t.collectEqualityConditions(clause)

		for _, index := range t.indexes {
			if t.indexMatchesConditions(index, conditions) {
				// 更新索引使用统计 Update index usage statistics
				index.Statistics.UsageCount++
				index.Statistics.LastUsed = time.Now()
				return index, nil // 返回nil表示需要范围查询 Return nil means range query needed
			}
		}
	}

	return nil, nil
}

// collectEqualityConditions 收集等值条件 Collect equality conditions
func (t *Table) collectEqualityConditions(clause *WhereClause) map[string]*types.Value {
	conditions := make(map[string]*types.Value)

	if clause.Operator == "=" || clause.Operator == "EQ" {
		conditions[clause.Column] = clause.Value
		return conditions
	}

	if clause.Operator == "AND" {
		if clause.Left != nil {
			for k, v := range t.collectEqualityConditions(clause.Left) {
				conditions[k] = v
			}
		}
		if clause.Right != nil {
			for k, v := range t.collectEqualityConditions(clause.Right) {
				conditions[k] = v
			}
		}
	}

	return conditions
}

// indexMatchesConditions 检查索引是否匹配条件 Check if index matches conditions
func (t *Table) indexMatchesConditions(index *Index, conditions map[string]*types.Value) bool {
	// 检查索引的所有列是否都有对应的等值条件 Check if all index columns have corresponding equality conditions
	for _, col := range index.Columns {
		if _, exists := conditions[col]; !exists {
			return false
		}
	}
	return true
}

// tryGenerateIndexKeyForValue 尝试为值生成索引键 Try to generate index key for value
func (t *Table) tryGenerateIndexKeyForValue(index *Index, column string, value *types.Value) []byte {
	if value == nil || value.IsNull() {
		return nil
	}

	// 创建虚拟行用于生成索引键 Create dummy row for generating index key
	row := &types.Row{
		Values: make([]*types.Value, len(t.schema.Columns)),
	}

	// 填充指定列的值 Fill the specified column value
	colIndex := t.getColumnIndex(column)
	if colIndex < 0 {
		return nil
	}
	row.Values[colIndex] = value

	// 生成索引键 Generate index key
	indexKey, err := t.generateIndexKey(index, row)
	if err != nil {
		return nil
	}

	return indexKey
}

// getRowByKey 通过键获取行 Get row by key
func (t *Table) getRowByKey(txn *badger.Txn, rowKey []byte) (*types.Row, error) {
	item, err := txn.Get(rowKey)
	if err != nil {
		return nil, err
	}

	var rowData []byte
	if err := item.Value(func(data []byte) error {
		rowData = append([]byte(nil), data...)
		return nil
	}); err != nil {
		return nil, err
	}

	return t.valueEncoder.DecodeRow(rowData)
}

// containsAllColumns 检查是否包含所有列 Check if contains all columns
func (t *Table) containsAllColumns(columns []string) bool {
	if len(columns) == 0 {
		return true // 空列表表示选择所有列 Empty list means select all columns
	}

	for _, col := range columns {
		if col == "*" {
			return true
		}
	}

	return len(columns) == len(t.schema.Columns)
}

// selectColumns 选择指定列 Select specified columns
func (t *Table) selectColumns(row *types.Row, columns []string) *types.Row {
	if len(columns) == 0 || (len(columns) == 1 && columns[0] == "*") {
		return row // 返回所有列 Return all columns
	}

	newRow := &types.Row{
		Values:    make([]*types.Value, len(columns)),
		Version:   row.Version,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Deleted:   row.Deleted,
		Metadata:  row.Metadata,
	}

	for i, colName := range columns {
		colIndex := t.getColumnIndex(colName)
		if colIndex >= 0 {
			newRow.Values[i] = row.Values[colIndex]
		} else {
			newRow.Values[i] = types.NewNullValue()
		}
	}

	return newRow
}

// applySorting 应用排序 Apply sorting
func (t *Table) applySorting(rows []*types.Row, orderBy []OrderByClause) []*types.Row {
	if len(orderBy) == 0 {
		return rows
	}

	sort.Slice(rows, func(i, j int) bool {
		return t.compareRows(rows[i], rows[j], orderBy) < 0
	})

	return rows
}

// compareRows 比较行 Compare rows
func (t *Table) compareRows(row1, row2 *types.Row, orderBy []OrderByClause) int {
	for _, clause := range orderBy {
		colIndex := t.getColumnIndex(clause.Column)
		if colIndex < 0 {
			continue
		}

		val1 := row1.Values[colIndex]
		val2 := row2.Values[colIndex]

		var cmp int
		if val1 == nil || val1.IsNull() {
			if val2 == nil || val2.IsNull() {
				cmp = 0
			} else {
				cmp = -1 // NULL值排在前面 NULL values come first
			}
		} else if val2 == nil || val2.IsNull() {
			cmp = 1
		} else {
			var err error
			cmp, err = val1.Compare(val2)
			if err != nil {
				continue
			}
		}

		if cmp != 0 {
			if strings.ToUpper(clause.Direction) == "DESC" {
				return -cmp
			}
			return cmp
		}
	}

	return 0
}

// applyDistinct 应用去重 Apply distinct
func (t *Table) applyDistinct(rows []*types.Row) []*types.Row {
	if len(rows) <= 1 {
		return rows
	}

	seen := make(map[string]bool)
	result := make([]*types.Row, 0, len(rows))

	for _, row := range rows {
		// 生成行的哈希值 Generate hash for row
		hash := t.generateRowHash(row)
		if !seen[hash] {
			seen[hash] = true
			result = append(result, row)
		}
	}

	return result
}

// generateRowHash 生成行哈希 Generate row hash
func (t *Table) generateRowHash(row *types.Row) string {
	var parts []string
	for _, value := range row.Values {
		if value == nil || value.IsNull() {
			parts = append(parts, "NULL")
		} else {
			str, _ := value.ToString()
			parts = append(parts, str)
		}
	}
	return strings.Join(parts, "|")
}

// validateIndex 验证索引 Validate index
func (t *Table) validateIndex(index *Index) error {
	if index.Name == "" {
		return fmt.Errorf("index name cannot be empty")
	}

	if len(index.Name) > 64 {
		return fmt.Errorf("index name too long (max 64 characters)")
	}

	if len(index.Columns) == 0 {
		return fmt.Errorf("index must have at least one column")
	}

	if len(index.Columns) > 16 {
		return fmt.Errorf("too many columns in index (max 16)")
	}

	// 检查列名重复 Check for duplicate column names
	columnSet := make(map[string]bool)
	for _, col := range index.Columns {
		if columnSet[col] {
			return fmt.Errorf("duplicate column in index: %s", col)
		}
		columnSet[col] = true
	}

	return nil
}

// buildIndexForExistingData 为现有数据构建索引 Build index for existing data
func (t *Table) buildIndexForExistingData(index *Index) error {
	return t.database.db.Update(func(txn *badger.Txn) error {
		// 扫描所有行数据 Scan all row data
		prefix := t.encoder.EncodeRowPrefix(t.name)
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			rowKey := item.KeyCopy(nil)

			// 获取行数据 Get row data
			var rowData []byte
			if err := item.Value(func(data []byte) error {
				rowData = append([]byte(nil), data...)
				return nil
			}); err != nil {
				continue
			}

			// 解码行 Decode row
			row, err := t.valueEncoder.DecodeRow(rowData)
			if err != nil {
				continue
			}

			// 为该行添加索引条目 Add index entry for this row
			if err := t.updateIndexForInsert(txn, index, row, rowKey); err != nil {
				return fmt.Errorf("failed to build index entry: %w", err)
			}
		}

		return nil
	})
}

// deleteKeysWithPrefix 删除指定前缀的所有键 Delete all keys with prefix
func (t *Table) deleteKeysWithPrefix(txn *badger.Txn, prefix []byte) error {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false // 只需要键 Only need keys
	it := txn.NewIterator(opts)
	defer it.Close()

	keysToDelete := make([][]byte, 0)

	// 收集要删除的键 Collect keys to delete
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		key := it.Item().KeyCopy(nil)
		keysToDelete = append(keysToDelete, key)
	}

	// 删除键 Delete keys
	for _, key := range keysToDelete {
		if err := txn.Delete(key); err != nil {
			return err
		}
	}

	return nil
}

// addColumn 添加列 Add column
func (t *Table) addColumn(column *types.Column) error {
	// 检查列名是否已存在 Check if column name already exists
	for _, col := range t.schema.Columns {
		if col.Name == column.Name {
			return fmt.Errorf("column %s already exists", column.Name)
		}
	}

	// 添加列到结构 Add column to schema
	t.schema.Columns = append(t.schema.Columns, column)

	// 如果表有数据，需要为所有现有行添加默认值 If table has data, add default value for all existing rows
	if t.statistics.RowCount > 0 {
		if err := t.addDefaultValueToExistingRows(column); err != nil {
			return fmt.Errorf("failed to add default value to existing rows: %w", err)
		}
	}

	// 更新统计信息 Update statistics
	t.statistics.ColumnCount++
	t.statistics.UpdatedAt = time.Now()

	return nil
}

// dropColumn 删除列 Drop column
func (t *Table) dropColumn(columnName string) error {
	// 查找列索引 Find column index
	colIndex := t.getColumnIndex(columnName)
	if colIndex < 0 {
		return fmt.Errorf("column %s not found", columnName)
	}

	column := t.schema.Columns[colIndex]

	// 检查是否为主键列 Check if it's a primary key column
	for _, pkCol := range t.schema.PrimaryKey {
		if pkCol == columnName {
			return fmt.Errorf("cannot drop primary key column %s", columnName)
		}
	}

	// 删除相关索引 Drop related indexes
	indexesToDrop := make([]string, 0)
	for _, index := range t.indexes {
		for _, indexCol := range index.Columns {
			if indexCol == columnName {
				indexesToDrop = append(indexesToDrop, index.Name)
				break
			}
		}
	}

	for _, indexName := range indexesToDrop {
		if err := t.dropIndex(indexName); err != nil {
			t.logger.Warn("Failed to drop index when dropping column",
				"column", columnName,
				"index", indexName,
				"error", err)
		}
	}

	// 从所有现有行中删除该列的值 Remove column value from all existing rows
	if t.statistics.RowCount > 0 {
		if err := t.removeColumnFromExistingRows(colIndex); err != nil {
			return fmt.Errorf("failed to remove column from existing rows: %w", err)
		}
	}

	// 从结构中删除列 Remove column from schema
	newColumns := make([]*types.Column, 0, len(t.schema.Columns)-1)
	for i, col := range t.schema.Columns {
		if i != colIndex {
			newColumns = append(newColumns, col)
		}
	}
	t.schema.Columns = newColumns

	// 更新统计信息 Update statistics
	t.statistics.ColumnCount--
	t.statistics.UpdatedAt = time.Now()

	t.logger.Info("Column dropped successfully",
		"table", t.name,
		"column", columnName)

	return nil
}

// modifyColumn 修改列 Modify column
func (t *Table) modifyColumn(column *types.Column) error {
	// 查找列索引 Find column index
	colIndex := t.getColumnIndex(column.Name)
	if colIndex < 0 {
		return fmt.Errorf("column %s not found", column.Name)
	}

	oldColumn := t.schema.Columns[colIndex]

	// 检查类型兼容性 Check type compatibility
	if !t.isTypeChangeCompatible(oldColumn.Type, column.Type) {
		return fmt.Errorf("incompatible type change from %s to %s", oldColumn.Type, column.Type)
	}

	// 如果表有数据，需要转换现有数据 If table has data, convert existing data
	if t.statistics.RowCount > 0 {
		if err := t.convertColumnDataType(colIndex, oldColumn, column); err != nil {
			return fmt.Errorf("failed to convert column data type: %w", err)
		}
	}

	// 更新列定义 Update column definition
	t.schema.Columns[colIndex] = column

	// 更新统计信息 Update statistics
	t.statistics.UpdatedAt = time.Now()

	t.logger.Info("Column modified successfully",
		"table", t.name,
		"column", column.Name,
		"old_type", oldColumn.Type,
		"new_type", column.Type)

	return nil
}

// addDefaultValueToExistingRows 为现有行添加默认值 Add default value to existing rows
func (t *Table) addDefaultValueToExistingRows(column *types.Column) error {
	// 获取默认值 Get default value
	var defaultValue *types.Value
	if column.DefaultValue != "" {
		// 解析默认值 Parse default value
		var err error
		defaultValue, err = t.parseDefaultValue(column.DefaultValue, column.Type)
		if err != nil {
			return fmt.Errorf("failed to parse default value: %w", err)
		}
	} else if column.Nullable {
		defaultValue = types.NewNullValue()
	} else {
		return fmt.Errorf("column %s is not nullable and has no default value", column.Name)
	}

	// 更新所有现有行 Update all existing rows
	return t.database.db.Update(func(txn *badger.Txn) error {
		prefix := t.encoder.EncodeRowPrefix(t.name)
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			rowKey := item.KeyCopy(nil)

			// 获取行数据 Get row data
			var rowData []byte
			if err := item.Value(func(data []byte) error {
				rowData = append([]byte(nil), data...)
				return nil
			}); err != nil {
				continue
			}

			// 解码行 Decode row
			row, err := t.valueEncoder.DecodeRow(rowData)
			if err != nil {
				continue
			}

			// 添加新列的值 Add new column value
			row.Values = append(row.Values, defaultValue)
			row.Version++
			row.UpdatedAt = time.Now()

			// 重新编码并保存 Re-encode and save
			newRowData, err := t.valueEncoder.EncodeRow(row)
			if err != nil {
				continue
			}

			if err := txn.Set(rowKey, newRowData); err != nil {
				return err
			}
		}

		return nil
	})
}

// removeColumnFromExistingRows 从现有行中删除列 Remove column from existing rows
func (t *Table) removeColumnFromExistingRows(colIndex int) error {
	return t.database.db.Update(func(txn *badger.Txn) error {
		prefix := t.encoder.EncodeRowPrefix(t.name)
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			rowKey := item.KeyCopy(nil)

			// 获取行数据 Get row data
			var rowData []byte
			if err := item.Value(func(data []byte) error {
				rowData = append([]byte(nil), data...)
				return nil
			}); err != nil {
				continue
			}

			// 解码行 Decode row
			row, err := t.valueEncoder.DecodeRow(rowData)
			if err != nil {
				continue
			}

			// 删除指定列的值 Remove specified column value
			newValues := make([]*types.Value, 0, len(row.Values)-1)
			for i, value := range row.Values {
				if i != colIndex {
					newValues = append(newValues, value)
				}
			}
			row.Values = newValues
			row.Version++
			row.UpdatedAt = time.Now()

			// 重新编码并保存 Re-encode and save
			newRowData, err := t.valueEncoder.EncodeRow(row)
			if err != nil {
				continue
			}

			if err := txn.Set(rowKey, newRowData); err != nil {
				return err
			}
		}

		return nil
	})
}

// convertColumnDataType 转换列数据类型 Convert column data type
func (t *Table) convertColumnDataType(colIndex int, oldColumn, newColumn *types.Column) error {
	return t.database.db.Update(func(txn *badger.Txn) error {
		prefix := t.encoder.EncodeRowPrefix(t.name)
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			rowKey := item.KeyCopy(nil)

			// 获取行数据 Get row data
			var rowData []byte
			if err := item.Value(func(data []byte) error {
				rowData = append([]byte(nil), data...)
				return nil
			}); err != nil {
				continue
			}

			// 解码行 Decode row
			row, err := t.valueEncoder.DecodeRow(rowData)
			if err != nil {
				continue
			}

			// 转换列值 Convert column value
			if colIndex < len(row.Values) && row.Values[colIndex] != nil && !row.Values[colIndex].IsNull() {
				convertedValue, err := t.convertValue(row.Values[colIndex], oldColumn.Type, newColumn.Type)
				if err != nil {
					t.logger.Warn("Failed to convert column value",
						"table", t.name,
						"column", newColumn.Name,
						"row_key", string(rowKey),
						"error", err)
					continue
				}
				row.Values[colIndex] = convertedValue
				row.Version++
				row.UpdatedAt = time.Now()

				// 重新编码并保存 Re-encode and save
				newRowData, err := t.valueEncoder.EncodeRow(row)
				if err != nil {
					continue
				}

				if err := txn.Set(rowKey, newRowData); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// isTypeChangeCompatible 检查类型变更兼容性 Check type change compatibility
func (t *Table) isTypeChangeCompatible(oldType, newType string) bool {
	// 定义兼容的类型转换 Define compatible type conversions
	compatibleChanges := map[string][]string{
		"INT":       {"BIGINT", "FLOAT", "STRING"},
		"BIGINT":    {"FLOAT", "STRING"},
		"FLOAT":     {"STRING"},
		"STRING":    {"TEXT"},
		"VARCHAR":   {"TEXT", "STRING"},
		"BOOLEAN":   {"STRING"},
		"TIMESTAMP": {"STRING"},
	}

	if oldType == newType {
		return true
	}

	compatible, exists := compatibleChanges[oldType]
	if !exists {
		return false
	}

	for _, compatibleType := range compatible {
		if compatibleType == newType {
			return true
		}
	}

	return false
}

// convertValue 转换值类型 Convert value type
func (t *Table) convertValue(value *types.Value, oldType, newType string) (*types.Value, error) {
	if oldType == newType {
		return value, nil
	}

	switch newType {
	case "STRING", "VARCHAR", "TEXT":
		str, err := value.ToString()
		if err != nil {
			return nil, err
		}
		return types.NewStringValue(str), nil

	case "BIGINT":
		if oldType == "INT" {
			intVal, err := value.ToInt()
			if err != nil {
				return nil, err
			}
			return types.NewIntValue(intVal), nil
		}

	case "FLOAT", "DOUBLE":
		switch oldType {
		case "INT", "BIGINT":
			intVal, err := value.ToInt()
			if err != nil {
				return nil, err
			}
			return types.NewFloatValue(float64(intVal)), nil
		}
	}

	return nil, fmt.Errorf("unsupported type conversion from %s to %s", oldType, newType)
}

// parseDefaultValue 解析默认值 Parse default value
func (t *Table) parseDefaultValue(defaultValue, columnType string) (*types.Value, error) {
	switch columnType {
	case "INT", "INTEGER", "BIGINT":
		// 尝试解析为整数 Try to parse as integer
		if intVal := parseInt(defaultValue); intVal != nil {
			return intVal, nil
		}
		return nil, fmt.Errorf("invalid integer default value: %s", defaultValue)

	case "FLOAT", "DOUBLE":
		// 尝试解析为浮点数 Try to parse as float
		if floatVal := parseFloat(defaultValue); floatVal != nil {
			return floatVal, nil
		}
		return nil, fmt.Errorf("invalid float default value: %s", defaultValue)

	case "BOOLEAN", "BOOL":
		// 尝试解析为布尔值 Try to parse as boolean
		if boolVal := parseBool(defaultValue); boolVal != nil {
			return boolVal, nil
		}
		return nil, fmt.Errorf("invalid boolean default value: %s", defaultValue)

	case "STRING", "VARCHAR", "TEXT":
		// 字符串类型直接返回 String type return directly
		return types.NewStringValue(defaultValue), nil

	case "TIMESTAMP", "DATETIME":
		// 尝试解析为时间戳 Try to parse as timestamp
		if timeVal := parseTimestamp(defaultValue); timeVal != nil {
			return timeVal, nil
		}
		return nil, fmt.Errorf("invalid timestamp default value: %s", defaultValue)

	default:
		return nil, fmt.Errorf("unsupported column type for default value: %s", columnType)
	}
}

// parseInt 解析整数 Parse integer
func parseInt(str string) *types.Value {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}

	// 处理特殊关键字 Handle special keywords
	switch strings.ToUpper(str) {
	case "NULL":
		return types.NewNullValue()
	case "0":
		return types.NewIntValue(0)
	}

	// 尝试解析为整数 Try to parse as integer
	if val, err := strconv.ParseInt(str, 10, 64); err == nil {
		return types.NewIntValue(val)
	}

	return nil
}

// parseFloat 解析浮点数 Parse float
func parseFloat(str string) *types.Value {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}

	// 处理特殊关键字 Handle special keywords
	switch strings.ToUpper(str) {
	case "NULL":
		return types.NewNullValue()
	case "0", "0.0":
		return types.NewFloatValue(0.0)
	}

	// 尝试解析为浮点数 Try to parse as float
	if val, err := strconv.ParseFloat(str, 64); err == nil {
		return types.NewFloatValue(val)
	}

	return nil
}

// parseBool 解析布尔值 Parse boolean
func parseBool(str string) *types.Value {
	str = strings.TrimSpace(strings.ToUpper(str))
	if str == "" {
		return nil
	}

	switch str {
	case "NULL":
		return types.NewNullValue()
	case "TRUE", "T", "1", "YES", "Y":
		return types.NewBoolValue(true)
	case "FALSE", "F", "0", "NO", "N":
		return types.NewBoolValue(false)
	}

	return nil
}

// parseTimestamp 解析时间戳 Parse timestamp
func parseTimestamp(str string) *types.Value {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}

	// 处理特殊关键字 Handle special keywords
	switch strings.ToUpper(str) {
	case "NULL":
		return types.NewNullValue()
	case "NOW", "CURRENT_TIMESTAMP":
		return types.NewTimestampValue(time.Now())
	}

	// 尝试各种时间格式 Try various time formats
	timeFormats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"15:04:05",
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, str); err == nil {
			return types.NewTimestampValue(t)
		}
	}

	return nil
}

// saveSchema 保存表结构 Save table schema
func (t *Table) saveSchema() error {
	key := t.encoder.EncodeMetadataKey(t.name, "schema")
	data, err := t.valueEncoder.EncodeSchema(t.schema)
	if err != nil {
		return fmt.Errorf("failed to encode schema: %w", err)
	}

	return t.database.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

// saveStatistics 保存表统计信息 Save table statistics
func (t *Table) saveStatistics() error {
	key := t.encoder.EncodeMetadataKey(t.name, "statistics")
	data, err := t.valueEncoder.EncodeTableStatistics(t.statistics)
	if err != nil {
		return fmt.Errorf("failed to encode statistics: %w", err)
	}

	return t.database.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

// saveIndexDefinition 保存索引定义 Save index definition
func (t *Table) saveIndexDefinition(index *Index) error {
	key := t.encoder.EncodeMetadataKey(t.name, fmt.Sprintf("index_%s", index.Name))
	data, err := t.valueEncoder.EncodeIndex(index)
	if err != nil {
		return fmt.Errorf("failed to encode index: %w", err)
	}

	return t.database.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

// deleteIndexDefinition 删除索引定义 Delete index definition
func (t *Table) deleteIndexDefinition(indexName string) error {
	key := t.encoder.EncodeMetadataKey(t.name, fmt.Sprintf("index_%s", indexName))
	return t.database.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// deleteMetadata 删除表元数据 Delete table metadata
func (t *Table) deleteMetadata() error {
	return t.database.db.Update(func(txn *badger.Txn) error {
		// 删除表结构 Delete table schema
		schemaKey := t.encoder.EncodeMetadataKey(t.name, "schema")
		if err := txn.Delete(schemaKey); err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		// 删除表统计信息 Delete table statistics
		statsKey := t.encoder.EncodeMetadataKey(t.name, "statistics")
		if err := txn.Delete(statsKey); err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		// 删除所有索引定义 Delete all index definitions
		for indexName := range t.indexes {
			indexKey := t.encoder.EncodeMetadataKey(t.name, fmt.Sprintf("index_%s", indexName))
			if err := txn.Delete(indexKey); err != nil && err != badger.ErrKeyNotFound {
				t.logger.Warn("Failed to delete index definition",
					"table", t.name,
					"index", indexName,
					"error", err)
			}
		}

		return nil
	})
}

// loadSchema 加载表结构 Load table schema
func (t *Table) loadSchema() error {
	key := t.encoder.EncodeMetadataKey(t.name, "schema")

	return t.database.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("failed to get schema: %w", err)
		}

		var data []byte
		if err := item.Value(func(val []byte) error {
			data = append([]byte(nil), val...)
			return nil
		}); err != nil {
			return fmt.Errorf("failed to read schema data: %w", err)
		}

		schema, err := t.valueEncoder.DecodeSchema(data)
		if err != nil {
			return fmt.Errorf("failed to decode schema: %w", err)
		}

		t.schema = schema
		return nil
	})
}

// loadStatistics 加载表统计信息 Load table statistics
func (t *Table) loadStatistics() error {
	key := t.encoder.EncodeMetadataKey(t.name, "statistics")

	return t.database.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			// 如果没有统计信息，创建默认的 If no statistics, create default
			t.statistics = &TableStatistics{
				RowCount:        0,
				ColumnCount:     len(t.schema.Columns),
				IndexCount:      0,
				ConstraintCount: 0,
				AvgRowSize:      0,
				DataSize:        0,
				IndexSize:       0,
				LastAnalyzed:    time.Now(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get statistics: %w", err)
		}

		var data []byte
		if err := item.Value(func(val []byte) error {
			data = append([]byte(nil), val...)
			return nil
		}); err != nil {
			return fmt.Errorf("failed to read statistics data: %w", err)
		}

		stats, err := t.valueEncoder.DecodeTableStatistics(data)
		if err != nil {
			return fmt.Errorf("failed to decode statistics: %w", err)
		}

		t.statistics = stats
		return nil
	})
}

// loadIndexes 加载索引定义 Load index definitions
func (t *Table) loadIndexes() error {
	prefix := t.encoder.EncodeMetadataKey(t.name, "index_")

	return t.database.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			var data []byte
			if err := item.Value(func(val []byte) error {
				data = append([]byte(nil), val...)
				return nil
			}); err != nil {
				continue
			}

			index, err := t.valueEncoder.DecodeIndex(data)
			if err != nil {
				t.logger.Warn("Failed to decode index definition",
					"table", t.name,
					"error", err)
				continue
			}

			t.indexes[index.Name] = index
		}

		return nil
	})
}

// GetSchema 获取表结构 Get table schema
func (t *Table) GetSchema() *types.Schema {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 返回结构副本 Return schema copy
	schema := *t.schema
	return &schema
}

// GetStatistics 获取表统计信息 Get table statistics
func (t *Table) GetStatistics() *TableStatistics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 返回统计信息副本 Return statistics copy
	stats := *t.statistics
	return &stats
}

// GetName 获取表名 Get table name
func (t *Table) GetName() string {
	return t.name
}

// GetRowCount 获取行数 Get row count
func (t *Table) GetRowCount() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.statistics.RowCount
}

// GetColumnCount 获取列数 Get column count
func (t *Table) GetColumnCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.schema.Columns)
}

// GetIndexCount 获取索引数 Get index count
func (t *Table) GetIndexCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.indexes)
}

// IsEmpty 检查表是否为空 Check if table is empty
func (t *Table) IsEmpty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.statistics.RowCount == 0
}

// GetDataSize 获取数据大小 Get data size
func (t *Table) GetDataSize() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.statistics.DataSize
}

// GetIndexSize 获取索引大小 Get index size
func (t *Table) GetIndexSize() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.statistics.IndexSize
}

// GetLastUpdated 获取最后更新时间 Get last updated time
func (t *Table) GetLastUpdated() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.statistics.UpdatedAt
}

// GetCreatedTime 获取创建时间 Get created time
func (t *Table) GetCreatedTime() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.statistics.CreatedAt
}

// Optimize 优化表 Optimize table
func (t *Table) Optimize(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.logger.Info("Starting table optimization",
		"table", t.name)

	start := time.Now()

	// 刷新统计信息 Refresh statistics
	if err := t.RefreshStatistics(ctx); err != nil {
		return fmt.Errorf("failed to refresh statistics: %w", err)
	}

	// 重建索引 Rebuild indexes
	for _, index := range t.indexes {
		if err := t.rebuildIndex(index); err != nil {
			t.logger.Warn("Failed to rebuild index",
				"table", t.name,
				"index", index.Name,
				"error", err)
		}
	}

	// 压缩存储 Compact storage
	if err := t.compactStorage(); err != nil {
		t.logger.Warn("Failed to compact storage",
			"table", t.name,
			"error", err)
	}

	elapsed := time.Since(start)
	t.logger.Info("Table optimization completed",
		"table", t.name,
		"duration", elapsed)

	return nil
}

// rebuildIndex 重建索引 Rebuild index
func (t *Table) rebuildIndex(index *Index) error {
	// 删除现有索引数据 Delete existing index data
	if err := t.database.db.Update(func(txn *badger.Txn) error {
		prefix := t.encoder.EncodeIndexPrefix(t.name, index.Name)
		return t.deleteKeysWithPrefix(txn, prefix)
	}); err != nil {
		return fmt.Errorf("failed to delete existing index data: %w", err)
	}

	// 重新构建索引 Rebuild index
	if err := t.buildIndexForExistingData(index); err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	// 重置索引统计信息 Reset index statistics
	index.Statistics.KeyCount = 0
	index.Statistics.UniqueKeys = 0
	index.Statistics.AvgKeySize = 0

	// 重新计算索引统计信息 Recalculate index statistics
	if err := t.recalculateIndexStatistics(index); err != nil {
		t.logger.Warn("Failed to recalculate index statistics",
			"table", t.name,
			"index", index.Name,
			"error", err)
	}

	return nil
}

// recalculateIndexStatistics 重新计算索引统计信息 Recalculate index statistics
func (t *Table) recalculateIndexStatistics(index *Index) error {
	return t.database.db.View(func(txn *badger.Txn) error {
		prefix := t.encoder.EncodeIndexPrefix(t.name, index.Name)
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		var keyCount int64
		var totalKeySize int64
		uniqueKeys := make(map[string]bool)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()

			keyCount++
			totalKeySize += int64(len(key))
			uniqueKeys[string(key)] = true
		}

		// 更新索引统计信息 Update index statistics
		index.Statistics.KeyCount = keyCount
		index.Statistics.UniqueKeys = int64(len(uniqueKeys))
		if keyCount > 0 {
			index.Statistics.AvgKeySize = float64(totalKeySize) / float64(keyCount)
		}

		return nil
	})
}

// compactStorage 压缩存储 Compact storage
func (t *Table) compactStorage() error {
	// 触发Badger的垃圾回收 Trigger Badger's garbage collection
	return t.database.db.RunValueLogGC(0.5)
}

// BackupTable 备份表 Backup table
func (t *Table) BackupTable(ctx context.Context, writer io.Writer) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	t.logger.Info("Starting table backup",
		"table", t.name)

	start := time.Now()
	var rowCount int64

	// 写入表结构 Write table schema
	schemaData, err := t.valueEncoder.EncodeSchema(t.schema)
	if err != nil {
		return fmt.Errorf("failed to encode schema: %w", err)
	}

	if _, err := writer.Write(schemaData); err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}

	// 写入分隔符 Write separator
	if _, err := writer.Write([]byte("\n---SCHEMA_END---\n")); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// 备份行数据 Backup row data
	err = t.database.db.View(func(txn *badger.Txn) error {
		prefix := t.encoder.EncodeRowPrefix(t.name)
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			// 写入键 Write key
			key := item.Key()
			if _, err := writer.Write(key); err != nil {
				return fmt.Errorf("failed to write key: %w", err)
			}

			// 写入键长度分隔符 Write key length separator
			if _, err := writer.Write([]byte(fmt.Sprintf("\n---KEY_LEN_%d---\n", len(key)))); err != nil {
				return fmt.Errorf("failed to write key separator: %w", err)
			}

			// 写入值 Write value
			err := item.Value(func(val []byte) error {
				if _, err := writer.Write(val); err != nil {
					return fmt.Errorf("failed to write value: %w", err)
				}
				return nil
			})
			if err != nil {
				return err
			}

			// 写入行分隔符 Write row separator
			if _, err := writer.Write([]byte("\n---ROW_END---\n")); err != nil {
				return fmt.Errorf("failed to write row separator: %w", err)
			}

			rowCount++
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to backup row data: %w", err)
	}

	elapsed := time.Since(start)
	t.logger.Info("Table backup completed",
		"table", t.name,
		"rows", rowCount,
		"duration", elapsed)

	return nil
}

// RestoreTable 恢复表 Restore table
func (t *Table) RestoreTable(ctx context.Context, reader io.Reader) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.logger.Info("Starting table restore",
		"table", t.name)

	start := time.Now()

	// 读取所有数据 Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read backup data: %w", err)
	}

	content := string(data)

	// 分离结构和数据 Separate schema and data
	parts := strings.Split(content, "\n---SCHEMA_END---\n")
	if len(parts) != 2 {
		return fmt.Errorf("invalid backup format: schema separator not found")
	}

	schemaData := []byte(parts[0])
	rowsData := parts[1]

	// 恢复表结构 Restore table schema
	schema, err := t.valueEncoder.DecodeSchema(schemaData)
	if err != nil {
		return fmt.Errorf("failed to decode schema: %w", err)
	}

	t.schema = schema

	// 清空现有数据 Clear existing data
	if err := t.truncateData(ctx); err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// 恢复行数据 Restore row data
	var rowCount int64
	err = t.database.db.Update(func(txn *badger.Txn) error {
		rows := strings.Split(rowsData, "\n---ROW_END---\n")

		for _, rowData := range rows {
			if strings.TrimSpace(rowData) == "" {
				continue
			}

			// 解析键和值 Parse key and value
			if err := t.restoreRow(txn, rowData); err != nil {
				t.logger.Warn("Failed to restore row",
					"table", t.name,
					"error", err)
				continue
			}

			rowCount++
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to restore row data: %w", err)
	}

	// 更新统计信息 Update statistics
	t.statistics.RowCount = rowCount
	t.statistics.UpdatedAt = time.Now()

	elapsed := time.Since(start)
	t.logger.Info("Table restore completed",
		"table", t.name,
		"rows", rowCount,
		"duration", elapsed)

	return nil
}

// restoreRow 恢复单行数据 Restore single row data
func (t *Table) restoreRow(txn *badger.Txn, rowData string) error {
	// 查找键长度分隔符 Find key length separator
	keyLenStart := strings.Index(rowData, "\n---KEY_LEN_")
	if keyLenStart == -1 {
		return fmt.Errorf("key length separator not found")
	}

	key := []byte(rowData[:keyLenStart])

	// 解析键长度 Parse key length
	keyLenEnd := strings.Index(rowData[keyLenStart:], "---\n")
	if keyLenEnd == -1 {
		return fmt.Errorf("key length end separator not found")
	}

	keyLenStr := rowData[keyLenStart+12 : keyLenStart+keyLenEnd]
	keyLen, err := strconv.Atoi(keyLenStr)
	if err != nil {
		return fmt.Errorf("failed to parse key length: %w", err)
	}

	// 验证键长度 Verify key length
	if len(key) != keyLen {
		return fmt.Errorf("key length mismatch: expected %d, got %d", keyLen, len(key))
	}

	// 提取值 Extract value
	valueStart := keyLenStart + keyLenEnd + 4
	value := []byte(rowData[valueStart:])

	// 设置键值对 Set key-value pair
	return txn.Set(key, value)
}

// Close 关闭表 Close table
func (t *Table) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 保存最终统计信息 Save final statistics
	if err := t.saveStatistics(); err != nil {
		t.logger.Warn("Failed to save final statistics",
			"table", t.name,
			"error", err)
	}

	t.logger.Info("Table closed",
		"table", t.name)

	return nil
}

// String 字符串表示 String representation
func (t *Table) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return fmt.Sprintf("Table{name=%s, columns=%d, rows=%d, indexes=%d}",
		t.name,
		len(t.schema.Columns),
		t.statistics.RowCount,
		len(t.indexes))
}

// GetAutoIncrementValue 获取自增值 Get auto increment value
func (tm *TableManager) GetAutoIncrementValue(tableName string) int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.autoIncrement[tableName]
}

// SetAutoIncrementValue 设置自增值 Set auto increment value
func (tm *TableManager) SetAutoIncrementValue(tableName string, value int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.autoIncrement[tableName] = value
}

// Close 关闭表管理器 Close table manager
func (tm *TableManager) Close() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 关闭所有表 Close all tables
	for name, table := range tm.tables {
		if err := table.Close(); err != nil {
			tm.logger.Warn("Failed to close table",
				"table", name,
				"error", err)
		}
	}

	tm.logger.Info("Table manager closed")

	return nil
}

// GetTableCount 获取表数量 Get table count
func (tm *TableManager) GetTableCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return len(tm.tables)
}

// GetTotalRowCount 获取总行数 Get total row count
func (tm *TableManager) GetTotalRowCount() int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var total int64
	for _, table := range tm.tables {
		total += table.GetRowCount()
	}

	return total
}

// GetTotalDataSize 获取总数据大小 Get total data size
func (tm *TableManager) GetTotalDataSize() int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var total int64
	for _, table := range tm.tables {
		total += table.GetDataSize()
	}

	return total
}

// GetTotalIndexSize 获取总索引大小 Get total index size
func (tm *TableManager) GetTotalIndexSize() int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var total int64
	for _, table := range tm.tables {
		total += table.GetIndexSize()
	}

	return total
}

// ValidateTableName 验证表名（导出方法） Validate table name (exported method)
func (tm *TableManager) ValidateTableName(tableName string) error {
	return tm.validateTableName(tableName)
}

// ValidateSchema 验证表结构（导出方法） Validate schema (exported method)
func (tm *TableManager) ValidateSchema(schema *types.Schema) error {
	return tm.validateSchema(schema)
}
