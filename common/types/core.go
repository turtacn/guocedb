// Package types defines core data types for GuoceDB
// 类型包，定义GuoceDB的核心数据类型
package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// DataType 数据类型枚举 Data type enumeration
type DataType int

const (
	// 基础数据类型 Basic data types
	DataTypeUnknown   DataType = iota // 未知类型 Unknown type
	DataTypeNull                      // 空值类型 Null type
	DataTypeBool                      // 布尔类型 Boolean type
	DataTypeInt8                      // 8位整数 8-bit integer
	DataTypeInt16                     // 16位整数 16-bit integer
	DataTypeInt32                     // 32位整数 32-bit integer
	DataTypeInt64                     // 64位整数 64-bit integer
	DataTypeUint8                     // 8位无符号整数 8-bit unsigned integer
	DataTypeUint16                    // 16位无符号整数 16-bit unsigned integer
	DataTypeUint32                    // 32位无符号整数 32-bit unsigned integer
	DataTypeUint64                    // 64位无符号整数 64-bit unsigned integer
	DataTypeFloat32                   // 32位浮点数 32-bit floating point
	DataTypeFloat64                   // 64位浮点数 64-bit floating point
	DataTypeDecimal                   // 高精度小数 High precision decimal
	DataTypeString                    // 字符串类型 String type
	DataTypeBytes                     // 字节数组类型 Byte array type
	DataTypeDate                      // 日期类型 Date type
	DataTypeTime                      // 时间类型 Time type
	DataTypeDateTime                  // 日期时间类型 DateTime type
	DataTypeTimestamp                 // 时间戳类型 Timestamp type
	DataTypeInterval                  // 时间间隔类型 Interval type

	// 复合数据类型 Composite data types
	DataTypeArray  // 数组类型 Array type
	DataTypeMap    // 映射类型 Map type
	DataTypeStruct // 结构类型 Struct type
	DataTypeJSON   // JSON类型 JSON type
	DataTypeXML    // XML类型 XML type

	// 特殊数据类型 Special data types
	DataTypeUUID      // UUID类型 UUID type
	DataTypeGeoPoint  // 地理坐标类型 Geographic point type
	DataTypeGeoShape  // 地理形状类型 Geographic shape type
	DataTypeIPAddress // IP地址类型 IP address type
	DataTypeBinary    // 二进制类型 Binary type
	DataTypeText      // 文本类型 Text type (长文本)
	DataTypeEnum      // 枚举类型 Enumeration type
	DataTypeSet       // 集合类型 Set type
	DataTypeRef       // 引用类型 Reference type
	DataTypeCustom    // 自定义类型 Custom type
	DataTypeVector    //
)

// String 返回数据类型的字符串表示 Returns string representation of data type
func (dt DataType) String() string {
	switch dt {
	case DataTypeUnknown:
		return "UNKNOWN"
	case DataTypeNull:
		return "NULL"
	case DataTypeBool:
		return "BOOL"
	case DataTypeInt8:
		return "INT8"
	case DataTypeInt16:
		return "INT16"
	case DataTypeInt32:
		return "INT32"
	case DataTypeInt64:
		return "INT64"
	case DataTypeUint8:
		return "UINT8"
	case DataTypeUint16:
		return "UINT16"
	case DataTypeUint32:
		return "UINT32"
	case DataTypeUint64:
		return "UINT64"
	case DataTypeFloat32:
		return "FLOAT32"
	case DataTypeFloat64:
		return "FLOAT64"
	case DataTypeDecimal:
		return "DECIMAL"
	case DataTypeString:
		return "STRING"
	case DataTypeBytes:
		return "BYTES"
	case DataTypeDate:
		return "DATE"
	case DataTypeTime:
		return "TIME"
	case DataTypeDateTime:
		return "DATETIME"
	case DataTypeTimestamp:
		return "TIMESTAMP"
	case DataTypeInterval:
		return "INTERVAL"
	case DataTypeArray:
		return "ARRAY"
	case DataTypeMap:
		return "MAP"
	case DataTypeStruct:
		return "STRUCT"
	case DataTypeJSON:
		return "JSON"
	case DataTypeXML:
		return "XML"
	case DataTypeUUID:
		return "UUID"
	case DataTypeGeoPoint:
		return "GEOPOINT"
	case DataTypeGeoShape:
		return "GEOSHAPE"
	case DataTypeIPAddress:
		return "IPADDRESS"
	case DataTypeBinary:
		return "BINARY"
	case DataTypeText:
		return "TEXT"
	case DataTypeEnum:
		return "ENUM"
	case DataTypeSet:
		return "SET"
	case DataTypeRef:
		return "REF"
	case DataTypeCustom:
		return "CUSTOM"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(dt))
	}
}

// IsNumeric 检查是否为数值类型 Check if it's a numeric type
func (dt DataType) IsNumeric() bool {
	switch dt {
	case DataTypeInt8, DataTypeInt16, DataTypeInt32, DataTypeInt64,
		DataTypeUint8, DataTypeUint16, DataTypeUint32, DataTypeUint64,
		DataTypeFloat32, DataTypeFloat64, DataTypeDecimal:
		return true
	default:
		return false
	}
}

// IsInteger 检查是否为整数类型 Check if it's an integer type
func (dt DataType) IsInteger() bool {
	switch dt {
	case DataTypeInt8, DataTypeInt16, DataTypeInt32, DataTypeInt64,
		DataTypeUint8, DataTypeUint16, DataTypeUint32, DataTypeUint64:
		return true
	default:
		return false
	}
}

// IsFloat 检查是否为浮点类型 Check if it's a floating point type
func (dt DataType) IsFloat() bool {
	switch dt {
	case DataTypeFloat32, DataTypeFloat64, DataTypeDecimal:
		return true
	default:
		return false
	}
}

// IsString 检查是否为字符串类型 Check if it's a string type
func (dt DataType) IsString() bool {
	switch dt {
	case DataTypeString, DataTypeText:
		return true
	default:
		return false
	}
}

// IsTime 检查是否为时间类型 Check if it's a time type
func (dt DataType) IsTime() bool {
	switch dt {
	case DataTypeDate, DataTypeTime, DataTypeDateTime, DataTypeTimestamp, DataTypeInterval:
		return true
	default:
		return false
	}
}

// IsComposite 检查是否为复合类型 Check if it's a composite type
func (dt DataType) IsComposite() bool {
	switch dt {
	case DataTypeArray, DataTypeMap, DataTypeStruct, DataTypeJSON, DataTypeXML:
		return true
	default:
		return false
	}
}

// Size 返回数据类型的字节大小，-1表示可变长度 Returns byte size of data type, -1 for variable length
func (dt DataType) Size() int {
	switch dt {
	case DataTypeBool, DataTypeInt8, DataTypeUint8:
		return 1
	case DataTypeInt16, DataTypeUint16:
		return 2
	case DataTypeInt32, DataTypeUint32, DataTypeFloat32:
		return 4
	case DataTypeInt64, DataTypeUint64, DataTypeFloat64, DataTypeTimestamp:
		return 8
	case DataTypeDate:
		return 4 // YYYY-MM-DD as int32
	case DataTypeTime:
		return 8 // nanoseconds since midnight
	case DataTypeDateTime:
		return 8 // Unix timestamp with nanoseconds
	case DataTypeUUID:
		return 16
	case DataTypeIPAddress:
		return 16 // IPv6 address
	default:
		return -1 // Variable length
	}
}

// ParseDataType 从字符串解析数据类型 Parse data type from string
func ParseDataType(s string) DataType {
	switch strings.ToUpper(s) {
	case "NULL":
		return DataTypeNull
	case "BOOL", "BOOLEAN":
		return DataTypeBool
	case "INT8", "TINYINT":
		return DataTypeInt8
	case "INT16", "SMALLINT":
		return DataTypeInt16
	case "INT32", "INT", "INTEGER":
		return DataTypeInt32
	case "INT64", "BIGINT":
		return DataTypeInt64
	case "UINT8":
		return DataTypeUint8
	case "UINT16":
		return DataTypeUint16
	case "UINT32":
		return DataTypeUint32
	case "UINT64":
		return DataTypeUint64
	case "FLOAT32", "FLOAT":
		return DataTypeFloat32
	case "FLOAT64", "DOUBLE":
		return DataTypeFloat64
	case "DECIMAL", "NUMERIC":
		return DataTypeDecimal
	case "STRING", "VARCHAR", "CHAR":
		return DataTypeString
	case "BYTES", "BINARY", "VARBINARY":
		return DataTypeBytes
	case "DATE":
		return DataTypeDate
	case "TIME":
		return DataTypeTime
	case "DATETIME":
		return DataTypeDateTime
	case "TIMESTAMP":
		return DataTypeTimestamp
	case "INTERVAL":
		return DataTypeInterval
	case "ARRAY":
		return DataTypeArray
	case "MAP":
		return DataTypeMap
	case "STRUCT":
		return DataTypeStruct
	case "JSON":
		return DataTypeJSON
	case "XML":
		return DataTypeXML
	case "UUID":
		return DataTypeUUID
	case "GEOPOINT":
		return DataTypeGeoPoint
	case "GEOSHAPE":
		return DataTypeGeoShape
	case "IPADDRESS":
		return DataTypeIPAddress
	case "TEXT":
		return DataTypeText
	case "ENUM":
		return DataTypeEnum
	case "SET":
		return DataTypeSet
	case "REF", "REFERENCE":
		return DataTypeRef
	default:
		return DataTypeUnknown
	}
}

// Value 数据值结构 Data value structure
type Value struct {
	Type     DataType               `json:"type"`               // 数据类型 Data type
	Data     interface{}            `json:"data"`               // 数据内容 Data content
	Nullable bool                   `json:"nullable"`           // 是否可为空 Whether nullable
	Size     int                    `json:"size,omitempty"`     // 数据大小 Data size
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 元数据 Metadata
}

// NewValue 创建新的数据值 Create new data value
func NewValue(dataType DataType, data interface{}) *Value {
	return &Value{
		Type:     dataType,
		Data:     data,
		Nullable: false,
		Size:     calculateSize(dataType, data),
		Metadata: make(map[string]interface{}),
	}
}

// NewNullableValue 创建可空的数据值 Create nullable data value
func NewNullableValue(dataType DataType, data interface{}) *Value {
	return &Value{
		Type:     dataType,
		Data:     data,
		Nullable: true,
		Size:     calculateSize(dataType, data),
		Metadata: make(map[string]interface{}),
	}
}

// IsNull 检查值是否为空 Check if value is null
func (v *Value) IsNull() bool {
	return v.Data == nil
}

// String 返回值的字符串表示 Returns string representation of value
func (v *Value) String() string {
	if v.IsNull() {
		return "NULL"
	}

	switch v.Type {
	case DataTypeString, DataTypeText:
		return fmt.Sprintf("'%s'", v.Data)
	case DataTypeDate, DataTypeTime, DataTypeDateTime, DataTypeTimestamp:
		return fmt.Sprintf("'%s'", v.Data)
	default:
		return fmt.Sprintf("%v", v.Data)
	}
}

// Equals 比较两个值是否相等 Compare if two values are equal
func (v *Value) Equals(other *Value) bool {
	if v == nil && other == nil {
		return true
	}
	if v == nil || other == nil {
		return false
	}

	if v.Type != other.Type {
		return false
	}

	if v.IsNull() && other.IsNull() {
		return true
	}
	if v.IsNull() || other.IsNull() {
		return false
	}

	return v.Data == other.Data
}

// Clone 克隆值 Clone value
func (v *Value) Clone() *Value {
	if v == nil {
		return nil
	}

	clone := &Value{
		Type:     v.Type,
		Data:     v.Data, // 浅拷贝，对于复杂类型可能需要深拷贝
		Nullable: v.Nullable,
		Size:     v.Size,
		Metadata: make(map[string]interface{}),
	}

	// 拷贝元数据 Copy metadata
	for k, val := range v.Metadata {
		clone.Metadata[k] = val
	}

	return clone
}

// ConvertTo 转换为指定类型 Convert to specified type
func (v *Value) ConvertTo(targetType DataType) (*Value, error) {
	if v.Type == targetType {
		return v.Clone(), nil
	}

	if v.IsNull() {
		return NewNullableValue(targetType, nil), nil
	}

	converted, err := convertValue(v.Data, v.Type, targetType)
	if err != nil {
		return nil, err
	}

	result := NewValue(targetType, converted)
	result.Nullable = v.Nullable
	return result, nil
}

// calculateSize 计算数据大小 Calculate data size
func calculateSize(dataType DataType, data interface{}) int {
	if data == nil {
		return 0
	}

	fixedSize := dataType.Size()
	if fixedSize > 0 {
		return fixedSize
	}

	// 可变长度类型的大小计算 Size calculation for variable length types
	switch dataType {
	case DataTypeString, DataTypeText:
		if s, ok := data.(string); ok {
			return utf8.RuneCountInString(s)
		}
	case DataTypeBytes, DataTypeBinary:
		if b, ok := data.([]byte); ok {
			return len(b)
		}
	case DataTypeArray:
		// 简化处理，实际应该递归计算 Simplified handling, should recursively calculate
		if arr, ok := data.([]interface{}); ok {
			return len(arr) * 8 // 假设每个元素8字节 Assume 8 bytes per element
		}
	case DataTypeMap:
		if m, ok := data.(map[string]interface{}); ok {
			return len(m) * 16 // 假设每个键值对16字节 Assume 16 bytes per key-value pair
		}
	case DataTypeJSON:
		if jsonData, err := json.Marshal(data); err == nil {
			return len(jsonData)
		}
	}

	return 0
}

// convertValue 转换值类型 Convert value type
func convertValue(data interface{}, fromType, toType DataType) (interface{}, error) {
	if data == nil {
		return nil, nil
	}

	// 字符串到其他类型的转换 String to other types conversion
	if fromType == DataTypeString {
		str := data.(string)
		switch toType {
		case DataTypeBool:
			return strconv.ParseBool(str)
		case DataTypeInt32:
			val, err := strconv.ParseInt(str, 10, 32)
			return int32(val), err
		case DataTypeInt64:
			return strconv.ParseInt(str, 10, 64)
		case DataTypeFloat32:
			val, err := strconv.ParseFloat(str, 32)
			return float32(val), err
		case DataTypeFloat64:
			return strconv.ParseFloat(str, 64)
		case DataTypeDateTime:
			return time.Parse(time.RFC3339, str)
		}
	}

	// 数值类型之间的转换 Conversion between numeric types
	if fromType.IsNumeric() && toType.IsNumeric() {
		return convertNumeric(data, toType)
	}

	// 其他类型到字符串的转换 Other types to string conversion
	if toType == DataTypeString {
		return fmt.Sprintf("%v", data), nil
	}

	return nil, fmt.Errorf("cannot convert from %s to %s", fromType.String(), toType.String())
}

// convertNumeric 转换数值类型 Convert numeric types
func convertNumeric(data interface{}, toType DataType) (interface{}, error) {
	switch toType {
	case DataTypeInt8:
		if val, ok := convertToInt64(data); ok {
			return int8(val), nil
		}
	case DataTypeInt16:
		if val, ok := convertToInt64(data); ok {
			return int16(val), nil
		}
	case DataTypeInt32:
		if val, ok := convertToInt64(data); ok {
			return int32(val), nil
		}
	case DataTypeInt64:
		if val, ok := convertToInt64(data); ok {
			return val, nil
		}
	case DataTypeFloat32:
		if val, ok := convertToFloat64(data); ok {
			return float32(val), nil
		}
	case DataTypeFloat64:
		if val, ok := convertToFloat64(data); ok {
			return val, nil
		}
	}

	return nil, fmt.Errorf("cannot convert %v to %s", data, toType.String())
}

// convertToInt64 转换为int64 Convert to int64
func convertToInt64(data interface{}) (int64, bool) {
	switch v := data.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

// convertToFloat64 转换为float64 Convert to float64
func convertToFloat64(data interface{}) (float64, bool) {
	switch v := data.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// Column 列定义 Column definition
type Column struct {
	Name          string                 `json:"name"`                    // 列名 Column name
	Type          DataType               `json:"type"`                    // 数据类型 Data type
	Nullable      bool                   `json:"nullable"`                // 是否可为空 Whether nullable
	DefaultValue  *Value                 `json:"default_value,omitempty"` // 默认值 Default value
	PrimaryKey    bool                   `json:"primary_key"`             // 是否主键 Whether primary key
	AutoIncrement bool                   `json:"auto_increment"`          // 是否自增 Whether auto increment
	Unique        bool                   `json:"unique"`                  // 是否唯一 Whether unique
	Index         bool                   `json:"index"`                   // 是否索引 Whether indexed
	Comment       string                 `json:"comment,omitempty"`       // 注释 Comment
	Constraints   []Constraint           `json:"constraints,omitempty"`   // 约束条件 Constraints
	Metadata      map[string]interface{} `json:"metadata,omitempty"`      // 元数据 Metadata

	// 扩展属性 Extended properties
	Length     int      `json:"length,omitempty"`      // 长度限制 Length limit
	Precision  int      `json:"precision,omitempty"`   // 精度 Precision
	Scale      int      `json:"scale,omitempty"`       // 小数位数 Scale
	EnumValues []string `json:"enum_values,omitempty"` // 枚举值 Enum values
	Format     string   `json:"format,omitempty"`      // 格式 Format
	Encoding   string   `json:"encoding,omitempty"`    // 编码 Encoding
}

// NewColumn 创建新列 Create new column
func NewColumn(name string, dataType DataType) *Column {
	return &Column{
		Name:        name,
		Type:        dataType,
		Nullable:    true,
		Constraints: make([]Constraint, 0),
		Metadata:    make(map[string]interface{}),
	}
}

// Clone 克隆列定义 Clone column definition
func (c *Column) Clone() *Column {
	if c == nil {
		return nil
	}

	clone := &Column{
		Name:          c.Name,
		Type:          c.Type,
		Nullable:      c.Nullable,
		DefaultValue:  c.DefaultValue.Clone(),
		PrimaryKey:    c.PrimaryKey,
		AutoIncrement: c.AutoIncrement,
		Unique:        c.Unique,
		Index:         c.Index,
		Comment:       c.Comment,
		Length:        c.Length,
		Precision:     c.Precision,
		Scale:         c.Scale,
		EnumValues:    append([]string(nil), c.EnumValues...),
		Format:        c.Format,
		Encoding:      c.Encoding,
		Constraints:   make([]Constraint, len(c.Constraints)),
		Metadata:      make(map[string]interface{}),
	}

	// 拷贝约束 Copy constraints
	copy(clone.Constraints, c.Constraints)

	// 拷贝元数据 Copy metadata
	for k, v := range c.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

// ValidateValue 验证值是否符合列定义 Validate if value conforms to column definition
func (c *Column) ValidateValue(value *Value) error {
	if value == nil {
		if !c.Nullable {
			return fmt.Errorf("column %s cannot be null", c.Name)
		}
		return nil
	}

	if value.IsNull() {
		if !c.Nullable {
			return fmt.Errorf("column %s cannot be null", c.Name)
		}
		return nil
	}

	// 类型检查 Type check
	if value.Type != c.Type {
		return fmt.Errorf("column %s expects type %s, got %s", c.Name, c.Type.String(), value.Type.String())
	}

	// 长度检查 Length check
	if c.Length > 0 && c.Type.IsString() {
		if str, ok := value.Data.(string); ok {
			if utf8.RuneCountInString(str) > c.Length {
				return fmt.Errorf("column %s value length exceeds limit %d", c.Name, c.Length)
			}
		}
	}

	// 枚举值检查 Enum values check
	if len(c.EnumValues) > 0 && c.Type == DataTypeEnum {
		if str, ok := value.Data.(string); ok {
			found := false
			for _, enumVal := range c.EnumValues {
				if str == enumVal {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("column %s value %s is not in allowed enum values", c.Name, str)
			}
		}
	}

	// 约束检查 Constraints check
	for _, constraint := range c.Constraints {
		if err := constraint.Validate(value); err != nil {
			return fmt.Errorf("column %s constraint violation: %w", c.Name, err)
		}
	}

	return nil
}

// Schema 表结构定义 Table schema definition
type Schema struct {
	Name        string                 `json:"name"`                  // 表名 Table name
	Columns     []*Column              `json:"columns"`               // 列定义 Column definitions
	PrimaryKeys []string               `json:"primary_keys"`          // 主键列名 Primary key column names
	Indexes     []*Index               `json:"indexes,omitempty"`     // 索引定义 Index definitions
	Constraints []Constraint           `json:"constraints,omitempty"` // 表级约束 Table-level constraints
	Comment     string                 `json:"comment,omitempty"`     // 注释 Comment
	Options     map[string]interface{} `json:"options,omitempty"`     // 选项 Options
	Version     int                    `json:"version"`               // 版本号 Version number
	CreatedAt   time.Time              `json:"created_at"`            // 创建时间 Creation time
	UpdatedAt   time.Time              `json:"updated_at"`            // 更新时间 Update time
}

// NewSchema 创建新的表结构 Create new table schema
func NewSchema(name string) *Schema {
	now := time.Now()
	return &Schema{
		Name:        name,
		Columns:     make([]*Column, 0),
		PrimaryKeys: make([]string, 0),
		Indexes:     make([]*Index, 0),
		Constraints: make([]Constraint, 0),
		Options:     make(map[string]interface{}),
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AddColumn 添加列 Add column
func (s *Schema) AddColumn(column *Column) {
	s.Columns = append(s.Columns, column)

	if column.PrimaryKey {
		s.PrimaryKeys = append(s.PrimaryKeys, column.Name)
	}

	s.UpdatedAt = time.Now()
}

// RemoveColumn 移除列 Remove column
func (s *Schema) RemoveColumn(name string) bool {
	for i, col := range s.Columns {
		if col.Name == name {
			s.Columns = append(s.Columns[:i], s.Columns[i+1:]...)

			// 从主键中移除 Remove from primary keys
			for j, pk := range s.PrimaryKeys {
				if pk == name {
					s.PrimaryKeys = append(s.PrimaryKeys[:j], s.PrimaryKeys[j+1:]...)
					break
				}
			}

			s.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetColumn 获取列定义 Get column definition
func (s *Schema) GetColumn(name string) *Column {
	for _, col := range s.Columns {
		if col.Name == name {
			return col
		}
	}
	return nil
}

// GetColumnNames 获取所有列名 Get all column names
func (s *Schema) GetColumnNames() []string {
	names := make([]string, len(s.Columns))
	for i, col := range s.Columns {
		names[i] = col.Name
	}
	return names
}

// Clone 克隆表结构 Clone table schema
func (s *Schema) Clone() *Schema {
	if s == nil {
		return nil
	}

	clone := &Schema{
		Name:        s.Name,
		PrimaryKeys: append([]string(nil), s.PrimaryKeys...),
		Comment:     s.Comment,
		Version:     s.Version,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Columns:     make([]*Column, len(s.Columns)),
		Indexes:     make([]*Index, len(s.Indexes)),
		Constraints: make([]Constraint, len(s.Constraints)),
		Options:     make(map[string]interface{}),
	}

	// 克隆列 Clone columns
	for i, col := range s.Columns {
		clone.Columns[i] = col.Clone()
	}

	// 克隆索引 Clone indexes
	for i, idx := range s.Indexes {
		clone.Indexes[i] = idx.Clone()
	}

	// 克隆约束 Clone constraints
	copy(clone.Constraints, s.Constraints)

	// 克隆选项 Clone options
	for k, v := range s.Options {
		clone.Options[k] = v
	}

	return clone
}

// ValidateRow 验证行数据是否符合表结构 Validate if row data conforms to table schema
func (s *Schema) ValidateRow(row *Row) error {
	if row == nil {
		return fmt.Errorf("row cannot be nil")
	}

	// 检查列数量 Check column count
	if len(row.Values) != len(s.Columns) {
		return fmt.Errorf("row has %d values, schema expects %d", len(row.Values), len(s.Columns))
	}

	// 验证每个值 Validate each value
	for i, col := range s.Columns {
		if i >= len(row.Values) {
			return fmt.Errorf("missing value for column %s", col.Name)
		}

		if err := col.ValidateValue(row.Values[i]); err != nil {
			return fmt.Errorf("column %s validation failed: %w", col.Name, err)
		}
	}

	// 验证表级约束 Validate table-level constraints
	for _, constraint := range s.Constraints {
		if err := constraint.ValidateRow(row); err != nil {
			return fmt.Errorf("table constraint violation: %w", err)
		}
	}

	return nil
}

// Row 数据行结构 Data row structure
type Row struct {
	ID        string                 `json:"id,omitempty"`       // 行ID Row ID
	Values    []*Value               `json:"values"`             // 值数组 Value array
	Version   int64                  `json:"version"`            // 版本号 Version number
	CreatedAt time.Time              `json:"created_at"`         // 创建时间 Creation time
	UpdatedAt time.Time              `json:"updated_at"`         // 更新时间 Update time
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // 元数据 Metadata
	Deleted   bool                   `json:"deleted"`            // 是否已删除 Whether deleted
}

// NewRow 创建新行 Create new row
func NewRow(values []*Value) *Row {
	now := time.Now()
	return &Row{
		Values:    values,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]interface{}),
		Deleted:   false,
	}
}

// Clone 克隆行 Clone row
func (r *Row) Clone() *Row {
	if r == nil {
		return nil
	}

	clone := &Row{
		ID:        r.ID,
		Version:   r.Version,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		Deleted:   r.Deleted,
		Values:    make([]*Value, len(r.Values)),
		Metadata:  make(map[string]interface{}),
	}

	// 克隆值 Clone values
	for i, val := range r.Values {
		clone.Values[i] = val.Clone()
	}

	// 克隆元数据 Clone metadata
	for k, v := range r.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

// GetValue 获取指定位置的值 Get value at specified position
func (r *Row) GetValue(index int) *Value {
	if index < 0 || index >= len(r.Values) {
		return nil
	}
	return r.Values[index]
}

// SetValue 设置指定位置的值 Set value at specified position
func (r *Row) SetValue(index int, value *Value) bool {
	if index < 0 || index >= len(r.Values) {
		return false
	}
	r.Values[index] = value
	r.UpdatedAt = time.Now()
	r.Version++
	return true
}

// GetValueByName 根据列名获取值 Get value by column name
func (r *Row) GetValueByName(schema *Schema, columnName string) *Value {
	for i, col := range schema.Columns {
		if col.Name == columnName {
			return r.GetValue(i)
		}
	}
	return nil
}

// SetValueByName 根据列名设置值 Set value by column name
func (r *Row) SetValueByName(schema *Schema, columnName string, value *Value) bool {
	for i, col := range schema.Columns {
		if col.Name == columnName {
			return r.SetValue(i, value)
		}
	}
	return false
}

// ToMap 转换为映射 Convert to map
func (r *Row) ToMap(schema *Schema) map[string]interface{} {
	result := make(map[string]interface{})

	for i, col := range schema.Columns {
		if i < len(r.Values) && r.Values[i] != nil {
			if r.Values[i].IsNull() {
				result[col.Name] = nil
			} else {
				result[col.Name] = r.Values[i].Data
			}
		}
	}

	return result
}

// Index 索引定义 Index definition
type Index struct {
	Name       string                 `json:"name"`                 // 索引名 Index name
	Table      string                 `json:"table"`                // 表名 Table name
	Columns    []string               `json:"columns"`              // 列名列表 Column name list
	Type       IndexType              `json:"type"`                 // 索引类型 Index type
	Unique     bool                   `json:"unique"`               // 是否唯一索引 Whether unique index
	Clustered  bool                   `json:"clustered"`            // 是否聚集索引 Whether clustered index
	Method     IndexMethod            `json:"method"`               // 索引方法 Index method
	Expression string                 `json:"expression,omitempty"` // 表达式索引 Expression index
	Condition  string                 `json:"condition,omitempty"`  // 部分索引条件 Partial index condition
	Options    map[string]interface{} `json:"options,omitempty"`    // 索引选项 Index options
	Comment    string                 `json:"comment,omitempty"`    // 注释 Comment
	CreatedAt  time.Time              `json:"created_at"`           // 创建时间 Creation time
}

// IndexType 索引类型 Index type
type IndexType int

const (
	IndexTypePrimary    IndexType = iota // 主键索引 Primary key index
	IndexTypeUnique                      // 唯一索引 Unique index
	IndexTypeRegular                     // 普通索引 Regular index
	IndexTypeComposite                   // 复合索引 Composite index
	IndexTypePartial                     // 部分索引 Partial index
	IndexTypeExpression                  // 表达式索引 Expression index
	IndexTypeSpatial                     // 空间索引 Spatial index
	IndexTypeFullText                    // 全文索引 Full-text index
)

// String 返回索引类型的字符串表示 Returns string representation of index type
func (it IndexType) String() string {
	switch it {
	case IndexTypePrimary:
		return "PRIMARY"
	case IndexTypeUnique:
		return "UNIQUE"
	case IndexTypeRegular:
		return "REGULAR"
	case IndexTypeComposite:
		return "COMPOSITE"
	case IndexTypePartial:
		return "PARTIAL"
	case IndexTypeExpression:
		return "EXPRESSION"
	case IndexTypeSpatial:
		return "SPATIAL"
	case IndexTypeFullText:
		return "FULLTEXT"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(it))
	}
}

// IndexMethod 索引方法 Index method
type IndexMethod int

const (
	IndexMethodBTree   IndexMethod = iota // B树索引 B-tree index
	IndexMethodHash                       // 哈希索引 Hash index
	IndexMethodGiST                       // GiST索引 GiST index
	IndexMethodGIN                        // GIN索引 GIN index
	IndexMethodBRIN                       // BRIN索引 BRIN index
	IndexMethodSP_GiST                    // SP-GiST索引 SP-GiST index
)

// String 返回索引方法的字符串表示 Returns string representation of index method
func (im IndexMethod) String() string {
	switch im {
	case IndexMethodBTree:
		return "BTREE"
	case IndexMethodHash:
		return "HASH"
	case IndexMethodGiST:
		return "GIST"
	case IndexMethodGIN:
		return "GIN"
	case IndexMethodBRIN:
		return "BRIN"
	case IndexMethodSP_GiST:
		return "SP_GIST"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(im))
	}
}

// NewIndex 创建新索引 Create new index
func NewIndex(name, table string, columns []string) *Index {
	return &Index{
		Name:      name,
		Table:     table,
		Columns:   columns,
		Type:      IndexTypeRegular,
		Method:    IndexMethodBTree,
		Options:   make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// Clone 克隆索引定义 Clone index definition
func (idx *Index) Clone() *Index {
	if idx == nil {
		return nil
	}

	clone := &Index{
		Name:       idx.Name,
		Table:      idx.Table,
		Columns:    append([]string(nil), idx.Columns...),
		Type:       idx.Type,
		Unique:     idx.Unique,
		Clustered:  idx.Clustered,
		Method:     idx.Method,
		Expression: idx.Expression,
		Condition:  idx.Condition,
		Comment:    idx.Comment,
		CreatedAt:  idx.CreatedAt,
		Options:    make(map[string]interface{}),
	}

	// 克隆选项 Clone options
	for k, v := range idx.Options {
		clone.Options[k] = v
	}

	return clone
}

// Constraint 约束接口 Constraint interface
type Constraint interface {
	GetName() string             // 获取约束名 Get constraint name
	GetType() ConstraintType     // 获取约束类型 Get constraint type
	Validate(value *Value) error // 验证值 Validate value
	ValidateRow(row *Row) error  // 验证行 Validate row
	String() string              // 字符串表示 String representation
}

// ConstraintType 约束类型 Constraint type
type ConstraintType int

const (
	ConstraintTypeNotNull    ConstraintType = iota // 非空约束 Not null constraint
	ConstraintTypeUnique                           // 唯一约束 Unique constraint
	ConstraintTypePrimaryKey                       // 主键约束 Primary key constraint
	ConstraintTypeForeignKey                       // 外键约束 Foreign key constraint
	ConstraintTypeCheck                            // 检查约束 Check constraint
	ConstraintTypeDefault                          // 默认值约束 Default constraint
	ConstraintTypeRange                            // 范围约束 Range constraint
	ConstraintTypeLength                           // 长度约束 Length constraint
	ConstraintTypePattern                          // 模式约束 Pattern constraint
	ConstraintTypeEnum                             // 枚举约束 Enum constraint
)

// String 返回约束类型的字符串表示 Returns string representation of constraint type
func (ct ConstraintType) String() string {
	switch ct {
	case ConstraintTypeNotNull:
		return "NOT_NULL"
	case ConstraintTypeUnique:
		return "UNIQUE"
	case ConstraintTypePrimaryKey:
		return "PRIMARY_KEY"
	case ConstraintTypeForeignKey:
		return "FOREIGN_KEY"
	case ConstraintTypeCheck:
		return "CHECK"
	case ConstraintTypeDefault:
		return "DEFAULT"
	case ConstraintTypeRange:
		return "RANGE"
	case ConstraintTypeLength:
		return "LENGTH"
	case ConstraintTypePattern:
		return "PATTERN"
	case ConstraintTypeEnum:
		return "ENUM"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(ct))
	}
}

// NotNullConstraint 非空约束 Not null constraint
type NotNullConstraint struct {
	Name string `json:"name"` // 约束名 Constraint name
}

func (c *NotNullConstraint) GetName() string {
	return c.Name
}

func (c *NotNullConstraint) GetType() ConstraintType {
	return ConstraintTypeNotNull
}

func (c *NotNullConstraint) Validate(value *Value) error {
	if value == nil || value.IsNull() {
		return fmt.Errorf("value cannot be null")
	}
	return nil
}

func (c *NotNullConstraint) ValidateRow(row *Row) error {
	return nil // 由列级验证处理 Handled by column-level validation
}

func (c *NotNullConstraint) String() string {
	return "NOT NULL"
}

// UniqueConstraint 唯一约束 Unique constraint
type UniqueConstraint struct {
	Name    string   `json:"name"`    // 约束名 Constraint name
	Columns []string `json:"columns"` // 列名列表 Column name list
}

func (c *UniqueConstraint) GetName() string {
	return c.Name
}

func (c *UniqueConstraint) GetType() ConstraintType {
	return ConstraintTypeUnique
}

func (c *UniqueConstraint) Validate(value *Value) error {
	return nil // 由表级验证处理 Handled by table-level validation
}

func (c *UniqueConstraint) ValidateRow(row *Row) error {
	// 这里需要与存储引擎交互检查唯一性
	// Here need to interact with storage engine to check uniqueness
	return nil
}

func (c *UniqueConstraint) String() string {
	return fmt.Sprintf("UNIQUE(%s)", strings.Join(c.Columns, ", "))
}

// CheckConstraint 检查约束 Check constraint
type CheckConstraint struct {
	Name       string `json:"name"`       // 约束名 Constraint name
	Expression string `json:"expression"` // 检查表达式 Check expression
}

func (c *CheckConstraint) GetName() string {
	return c.Name
}

func (c *CheckConstraint) GetType() ConstraintType {
	return ConstraintTypeCheck
}

func (c *CheckConstraint) Validate(value *Value) error {
	// 这里需要实现表达式求值 Here need to implement expression evaluation
	return nil
}

func (c *CheckConstraint) ValidateRow(row *Row) error {
	// 这里需要实现表达式求值 Here need to implement expression evaluation
	return nil
}

func (c *CheckConstraint) String() string {
	return fmt.Sprintf("CHECK(%s)", c.Expression)
}

// Filter 过滤条件 Filter condition
type Filter struct {
	Column   string        `json:"column"`             // 列名 Column name
	Operator FilterOp      `json:"operator"`           // 操作符 Operator
	Value    interface{}   `json:"value"`              // 值 Value
	Values   []interface{} `json:"values,omitempty"`   // 值列表(用于IN操作) Value list (for IN operation)
	Logic    LogicOp       `json:"logic,omitempty"`    // 逻辑操作符 Logic operator
	Children []*Filter     `json:"children,omitempty"` // 子条件 Child conditions
}

// FilterOp 过滤操作符 Filter operator
type FilterOp int

const (
	FilterOpEQ         FilterOp = iota // 等于 Equal
	FilterOpNE                         // 不等于 Not equal
	FilterOpGT                         // 大于 Greater than
	FilterOpGE                         // 大于等于 Greater than or equal
	FilterOpLT                         // 小于 Less than
	FilterOpLE                         // 小于等于 Less than or equal
	FilterOpLike                       // 模糊匹配 Like
	FilterOpNotLike                    // 不匹配 Not like
	FilterOpIn                         // 包含 In
	FilterOpNotIn                      // 不包含 Not in
	FilterOpIsNull                     // 为空 Is null
	FilterOpIsNotNull                  // 不为空 Is not null
	FilterOpBetween                    // 介于之间 Between
	FilterOpNotBetween                 // 不在之间 Not between
	FilterOpRegex                      // 正则匹配 Regex match
	FilterOpNotRegex                   // 正则不匹配 Regex not match
)

// String 返回过滤操作符的字符串表示 Returns string representation of filter operator
func (op FilterOp) String() string {
	switch op {
	case FilterOpEQ:
		return "="
	case FilterOpNE:
		return "!="
	case FilterOpGT:
		return ">"
	case FilterOpGE:
		return ">="
	case FilterOpLT:
		return "<"
	case FilterOpLE:
		return "<="
	case FilterOpLike:
		return "LIKE"
	case FilterOpNotLike:
		return "NOT LIKE"
	case FilterOpIn:
		return "IN"
	case FilterOpNotIn:
		return "NOT IN"
	case FilterOpIsNull:
		return "IS NULL"
	case FilterOpIsNotNull:
		return "IS NOT NULL"
	case FilterOpBetween:
		return "BETWEEN"
	case FilterOpNotBetween:
		return "NOT BETWEEN"
	case FilterOpRegex:
		return "REGEX"
	case FilterOpNotRegex:
		return "NOT REGEX"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(op))
	}
}

// LogicOp 逻辑操作符 Logic operator
type LogicOp int

const (
	LogicOpAnd LogicOp = iota // 与 AND
	LogicOpOr                 // 或 OR
	LogicOpNot                // 非 NOT
)

// String 返回逻辑操作符的字符串表示 Returns string representation of logic operator
func (op LogicOp) String() string {
	switch op {
	case LogicOpAnd:
		return "AND"
	case LogicOpOr:
		return "OR"
	case LogicOpNot:
		return "NOT"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(op))
	}
}

// NewFilter 创建新的过滤条件 Create new filter condition
func NewFilter(column string, operator FilterOp, value interface{}) *Filter {
	return &Filter{
		Column:   column,
		Operator: operator,
		Value:    value,
		Children: make([]*Filter, 0),
	}
}

// And 添加AND条件 Add AND condition
func (f *Filter) And(filter *Filter) *Filter {
	if f.Logic == LogicOpOr {
		// 需要重新组织结构 Need to reorganize structure
		return &Filter{
			Logic:    LogicOpAnd,
			Children: []*Filter{f, filter},
		}
	}

	if f.Logic == LogicOpAnd || len(f.Children) == 0 {
		f.Logic = LogicOpAnd
		f.Children = append(f.Children, filter)
		return f
	}

	return &Filter{
		Logic:    LogicOpAnd,
		Children: []*Filter{f, filter},
	}
}

// Or 添加OR条件 Add OR condition
func (f *Filter) Or(filter *Filter) *Filter {
	if f.Logic == LogicOpAnd {
		// 需要重新组织结构 Need to reorganize structure
		return &Filter{
			Logic:    LogicOpOr,
			Children: []*Filter{f, filter},
		}
	}

	if f.Logic == LogicOpOr || len(f.Children) == 0 {
		f.Logic = LogicOpOr
		f.Children = append(f.Children, filter)
		return f
	}

	return &Filter{
		Logic:    LogicOpOr,
		Children: []*Filter{f, filter},
	}
}

// Query 查询结构 Query structure
type Query struct {
	Select   []string   `json:"select,omitempty"`   // 选择的列 Selected columns
	From     string     `json:"from"`               // 源表 Source table
	Where    *Filter    `json:"where,omitempty"`    // 过滤条件 Filter conditions
	GroupBy  []string   `json:"group_by,omitempty"` // 分组列 Group by columns
	Having   *Filter    `json:"having,omitempty"`   // 分组过滤 Having conditions
	OrderBy  []*OrderBy `json:"order_by,omitempty"` // 排序 Order by
	Limit    int        `json:"limit,omitempty"`    // 限制数量 Limit count
	Offset   int        `json:"offset,omitempty"`   // 偏移量 Offset
	Distinct bool       `json:"distinct"`           // 去重 Distinct
}

// OrderBy 排序结构 Order by structure
type OrderBy struct {
	Column string    `json:"column"` // 列名 Column name
	Order  SortOrder `json:"order"`  // 排序方向 Sort direction
}

// SortOrder 排序方向 Sort order
type SortOrder int

const (
	SortOrderAsc  SortOrder = iota // 升序 Ascending
	SortOrderDesc                  // 降序 Descending
)

// String 返回排序方向的字符串表示 Returns string representation of sort order
func (so SortOrder) String() string {
	switch so {
	case SortOrderAsc:
		return "ASC"
	case SortOrderDesc:
		return "DESC"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(so))
	}
}

// NewQuery 创建新查询 Create new query
func NewQuery(from string) *Query {
	return &Query{
		From:    from,
		Select:  make([]string, 0),
		GroupBy: make([]string, 0),
		OrderBy: make([]*OrderBy, 0),
	}
}

// SelectColumns 选择列 Select columns
func (q *Query) SelectColumns(columns ...string) *Query {
	q.Select = append(q.Select, columns...)
	return q
}

// WhereFilter 添加WHERE条件 Add WHERE condition
func (q *Query) WhereFilter(filter *Filter) *Query {
	q.Where = filter
	return q
}

// GroupByColumns 添加分组列 Add group by columns
func (q *Query) GroupByColumns(columns ...string) *Query {
	q.GroupBy = append(q.GroupBy, columns...)
	return q
}

// OrderByColumn 添加排序列 Add order by column
func (q *Query) OrderByColumn(column string, order SortOrder) *Query {
	q.OrderBy = append(q.OrderBy, &OrderBy{Column: column, Order: order})
	return q
}

// LimitOffset 设置限制和偏移 Set limit and offset
func (q *Query) LimitOffset(limit, offset int) *Query {
	q.Limit = limit
	q.Offset = offset
	return q
}

// ResultSet 结果集 Result set
type ResultSet struct {
	Schema   *Schema                `json:"schema"`             // 结果结构 Result schema
	Rows     []*Row                 `json:"rows"`               // 行数据 Row data
	Count    int                    `json:"count"`              // 总行数 Total row count
	Affected int                    `json:"affected"`           // 影响行数 Affected row count
	Duration time.Duration          `json:"duration"`           // 执行时间 Execution time
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 元数据 Metadata
}

// NewResultSet 创建新结果集 Create new result set
func NewResultSet(schema *Schema) *ResultSet {
	return &ResultSet{
		Schema:   schema,
		Rows:     make([]*Row, 0),
		Metadata: make(map[string]interface{}),
	}
}

// AddRow 添加行 Add row
func (rs *ResultSet) AddRow(row *Row) {
	rs.Rows = append(rs.Rows, row)
	rs.Count = len(rs.Rows)
}

// GetRow 获取指定索引的行 Get row at specified index
func (rs *ResultSet) GetRow(index int) *Row {
	if index < 0 || index >= len(rs.Rows) {
		return nil
	}
	return rs.Rows[index]
}

// IsEmpty 检查结果集是否为空 Check if result set is empty
func (rs *ResultSet) IsEmpty() bool {
	return len(rs.Rows) == 0
}

// ToMaps 转换为映射数组 Convert to map array
func (rs *ResultSet) ToMaps() []map[string]interface{} {
	result := make([]map[string]interface{}, len(rs.Rows))
	for i, row := range rs.Rows {
		result[i] = row.ToMap(rs.Schema)
	}
	return result
}

// DatabaseOptions 数据库选项 Database options
type DatabaseOptions struct {
	Encoding   string                 `json:"encoding,omitempty"`   // 编码 Encoding
	Collation  string                 `json:"collation,omitempty"`  // 排序规则 Collation
	TimeZone   string                 `json:"timezone,omitempty"`   // 时区 Time zone
	MaxSize    int64                  `json:"max_size,omitempty"`   // 最大大小 Maximum size
	MaxTables  int                    `json:"max_tables,omitempty"` // 最大表数 Maximum tables
	AutoVacuum bool                   `json:"auto_vacuum"`          // 自动清理 Auto vacuum
	Compressed bool                   `json:"compressed"`           // 是否压缩 Whether compressed
	Encrypted  bool                   `json:"encrypted"`            // 是否加密 Whether encrypted
	Options    map[string]interface{} `json:"options,omitempty"`    // 其他选项 Other options
}

// TransactionOptions 事务选项 Transaction options
type TransactionOptions struct {
	Isolation  IsolationLevel `json:"isolation"`         // 隔离级别 Isolation level
	ReadOnly   bool           `json:"read_only"`         // 只读事务 Read-only transaction
	Timeout    time.Duration  `json:"timeout,omitempty"` // 超时时间 Timeout
	RetryCount int            `json:"retry_count"`       // 重试次数 Retry count
	AutoCommit bool           `json:"auto_commit"`       // 自动提交 Auto commit
}

// IsolationLevel 隔离级别 Isolation level
type IsolationLevel int

const (
	IsolationReadUncommitted IsolationLevel = iota // 读未提交 Read uncommitted
	IsolationReadCommitted                         // 读已提交 Read committed
	IsolationRepeatableRead                        // 可重复读 Repeatable read
	IsolationSerializable                          // 串行化 Serializable
)

// String 返回隔离级别的字符串表示 Returns string representation of isolation level
func (il IsolationLevel) String() string {
	switch il {
	case IsolationReadUncommitted:
		return "READ_UNCOMMITTED"
	case IsolationReadCommitted:
		return "READ_COMMITTED"
	case IsolationRepeatableRead:
		return "REPEATABLE_READ"
	case IsolationSerializable:
		return "SERIALIZABLE"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(il))
	}
}

// ErrorCode 错误代码 Error code
type ErrorCode int

const (
	ErrCodeSuccess          ErrorCode = 0    // 成功 Success
	ErrCodeGeneral          ErrorCode = 1000 // 通用错误 General error
	ErrCodeInvalidParameter ErrorCode = 1001 // 无效参数 Invalid parameter
	ErrCodeNotFound         ErrorCode = 1002 // 未找到 Not found
	ErrCodeAlreadyExists    ErrorCode = 1003 // 已存在 Already exists
	ErrCodePermissionDenied ErrorCode = 1004 // 权限拒绝 Permission denied
	ErrCodeTimeout          ErrorCode = 1005 // 超时 Timeout
	ErrCodeConnectionFailed ErrorCode = 1006 // 连接失败 Connection failed
	ErrCodeInvalidState     ErrorCode = 1007 // 无效状态 Invalid state
	ErrCodeResourceLimit    ErrorCode = 1008 // 资源限制 Resource limit
	ErrCodeDataCorruption   ErrorCode = 1009 // 数据损坏 Data corruption

	// 数据库相关错误 Database related errors
	ErrCodeDatabaseNotFound    ErrorCode = 2000 // 数据库未找到 Database not found
	ErrCodeTableNotFound       ErrorCode = 2001 // 表未找到 Table not found
	ErrCodeColumnNotFound      ErrorCode = 2002 // 列未找到 Column not found
	ErrCodeIndexNotFound       ErrorCode = 2003 // 索引未找到 Index not found
	ErrCodeConstraintViolation ErrorCode = 2004 // 约束违反 Constraint violation
	ErrCodeDuplicateKey        ErrorCode = 2005 // 重复键 Duplicate key
	ErrCodeTypeMismatch        ErrorCode = 2006 // 类型不匹配 Type mismatch
	ErrCodeSyntaxError         ErrorCode = 2007 // 语法错误 Syntax error

	// 事务相关错误 Transaction related errors
	ErrCodeTransactionFailed    ErrorCode = 3000 // 事务失败 Transaction failed
	ErrCodeDeadlock             ErrorCode = 3001 // 死锁 Deadlock
	ErrCodeSerializationFailure ErrorCode = 3002 // 序列化失败 Serialization failure
	ErrCodeLockTimeout          ErrorCode = 3003 // 锁超时 Lock timeout
)

// String 返回错误代码的字符串表示 Returns string representation of error code
func (ec ErrorCode) String() string {
	switch ec {
	case ErrCodeSuccess:
		return "SUCCESS"
	case ErrCodeGeneral:
		return "GENERAL_ERROR"
	case ErrCodeInvalidParameter:
		return "INVALID_PARAMETER"
	case ErrCodeNotFound:
		return "NOT_FOUND"
	case ErrCodeAlreadyExists:
		return "ALREADY_EXISTS"
	case ErrCodePermissionDenied:
		return "PERMISSION_DENIED"
	case ErrCodeTimeout:
		return "TIMEOUT"
	case ErrCodeConnectionFailed:
		return "CONNECTION_FAILED"
	case ErrCodeInvalidState:
		return "INVALID_STATE"
	case ErrCodeResourceLimit:
		return "RESOURCE_LIMIT"
	case ErrCodeDataCorruption:
		return "DATA_CORRUPTION"
	case ErrCodeDatabaseNotFound:
		return "DATABASE_NOT_FOUND"
	case ErrCodeTableNotFound:
		return "TABLE_NOT_FOUND"
	case ErrCodeColumnNotFound:
		return "COLUMN_NOT_FOUND"
	case ErrCodeIndexNotFound:
		return "INDEX_NOT_FOUND"
	case ErrCodeConstraintViolation:
		return "CONSTRAINT_VIOLATION"
	case ErrCodeDuplicateKey:
		return "DUPLICATE_KEY"
	case ErrCodeTypeMismatch:
		return "TYPE_MISMATCH"
	case ErrCodeSyntaxError:
		return "SYNTAX_ERROR"
	case ErrCodeTransactionFailed:
		return "TRANSACTION_FAILED"
	case ErrCodeDeadlock:
		return "DEADLOCK"
	case ErrCodeSerializationFailure:
		return "SERIALIZATION_FAILURE"
	case ErrCodeLockTimeout:
		return "LOCK_TIMEOUT"
	default:
		return fmt.Sprintf("UNKNOWN_ERROR_%d", int(ec))
	}
}

// Statistics 统计信息 Statistics
type Statistics struct {
	TotalSize     int64                          `json:"total_size"`     // 总大小 Total size
	TotalRows     int64                          `json:"total_rows"`     // 总行数 Total rows
	TotalRequests int64                          `json:"total_requests"` // 总请求数 Total requests
	TotalErrors   int64                          `json:"total_errors"`   // 总错误数 Total errors
	Uptime        time.Duration                  `json:"uptime"`         // 运行时间 Uptime
	Databases     map[string]*DatabaseStatistics `json:"databases"`      // 数据库统计 Database statistics
}

// DatabaseStatistics 数据库统计信息 Database statistics
type DatabaseStatistics struct {
	Name       string `json:"name"`        // 数据库名 Database name
	TableCount int    `json:"table_count"` // 表数量 Table count
	RowCount   int64  `json:"row_count"`   // 行数量 Row count
	Size       int64  `json:"size"`        // 大小 Size
	IndexCount int    `json:"index_count"` // 索引数量 Index count
}
