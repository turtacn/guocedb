// Package encoding provides internal data encoding/decoding functionality for GuoceDB
// 包 encoding 为GuoceDB提供内部数据编码/解码功能
package encoding

import (
	"bytes"
	"compress/gzip"
	"compress/lz4"
	"compress/zlib"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/types"
)

// 编码格式常量 Encoding format constants
const (
	// 魔数 Magic numbers
	RowDataMagic  = 0x47DB0001 // 行数据魔数 Row data magic
	IndexKeyMagic = 0x47DB0002 // 索引键魔数 Index key magic
	MetadataMagic = 0x47DB0003 // 元数据魔数 Metadata magic
	SchemaMagic   = 0x47DB0004 // 表结构魔数 Schema magic

	// 版本号 Version numbers
	EncodingVersion1 = 1
	EncodingVersion2 = 2
	CurrentVersion   = EncodingVersion2

	// 字节序标识 Byte order markers
	LittleEndianMarker = 0x01
	BigEndianMarker    = 0x02

	// 压缩算法标识 Compression algorithm identifiers
	CompressionNone   = 0x00
	CompressionGzip   = 0x01
	CompressionZlib   = 0x02
	CompressionSnappy = 0x03
	CompressionLZ4    = 0x04
	CompressionZstd   = 0x05

	// 校验和算法标识 Checksum algorithm identifiers
	ChecksumNone   = 0x00
	ChecksumCRC32  = 0x01
	ChecksumMD5    = 0x02
	ChecksumSHA256 = 0x03

	// 数据类型标识 Data type identifiers
	TypeNull     = 0x00
	TypeBool     = 0x01
	TypeInt8     = 0x02
	TypeInt16    = 0x03
	TypeInt32    = 0x04
	TypeInt64    = 0x05
	TypeUint8    = 0x06
	TypeUint16   = 0x07
	TypeUint32   = 0x08
	TypeUint64   = 0x09
	TypeFloat32  = 0x0A
	TypeFloat64  = 0x0B
	TypeString   = 0x0C
	TypeBytes    = 0x0D
	TypeTime     = 0x0E
	TypeJSON     = 0x0F
	TypeArray    = 0x10
	TypeMap      = 0x11
	TypeDecimal  = 0x12
	TypeUUID     = 0x13
	TypeGeoPoint = 0x14
	TypeVector   = 0x15

	// 特殊值标识 Special value identifiers
	ValueNull    = 0x00
	ValueNotNull = 0x01
	ValueDefault = 0x02
	ValueEmpty   = 0x03

	// 长度编码标识 Length encoding identifiers
	LengthFixed1 = 0x01 // 1字节长度 1-byte length
	LengthFixed2 = 0x02 // 2字节长度 2-byte length
	LengthFixed4 = 0x04 // 4字节长度 4-byte length
	LengthFixed8 = 0x08 // 8字节长度 8-byte length
	LengthVar    = 0xFF // 变长编码 Variable length

	// 索引类型标识 Index type identifiers
	IndexTypePrimary   = 0x01 // 主键索引 Primary index
	IndexTypeSecondary = 0x02 // 二级索引 Secondary index
	IndexTypeUnique    = 0x03 // 唯一索引 Unique index
	IndexTypeComposite = 0x04 // 复合索引 Composite index
	IndexTypeFullText  = 0x05 // 全文索引 Full-text index
	IndexTypeVector    = 0x06 // 向量索引 Vector index
	IndexTypeGeo       = 0x07 // 地理索引 Geo index

	// 缓冲区大小 Buffer sizes
	DefaultBufferSize = 4096
	MaxBufferSize     = 1024 * 1024 // 1MB
	MinBufferSize     = 256
)

// ByteOrder 字节序类型 Byte order type
type ByteOrder int

const (
	LittleEndian ByteOrder = iota // 小端序 Little endian
	BigEndian                     // 大端序 Big endian
	NativeEndian                  // 本机字节序 Native endian
)

// CompressionType 压缩类型 Compression type
type CompressionType int

const (
	None   CompressionType = iota // 无压缩 No compression
	Gzip                          // Gzip压缩 Gzip compression
	Zlib                          // Zlib压缩 Zlib compression
	Snappy                        // Snappy压缩 Snappy compression
	LZ4                           // LZ4压缩 LZ4 compression
	Zstd                          // Zstd压缩 Zstd compression
)

// ChecksumType 校验和类型 Checksum type
type ChecksumType int

const (
	NoChecksum ChecksumType = iota // 无校验和 No checksum
	CRC32                          // CRC32校验和 CRC32 checksum
	MD5                            // MD5校验和 MD5 checksum
	SHA256                         // SHA256校验和 SHA256 checksum
)

// EncodingConfig 编码配置 Encoding configuration
type EncodingConfig struct {
	Version          int             `json:"version"`           // 编码版本 Encoding version
	ByteOrder        ByteOrder       `json:"byte_order"`        // 字节序 Byte order
	Compression      CompressionType `json:"compression"`       // 压缩类型 Compression type
	Checksum         ChecksumType    `json:"checksum"`          // 校验和类型 Checksum type
	BufferSize       int             `json:"buffer_size"`       // 缓冲区大小 Buffer size
	EnableCaching    bool            `json:"enable_caching"`    // 启用缓存 Enable caching
	MaxCacheSize     int             `json:"max_cache_size"`    // 最大缓存大小 Max cache size
	CompressionLevel int             `json:"compression_level"` // 压缩级别 Compression level
}

// DefaultEncodingConfig 默认编码配置 Default encoding configuration
var DefaultEncodingConfig = &EncodingConfig{
	Version:          CurrentVersion,
	ByteOrder:        LittleEndian,
	Compression:      Snappy,
	Checksum:         CRC32,
	BufferSize:       DefaultBufferSize,
	EnableCaching:    true,
	MaxCacheSize:     1024 * 1024, // 1MB
	CompressionLevel: 6,
}

// Encoder 编码器接口 Encoder interface
type Encoder interface {
	// 基础编码方法 Basic encoding methods
	EncodeValue(value interface{}) ([]byte, error)     // 编码值 Encode value
	DecodeValue(data []byte, target interface{}) error // 解码值 Decode value

	// 行数据编码 Row data encoding
	EncodeRow(row *Row) ([]byte, error)  // 编码行 Encode row
	DecodeRow(data []byte) (*Row, error) // 解码行 Decode row

	// 索引键编码 Index key encoding
	EncodeIndexKey(key *IndexKey) ([]byte, error)  // 编码索引键 Encode index key
	DecodeIndexKey(data []byte) (*IndexKey, error) // 解码索引键 Decode index key

	// 元数据编码 Metadata encoding
	EncodeMetadata(metadata interface{}) ([]byte, error)  // 编码元数据 Encode metadata
	DecodeMetadata(data []byte, target interface{}) error // 解码元数据 Decode metadata

	// 配置管理 Configuration management
	SetConfig(config *EncodingConfig) // 设置配置 Set configuration
	GetConfig() *EncodingConfig       // 获取配置 Get configuration
}

// BinaryEncoder 二进制编码器 Binary encoder
type BinaryEncoder struct {
	config       *EncodingConfig        // 编码配置 Encoding configuration
	buffer       *bytes.Buffer          // 缓冲区 Buffer
	compressor   Compressor             // 压缩器 Compressor
	checksummer  Checksummer            // 校验器 Checksummer
	typeRegistry map[string]TypeEncoder // 类型注册表 Type registry
	cache        map[string][]byte      // 编码缓存 Encoding cache
	cacheSize    int                    // 缓存大小 Cache size
}

// NewBinaryEncoder 创建二进制编码器 Create binary encoder
func NewBinaryEncoder(config *EncodingConfig) *BinaryEncoder {
	if config == nil {
		config = DefaultEncodingConfig
	}

	encoder := &BinaryEncoder{
		config:       config,
		buffer:       bytes.NewBuffer(make([]byte, 0, config.BufferSize)),
		typeRegistry: make(map[string]TypeEncoder),
		cache:        make(map[string][]byte),
		cacheSize:    0,
	}

	// 初始化压缩器 Initialize compressor
	encoder.compressor = NewCompressor(config.Compression, config.CompressionLevel)

	// 初始化校验器 Initialize checksummer
	encoder.checksummer = NewChecksummer(config.Checksum)

	// 注册基础类型编码器 Register basic type encoders
	encoder.registerBasicTypes()

	return encoder
}

// SetConfig 设置配置 Set configuration
func (e *BinaryEncoder) SetConfig(config *EncodingConfig) {
	e.config = config
	e.compressor = NewCompressor(config.Compression, config.CompressionLevel)
	e.checksummer = NewChecksummer(config.Checksum)

	// 清空缓存 Clear cache
	if !config.EnableCaching {
		e.cache = make(map[string][]byte)
		e.cacheSize = 0
	}
}

// GetConfig 获取配置 Get configuration
func (e *BinaryEncoder) GetConfig() *EncodingConfig {
	return e.config
}

// EncodeValue 编码值 Encode value
func (e *BinaryEncoder) EncodeValue(value interface{}) ([]byte, error) {
	// 检查缓存 Check cache
	if e.config.EnableCaching {
		if cached, exists := e.getCachedEncoding(value); exists {
			return cached, nil
		}
	}

	e.buffer.Reset()

	// 写入头部信息 Write header information
	if err := e.writeHeader(); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	// 编码值 Encode value
	if err := e.encodeValueInternal(value); err != nil {
		return nil, fmt.Errorf("failed to encode value: %w", err)
	}

	// 压缩数据 Compress data
	data := e.buffer.Bytes()
	if e.compressor != nil {
		compressed, err := e.compressor.Compress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to compress data: %w", err)
		}
		data = compressed
	}

	// 计算校验和 Calculate checksum
	if e.checksummer != nil {
		checksum := e.checksummer.Checksum(data)
		data = append(data, checksum...)
	}

	// 缓存结果 Cache result
	if e.config.EnableCaching {
		e.setCachedEncoding(value, data)
	}

	return data, nil
}

// DecodeValue 解码值 Decode value
func (e *BinaryEncoder) DecodeValue(data []byte, target interface{}) error {
	if len(data) == 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "empty data")
	}

	// 验证校验和 Verify checksum
	if e.checksummer != nil {
		checksumSize := e.checksummer.Size()
		if len(data) < checksumSize {
			return errors.NewError(errors.ErrCodeDataCorruption, "data too short for checksum")
		}

		payload := data[:len(data)-checksumSize]
		expectedChecksum := data[len(data)-checksumSize:]
		actualChecksum := e.checksummer.Checksum(payload)

		if !bytes.Equal(expectedChecksum, actualChecksum) {
			return errors.NewError(errors.ErrCodeDataCorruption, "checksum mismatch")
		}

		data = payload
	}

	// 解压数据 Decompress data
	if e.compressor != nil {
		decompressed, err := e.compressor.Decompress(data)
		if err != nil {
			return fmt.Errorf("failed to decompress data: %w", err)
		}
		data = decompressed
	}

	reader := bytes.NewReader(data)

	// 读取头部信息 Read header information
	if err := e.readHeader(reader); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// 解码值 Decode value
	return e.decodeValueInternal(reader, target)
}

// EncodeRow 编码行 Encode row
func (e *BinaryEncoder) EncodeRow(row *Row) ([]byte, error) {
	e.buffer.Reset()

	// 写入行头部 Write row header
	if err := e.writeRowHeader(row); err != nil {
		return nil, fmt.Errorf("failed to write row header: %w", err)
	}

	// 编码列数据 Encode column data
	for i, value := range row.Values {
		// 写入列信息 Write column information
		if err := e.writeColumnInfo(row.Schema.Columns[i]); err != nil {
			return nil, fmt.Errorf("failed to write column info for column %d: %w", i, err)
		}

		// 编码列值 Encode column value
		if err := e.encodeColumnValue(value, row.Schema.Columns[i].Type); err != nil {
			return nil, fmt.Errorf("failed to encode column value for column %d: %w", i, err)
		}
	}

	return e.finalizeEncoding()
}

// DecodeRow 解码行 Decode row
func (e *BinaryEncoder) DecodeRow(data []byte) (*Row, error) {
	if len(data) == 0 {
		return nil, errors.NewError(errors.ErrCodeInvalidParameter, "empty row data")
	}

	// 预处理数据 Preprocess data
	processedData, err := e.preprocessData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess data: %w", err)
	}

	reader := bytes.NewReader(processedData)

	// 读取行头部 Read row header
	rowHeader, err := e.readRowHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read row header: %w", err)
	}

	row := &Row{
		ID:      rowHeader.ID,
		Version: rowHeader.Version,
		Schema:  rowHeader.Schema,
		Values:  make([]interface{}, len(rowHeader.Schema.Columns)),
	}

	// 解码列数据 Decode column data
	for i, column := range rowHeader.Schema.Columns {
		// 读取列信息 Read column information
		columnInfo, err := e.readColumnInfo(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read column info for column %d: %w", i, err)
		}

		// 验证列信息 Verify column information
		if columnInfo.Name != column.Name {
			return nil, fmt.Errorf("column name mismatch: expected %s, got %s", column.Name, columnInfo.Name)
		}

		// 解码列值 Decode column value
		value, err := e.decodeColumnValue(reader, column.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to decode column value for column %d: %w", i, err)
		}

		row.Values[i] = value
	}

	return row, nil
}

// EncodeIndexKey 编码索引键 Encode index key
func (e *BinaryEncoder) EncodeIndexKey(key *IndexKey) ([]byte, error) {
	e.buffer.Reset()

	// 写入索引键头部 Write index key header
	if err := e.writeIndexKeyHeader(key); err != nil {
		return nil, fmt.Errorf("failed to write index key header: %w", err)
	}

	// 编码键值 Encode key values
	for i, value := range key.Values {
		if err := e.encodeKeyValue(value, key.Types[i]); err != nil {
			return nil, fmt.Errorf("failed to encode key value at index %d: %w", i, err)
		}
	}

	// 编码行ID Encode row ID
	if err := e.encodeRowID(key.RowID); err != nil {
		return nil, fmt.Errorf("failed to encode row ID: %w", err)
	}

	return e.finalizeEncoding()
}

// DecodeIndexKey 解码索引键 Decode index key
func (e *BinaryEncoder) DecodeIndexKey(data []byte) (*IndexKey, error) {
	if len(data) == 0 {
		return nil, errors.NewError(errors.ErrCodeInvalidParameter, "empty index key data")
	}

	// 预处理数据 Preprocess data
	processedData, err := e.preprocessData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess data: %w", err)
	}

	reader := bytes.NewReader(processedData)

	// 读取索引键头部 Read index key header
	keyHeader, err := e.readIndexKeyHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read index key header: %w", err)
	}

	key := &IndexKey{
		TableName: keyHeader.TableName,
		IndexName: keyHeader.IndexName,
		Types:     keyHeader.Types,
		Values:    make([]interface{}, len(keyHeader.Types)),
	}

	// 解码键值 Decode key values
	for i, dataType := range keyHeader.Types {
		value, err := e.decodeKeyValue(reader, dataType)
		if err != nil {
			return nil, fmt.Errorf("failed to decode key value at index %d: %w", i, err)
		}
		key.Values[i] = value
	}

	// 解码行ID Decode row ID
	rowID, err := e.decodeRowID(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode row ID: %w", err)
	}
	key.RowID = rowID

	return key, nil
}

// EncodeMetadata 编码元数据 Encode metadata
func (e *BinaryEncoder) EncodeMetadata(metadata interface{}) ([]byte, error) {
	e.buffer.Reset()

	// 写入元数据头部 Write metadata header
	if err := e.writeMetadataHeader(metadata); err != nil {
		return nil, fmt.Errorf("failed to write metadata header: %w", err)
	}

	// 根据元数据类型选择编码方式 Choose encoding method based on metadata type
	switch v := metadata.(type) {
	case *DatabaseMetadata:
		return e.encodeDatabaseMetadata(v)
	case *TableMetadata:
		return e.encodeTableMetadata(v)
	case *IndexMetadata:
		return e.encodeIndexMetadata(v)
	case *SchemaMetadata:
		return e.encodeSchemaMetadata(v)
	default:
		// 使用通用编码 Use generic encoding
		return e.encodeGenericMetadata(metadata)
	}
}

// DecodeMetadata 解码元数据 Decode metadata
func (e *BinaryEncoder) DecodeMetadata(data []byte, target interface{}) error {
	if len(data) == 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter, "empty metadata")
	}

	// 预处理数据 Preprocess data
	processedData, err := e.preprocessData(data)
	if err != nil {
		return fmt.Errorf("failed to preprocess data: %w", err)
	}

	reader := bytes.NewReader(processedData)

	// 读取元数据头部 Read metadata header
	metadataHeader, err := e.readMetadataHeader(reader)
	if err != nil {
		return fmt.Errorf("failed to read metadata header: %w", err)
	}

	// 根据元数据类型选择解码方式 Choose decoding method based on metadata type
	switch metadataHeader.Type {
	case "DatabaseMetadata":
		return e.decodeDatabaseMetadata(reader, target.(*DatabaseMetadata))
	case "TableMetadata":
		return e.decodeTableMetadata(reader, target.(*TableMetadata))
	case "IndexMetadata":
		return e.decodeIndexMetadata(reader, target.(*IndexMetadata))
	case "SchemaMetadata":
		return e.decodeSchemaMetadata(reader, target.(*SchemaMetadata))
	default:
		// 使用通用解码 Use generic decoding
		return e.decodeGenericMetadata(reader, target)
	}
}

// 内部编码方法 Internal encoding methods

// writeHeader 写入头部信息 Write header information
func (e *BinaryEncoder) writeHeader() error {
	// 写入版本号 Write version
	if err := e.writeUint32(uint32(e.config.Version)); err != nil {
		return err
	}

	// 写入字节序 Write byte order
	byteOrderMarker := LittleEndianMarker
	if e.config.ByteOrder == BigEndian {
		byteOrderMarker = BigEndianMarker
	}
	if err := e.writeByte(byte(byteOrderMarker)); err != nil {
		return err
	}

	// 写入压缩类型 Write compression type
	if err := e.writeByte(byte(e.config.Compression)); err != nil {
		return err
	}

	// 写入校验和类型 Write checksum type
	if err := e.writeByte(byte(e.config.Checksum)); err != nil {
		return err
	}

	return nil
}

// readHeader 读取头部信息 Read header information
func (e *BinaryEncoder) readHeader(reader *bytes.Reader) error {
	// 读取版本号 Read version
	version, err := e.readUint32(reader)
	if err != nil {
		return err
	}

	if version != uint32(e.config.Version) {
		return fmt.Errorf("unsupported encoding version: %d", version)
	}

	// 读取字节序 Read byte order
	byteOrderMarker, err := e.readByte(reader)
	if err != nil {
		return err
	}

	expectedMarker := LittleEndianMarker
	if e.config.ByteOrder == BigEndian {
		expectedMarker = BigEndianMarker
	}

	if byteOrderMarker != byte(expectedMarker) {
		return fmt.Errorf("byte order mismatch")
	}

	// 读取压缩类型 Read compression type
	compressionType, err := e.readByte(reader)
	if err != nil {
		return err
	}

	if compressionType != byte(e.config.Compression) {
		return fmt.Errorf("compression type mismatch")
	}

	// 读取校验和类型 Read checksum type
	checksumType, err := e.readByte(reader)
	if err != nil {
		return err
	}

	if checksumType != byte(e.config.Checksum) {
		return fmt.Errorf("checksum type mismatch")
	}

	return nil
}

// encodeValueInternal 内部值编码 Internal value encoding
func (e *BinaryEncoder) encodeValueInternal(value interface{}) error {
	if value == nil {
		return e.writeByte(TypeNull)
	}

	switch v := value.(type) {
	case bool:
		if err := e.writeByte(TypeBool); err != nil {
			return err
		}
		return e.encodeBool(v)

	case int8:
		if err := e.writeByte(TypeInt8); err != nil {
			return err
		}
		return e.writeInt8(v)

	case int16:
		if err := e.writeByte(TypeInt16); err != nil {
			return err
		}
		return e.writeInt16(v)

	case int32:
		if err := e.writeByte(TypeInt32); err != nil {
			return err
		}
		return e.writeInt32(v)

	case int64:
		if err := e.writeByte(TypeInt64); err != nil {
			return err
		}
		return e.writeInt64(v)

	case int:
		if err := e.writeByte(TypeInt64); err != nil {
			return err
		}
		return e.writeInt64(int64(v))

	case uint8:
		if err := e.writeByte(TypeUint8); err != nil {
			return err
		}
		return e.writeUint8(v)

	case uint16:
		if err := e.writeByte(TypeUint16); err != nil {
			return err
		}
		return e.writeUint16(v)

	case uint32:
		if err := e.writeByte(TypeUint32); err != nil {
			return err
		}
		return e.writeUint32(v)

	case uint64:
		if err := e.writeByte(TypeUint64); err != nil {
			return err
		}
		return e.writeUint64(v)

	case uint:
		if err := e.writeByte(TypeUint64); err != nil {
			return err
		}
		return e.writeUint64(uint64(v))

	case float32:
		if err := e.writeByte(TypeFloat32); err != nil {
			return err
		}
		return e.writeFloat32(v)

	case float64:
		if err := e.writeByte(TypeFloat64); err != nil {
			return err
		}
		return e.writeFloat64(v)

	case string:
		if err := e.writeByte(TypeString); err != nil {
			return err
		}
		return e.writeString(v)

	case []byte:
		if err := e.writeByte(TypeBytes); err != nil {
			return err
		}
		return e.writeBytes(v)

	case time.Time:
		if err := e.writeByte(TypeTime); err != nil {
			return err
		}
		return e.writeTime(v)

	case []interface{}:
		if err := e.writeByte(TypeArray); err != nil {
			return err
		}
		return e.encodeArray(v)

	case map[string]interface{}:
		if err := e.writeByte(TypeMap); err != nil {
			return err
		}
		return e.encodeMap(v)

	default:
		// 尝试使用注册的类型编码器 Try using registered type encoder
		typeName := reflect.TypeOf(value).String()
		if typeEncoder, exists := e.typeRegistry[typeName]; exists {
			return typeEncoder.Encode(e, value)
		}

		// 使用反射进行通用编码 Use reflection for generic encoding
		return e.encodeGeneric(value)
	}
}

// decodeValueInternal 内部值解码 Internal value decoding
func (e *BinaryEncoder) decodeValueInternal(reader *bytes.Reader, target interface{}) error {
	// 读取类型标识 Read type identifier
	typeID, err := e.readByte(reader)
	if err != nil {
		return err
	}

	switch typeID {
	case TypeNull:
		return e.setTargetValue(target, nil)

	case TypeBool:
		value, err := e.decodeBool(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeInt8:
		value, err := e.readInt8(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeInt16:
		value, err := e.readInt16(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeInt32:
		value, err := e.readInt32(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeInt64:
		value, err := e.readInt64(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeUint8:
		value, err := e.readUint8(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeUint16:
		value, err := e.readUint16(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeUint32:
		value, err := e.readUint32(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeUint64:
		value, err := e.readUint64(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeFloat32:
		value, err := e.readFloat32(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeFloat64:
		value, err := e.readFloat64(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeString:
		value, err := e.readString(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeBytes:
		value, err := e.readBytes(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeTime:
		value, err := e.readTime(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeArray:
		value, err := e.decodeArray(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeMap:
		value, err := e.decodeMap(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeJSON:
		value, err := e.decodeJSON(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeDecimal:
		value, err := e.decodeDecimal(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeUUID:
		value, err := e.decodeUUID(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeGeoPoint:
		value, err := e.decodeGeoPoint(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	case TypeVector:
		value, err := e.decodeVector(reader)
		if err != nil {
			return err
		}
		return e.setTargetValue(target, value)

	default:
		return fmt.Errorf("unsupported type ID: %d", typeID)
	}
}

// 基础类型编码/解码方法 Basic type encoding/decoding methods

// writeByte 写入字节 Write byte
func (e *BinaryEncoder) writeByte(b byte) error {
	return e.buffer.WriteByte(b)
}

// readByte 读取字节 Read byte
func (e *BinaryEncoder) readByte(reader *bytes.Reader) (byte, error) {
	return reader.ReadByte()
}

// writeInt8 写入8位整数 Write 8-bit integer
func (e *BinaryEncoder) writeInt8(value int8) error {
	return e.buffer.WriteByte(byte(value))
}

// readInt8 读取8位整数 Read 8-bit integer
func (e *BinaryEncoder) readInt8(reader *bytes.Reader) (int8, error) {
	b, err := reader.ReadByte()
	return int8(b), err
}

// writeInt16 写入16位整数 Write 16-bit integer
func (e *BinaryEncoder) writeInt16(value int16) error {
	buf := make([]byte, 2)
	if e.config.ByteOrder == LittleEndian {
		binary.LittleEndian.PutUint16(buf, uint16(value))
	} else {
		binary.BigEndian.PutUint16(buf, uint16(value))
	}
	_, err := e.buffer.Write(buf)
	return err
}

// readInt16 读取16位整数 Read 16-bit integer
func (e *BinaryEncoder) readInt16(reader *bytes.Reader) (int16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return 0, err
	}

	if e.config.ByteOrder == LittleEndian {
		return int16(binary.LittleEndian.Uint16(buf)), nil
	}
	return int16(binary.BigEndian.Uint16(buf)), nil
}

// writeInt32 写入32位整数 Write 32-bit integer
func (e *BinaryEncoder) writeInt32(value int32) error {
	buf := make([]byte, 4)
	if e.config.ByteOrder == LittleEndian {
		binary.LittleEndian.PutUint32(buf, uint32(value))
	} else {
		binary.BigEndian.PutUint32(buf, uint32(value))
	}
	_, err := e.buffer.Write(buf)
	return err
}

// readInt32 读取32位整数 Read 32-bit integer
func (e *BinaryEncoder) readInt32(reader *bytes.Reader) (int32, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return 0, err
	}

	if e.config.ByteOrder == LittleEndian {
		return int32(binary.LittleEndian.Uint32(buf)), nil
	}
	return int32(binary.BigEndian.Uint32(buf)), nil
}

// writeInt64 写入64位整数 Write 64-bit integer
func (e *BinaryEncoder) writeInt64(value int64) error {
	buf := make([]byte, 8)
	if e.config.ByteOrder == LittleEndian {
		binary.LittleEndian.PutUint64(buf, uint64(value))
	} else {
		binary.BigEndian.PutUint64(buf, uint64(value))
	}
	_, err := e.buffer.Write(buf)
	return err
}

// readInt64 读取64位整数 Read 64-bit integer
func (e *BinaryEncoder) readInt64(reader *bytes.Reader) (int64, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return 0, err
	}

	if e.config.ByteOrder == LittleEndian {
		return int64(binary.LittleEndian.Uint64(buf)), nil
	}
	return int64(binary.BigEndian.Uint64(buf)), nil
}

// writeUint8 写入8位无符号整数 Write 8-bit unsigned integer
func (e *BinaryEncoder) writeUint8(value uint8) error {
	return e.buffer.WriteByte(value)
}

// readUint8 读取8位无符号整数 Read 8-bit unsigned integer
func (e *BinaryEncoder) readUint8(reader *bytes.Reader) (uint8, error) {
	return reader.ReadByte()
}

// writeUint16 写入16位无符号整数 Write 16-bit unsigned integer
func (e *BinaryEncoder) writeUint16(value uint16) error {
	buf := make([]byte, 2)
	if e.config.ByteOrder == LittleEndian {
		binary.LittleEndian.PutUint16(buf, value)
	} else {
		binary.BigEndian.PutUint16(buf, value)
	}
	_, err := e.buffer.Write(buf)
	return err
}

// readUint16 读取16位无符号整数 Read 16-bit unsigned integer
func (e *BinaryEncoder) readUint16(reader *bytes.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return 0, err
	}

	if e.config.ByteOrder == LittleEndian {
		return binary.LittleEndian.Uint16(buf), nil
	}
	return binary.BigEndian.Uint16(buf), nil
}

// writeUint32 写入32位无符号整数 Write 32-bit unsigned integer
func (e *BinaryEncoder) writeUint32(value uint32) error {
	buf := make([]byte, 4)
	if e.config.ByteOrder == LittleEndian {
		binary.LittleEndian.PutUint32(buf, value)
	} else {
		binary.BigEndian.PutUint32(buf, value)
	}
	_, err := e.buffer.Write(buf)
	return err
}

// readUint32 读取32位无符号整数 Read 32-bit unsigned integer
func (e *BinaryEncoder) readUint32(reader *bytes.Reader) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return 0, err
	}

	if e.config.ByteOrder == LittleEndian {
		return binary.LittleEndian.Uint32(buf), nil
	}
	return binary.BigEndian.Uint32(buf), nil
}

// writeUint64 写入64位无符号整数 Write 64-bit unsigned integer
func (e *BinaryEncoder) writeUint64(value uint64) error {
	buf := make([]byte, 8)
	if e.config.ByteOrder == LittleEndian {
		binary.LittleEndian.PutUint64(buf, value)
	} else {
		binary.BigEndian.PutUint64(buf, value)
	}
	_, err := e.buffer.Write(buf)
	return err
}

// readUint64 读取64位无符号整数 Read 64-bit unsigned integer
func (e *BinaryEncoder) readUint64(reader *bytes.Reader) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return 0, err
	}

	if e.config.ByteOrder == LittleEndian {
		return binary.LittleEndian.Uint64(buf), nil
	}
	return binary.BigEndian.Uint64(buf), nil
}

// writeFloat32 写入32位浮点数 Write 32-bit float
func (e *BinaryEncoder) writeFloat32(value float32) error {
	return e.writeUint32(math.Float32bits(value))
}

// readFloat32 读取32位浮点数 Read 32-bit float
func (e *BinaryEncoder) readFloat32(reader *bytes.Reader) (float32, error) {
	bits, err := e.readUint32(reader)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

// writeFloat64 写入64位浮点数 Write 64-bit float
func (e *BinaryEncoder) writeFloat64(value float64) error {
	return e.writeUint64(math.Float64bits(value))
}

// readFloat64 读取64位浮点数 Read 64-bit float
func (e *BinaryEncoder) readFloat64(reader *bytes.Reader) (float64, error) {
	bits, err := e.readUint64(reader)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

// encodeBool 编码布尔值 Encode boolean
func (e *BinaryEncoder) encodeBool(value bool) error {
	if value {
		return e.writeByte(1)
	}
	return e.writeByte(0)
}

// decodeBool 解码布尔值 Decode boolean
func (e *BinaryEncoder) decodeBool(reader *bytes.Reader) (bool, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

// writeString 写入字符串 Write string
func (e *BinaryEncoder) writeString(value string) error {
	data := []byte(value)
	if err := e.writeVarInt(uint64(len(data))); err != nil {
		return err
	}
	_, err := e.buffer.Write(data)
	return err
}

// readString 读取字符串 Read string
func (e *BinaryEncoder) readString(reader *bytes.Reader) (string, error) {
	length, err := e.readVarInt(reader)
	if err != nil {
		return "", err
	}

	if length > uint64(MaxBufferSize) {
		return "", fmt.Errorf("string too long: %d bytes", length)
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return "", err
	}

	return string(buf), nil
}

// writeBytes 写入字节数组 Write byte array
func (e *BinaryEncoder) writeBytes(value []byte) error {
	if err := e.writeVarInt(uint64(len(value))); err != nil {
		return err
	}
	_, err := e.buffer.Write(value)
	return err
}

// readBytes 读取字节数组 Read byte array
func (e *BinaryEncoder) readBytes(reader *bytes.Reader) ([]byte, error) {
	length, err := e.readVarInt(reader)
	if err != nil {
		return nil, err
	}

	if length > uint64(MaxBufferSize) {
		return nil, fmt.Errorf("byte array too long: %d bytes", length)
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

// writeTime 写入时间 Write time
func (e *BinaryEncoder) writeTime(value time.Time) error {
	// 写入Unix纳秒时间戳 Write Unix nanosecond timestamp
	timestamp := value.UnixNano()
	if err := e.writeInt64(timestamp); err != nil {
		return err
	}

	// 写入时区信息 Write timezone information
	_, offset := value.Zone()
	return e.writeInt32(int32(offset))
}

// readTime 读取时间 Read time
func (e *BinaryEncoder) readTime(reader *bytes.Reader) (time.Time, error) {
	// 读取Unix纳秒时间戳 Read Unix nanosecond timestamp
	timestamp, err := e.readInt64(reader)
	if err != nil {
		return time.Time{}, err
	}

	// 读取时区偏移 Read timezone offset
	offset, err := e.readInt32(reader)
	if err != nil {
		return time.Time{}, err
	}

	// 创建带时区的时间 Create time with timezone
	loc := time.FixedZone("", int(offset))
	return time.Unix(0, timestamp).In(loc), nil
}

// writeVarInt 写入变长整数 Write variable-length integer
func (e *BinaryEncoder) writeVarInt(value uint64) error {
	for value >= 0x80 {
		if err := e.writeByte(byte(value) | 0x80); err != nil {
			return err
		}
		value >>= 7
	}
	return e.writeByte(byte(value))
}

// readVarInt 读取变长整数 Read variable-length integer
func (e *BinaryEncoder) readVarInt(reader *bytes.Reader) (uint64, error) {
	var result uint64
	var shift uint

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		result |= uint64(b&0x7F) << shift
		shift += 7

		if b&0x80 == 0 {
			break
		}

		if shift >= 64 {
			return 0, fmt.Errorf("varint overflow")
		}
	}

	return result, nil
}

// 复合类型编码/解码方法 Composite type encoding/decoding methods

// encodeArray 编码数组 Encode array
func (e *BinaryEncoder) encodeArray(value []interface{}) error {
	// 写入数组长度 Write array length
	if err := e.writeVarInt(uint64(len(value))); err != nil {
		return err
	}

	// 编码每个元素 Encode each element
	for i, element := range value {
		if err := e.encodeValueInternal(element); err != nil {
			return fmt.Errorf("failed to encode array element at index %d: %w", i, err)
		}
	}

	return nil
}

// decodeArray 解码数组 Decode array
func (e *BinaryEncoder) decodeArray(reader *bytes.Reader) ([]interface{}, error) {
	// 读取数组长度 Read array length
	length, err := e.readVarInt(reader)
	if err != nil {
		return nil, err
	}

	if length > uint64(MaxBufferSize/8) { // 假设每个元素至少8字节 Assume at least 8 bytes per element
		return nil, fmt.Errorf("array too long: %d elements", length)
	}

	// 解码每个元素 Decode each element
	result := make([]interface{}, length)
	for i := uint64(0); i < length; i++ {
		var element interface{}
		if err := e.decodeValueInternal(reader, &element); err != nil {
			return nil, fmt.Errorf("failed to decode array element at index %d: %w", i, err)
		}
		result[i] = element
	}

	return result, nil
}

// encodeMap 编码映射 Encode map
func (e *BinaryEncoder) encodeMap(value map[string]interface{}) error {
	// 写入映射长度 Write map length
	if err := e.writeVarInt(uint64(len(value))); err != nil {
		return err
	}

	// 对键进行排序以确保一致性 Sort keys for consistency
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 编码每个键值对 Encode each key-value pair
	for _, key := range keys {
		// 编码键 Encode key
		if err := e.writeString(key); err != nil {
			return fmt.Errorf("failed to encode map key %s: %w", key, err)
		}

		// 编码值 Encode value
		if err := e.encodeValueInternal(value[key]); err != nil {
			return fmt.Errorf("failed to encode map value for key %s: %w", key, err)
		}
	}

	return nil
}

// decodeMap 解码映射 Decode map
func (e *BinaryEncoder) decodeMap(reader *bytes.Reader) (map[string]interface{}, error) {
	// 读取映射长度 Read map length
	length, err := e.readVarInt(reader)
	if err != nil {
		return nil, err
	}

	if length > uint64(MaxBufferSize/16) { // 假设每个键值对至少16字节 Assume at least 16 bytes per key-value pair
		return nil, fmt.Errorf("map too long: %d entries", length)
	}

	// 解码每个键值对 Decode each key-value pair
	result := make(map[string]interface{}, length)
	for i := uint64(0); i < length; i++ {
		// 解码键 Decode key
		key, err := e.readString(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode map key at index %d: %w", i, err)
		}

		// 解码值 Decode value
		var value interface{}
		if err := e.decodeValueInternal(reader, &value); err != nil {
			return nil, fmt.Errorf("failed to decode map value for key %s: %w", key, err)
		}

		result[key] = value
	}

	return result, nil
}

// 行数据编码相关方法 Row data encoding related methods

// writeRowHeader 写入行头部 Write row header
func (e *BinaryEncoder) writeRowHeader(row *Row) error {
	// 写入魔数 Write magic number
	if err := e.writeUint32(RowDataMagic); err != nil {
		return err
	}

	// 写入行ID Write row ID
	if err := e.writeString(row.ID); err != nil {
		return err
	}

	// 写入版本号 Write version
	if err := e.writeUint64(row.Version); err != nil {
		return err
	}

	// 写入列数量 Write column count
	if err := e.writeVarInt(uint64(len(row.Values))); err != nil {
		return err
	}

	// 写入表结构信息 Write schema information
	return e.encodeSchema(row.Schema)
}

// readRowHeader 读取行头部 Read row header
func (e *BinaryEncoder) readRowHeader(reader *bytes.Reader) (*RowHeader, error) {
	// 读取魔数 Read magic number
	magic, err := e.readUint32(reader)
	if err != nil {
		return nil, err
	}

	if magic != RowDataMagic {
		return nil, fmt.Errorf("invalid row data magic: %x", magic)
	}

	// 读取行ID Read row ID
	id, err := e.readString(reader)
	if err != nil {
		return nil, err
	}

	// 读取版本号 Read version
	version, err := e.readUint64(reader)
	if err != nil {
		return nil, err
	}

	// 读取列数量 Read column count
	columnCount, err := e.readVarInt(reader)
	if err != nil {
		return nil, err
	}

	// 读取表结构信息 Read schema information
	schema, err := e.decodeSchema(reader)
	if err != nil {
		return nil, err
	}

	return &RowHeader{
		ID:          id,
		Version:     version,
		ColumnCount: columnCount,
		Schema:      schema,
	}, nil
}

// writeColumnInfo 写入列信息 Write column information
func (e *BinaryEncoder) writeColumnInfo(column *types.Column) error {
	// 写入列名 Write column name
	if err := e.writeString(column.Name); err != nil {
		return err
	}

	// 写入列类型 Write column type
	if err := e.writeByte(byte(column.Type)); err != nil {
		return err
	}

	// 写入是否可空 Write nullable flag
	return e.encodeBool(column.Nullable)
}

// readColumnInfo 读取列信息 Read column information
func (e *BinaryEncoder) readColumnInfo(reader *bytes.Reader) (*ColumnInfo, error) {
	// 读取列名 Read column name
	name, err := e.readString(reader)
	if err != nil {
		return nil, err
	}

	// 读取列类型 Read column type
	typeID, err := e.readByte(reader)
	if err != nil {
		return nil, err
	}

	// 读取是否可空 Read nullable flag
	nullable, err := e.decodeBool(reader)
	if err != nil {
		return nil, err
	}

	return &ColumnInfo{
		Name:     name,
		Type:     types.DataType(typeID),
		Nullable: nullable,
	}, nil
}

// encodeColumnValue 编码列值 Encode column value
func (e *BinaryEncoder) encodeColumnValue(value interface{}, dataType types.DataType) error {
	// 处理NULL值 Handle NULL values
	if value == nil {
		return e.writeByte(ValueNull)
	}

	// 写入非NULL标记 Write non-NULL marker
	if err := e.writeByte(ValueNotNull); err != nil {
		return err
	}

	// 根据数据类型编码值 Encode value based on data type
	switch dataType {
	case types.DataTypeBool:
		return e.encodeBool(value.(bool))
	case types.DataTypeInt8:
		return e.writeInt8(value.(int8))
	case types.DataTypeInt16:
		return e.writeInt16(value.(int16))
	case types.DataTypeInt32:
		return e.writeInt32(value.(int32))
	case types.DataTypeInt64:
		return e.writeInt64(value.(int64))
	case types.DataTypeUint8:
		return e.writeUint8(value.(uint8))
	case types.DataTypeUint16:
		return e.writeUint16(value.(uint16))
	case types.DataTypeUint32:
		return e.writeUint32(value.(uint32))
	case types.DataTypeUint64:
		return e.writeUint64(value.(uint64))
	case types.DataTypeFloat32:
		return e.writeFloat32(value.(float32))
	case types.DataTypeFloat64:
		return e.writeFloat64(value.(float64))
	case types.DataTypeString:
		return e.writeString(value.(string))
	case types.DataTypeBytes:
		return e.writeBytes(value.([]byte))
	case types.DataTypeTime:
		return e.writeTime(value.(time.Time))
	case types.DataTypeJSON:
		return e.encodeJSON(value)
	case types.DataTypeDecimal:
		return e.encodeDecimal(value)
	case types.DataTypeUUID:
		return e.encodeUUID(value)
	case types.DataTypeGeoPoint:
		return e.encodeGeoPoint(value)
	case types.DataTypeVector:
		return e.encodeVector(value)
	default:
		return fmt.Errorf("unsupported data type: %v", dataType)
	}
}

// decodeColumnValue 解码列值 Decode column value
func (e *BinaryEncoder) decodeColumnValue(reader *bytes.Reader, dataType types.DataType) (interface{}, error) {
	// 读取NULL标记 Read NULL marker
	nullMarker, err := e.readByte(reader)
	if err != nil {
		return nil, err
	}

	// 处理NULL值 Handle NULL values
	if nullMarker == ValueNull {
		return nil, nil
	}

	if nullMarker != ValueNotNull {
		return nil, fmt.Errorf("invalid null marker: %d", nullMarker)
	}

	// 根据数据类型解码值 Decode value based on data type
	switch dataType {
	case types.DataTypeBool:
		return e.decodeBool(reader)
	case types.DataTypeInt8:
		return e.readInt8(reader)
	case types.DataTypeInt16:
		return e.readInt16(reader)
	case types.DataTypeInt32:
		return e.readInt32(reader)
	case types.DataTypeInt64:
		return e.readInt64(reader)
	case types.DataTypeUint8:
		return e.readUint8(reader)
	case types.DataTypeUint16:
		return e.readUint16(reader)
	case types.DataTypeUint32:
		return e.readUint32(reader)
	case types.DataTypeUint64:
		return e.readUint64(reader)
	case types.DataTypeFloat32:
		return e.readFloat32(reader)
	case types.DataTypeFloat64:
		return e.readFloat64(reader)
	case types.DataTypeString:
		return e.readString(reader)
	case types.DataTypeBytes:
		return e.readBytes(reader)
	case types.DataTypeTime:
		return e.readTime(reader)
	case types.DataTypeJSON:
		return e.decodeJSON(reader)
	case types.DataTypeDecimal:
		return e.decodeDecimal(reader)
	case types.DataTypeUUID:
		return e.decodeUUID(reader)
	case types.DataTypeGeoPoint:
		return e.decodeGeoPoint(reader)
	case types.DataTypeVector:
		return e.decodeVector(reader)
	default:
		return nil, fmt.Errorf("unsupported data type: %v", dataType)
	}
}

// 索引键编码相关方法 Index key encoding related methods

// writeIndexKeyHeader 写入索引键头部 Write index key header
func (e *BinaryEncoder) writeIndexKeyHeader(key *IndexKey) error {
	// 写入魔数 Write magic number
	if err := e.writeUint32(IndexKeyMagic); err != nil {
		return err
	}

	// 写入表名 Write table name
	if err := e.writeString(key.TableName); err != nil {
		return err
	}

	// 写入索引名 Write index name
	if err := e.writeString(key.IndexName); err != nil {
		return err
	}

	// 写入索引类型 Write index type
	if err := e.writeByte(byte(key.IndexType)); err != nil {
		return err
	}

	// 写入键值数量 Write key value count
	if err := e.writeVarInt(uint64(len(key.Values))); err != nil {
		return err
	}

	// 写入每个键值的类型 Write type for each key value
	for _, dataType := range key.Types {
		if err := e.writeByte(byte(dataType)); err != nil {
			return err
		}
	}

	return nil
}

// readIndexKeyHeader 读取索引键头部 Read index key header
func (e *BinaryEncoder) readIndexKeyHeader(reader *bytes.Reader) (*IndexKeyHeader, error) {
	// 读取魔数 Read magic number
	magic, err := e.readUint32(reader)
	if err != nil {
		return nil, err
	}

	if magic != IndexKeyMagic {
		return nil, fmt.Errorf("invalid index key magic: %x", magic)
	}

	// 读取表名 Read table name
	tableName, err := e.readString(reader)
	if err != nil {
		return nil, err
	}

	// 读取索引名 Read index name
	indexName, err := e.readString(reader)
	if err != nil {
		return nil, err
	}

	// 读取索引类型 Read index type
	indexType, err := e.readByte(reader)
	if err != nil {
		return nil, err
	}

	// 读取键值数量 Read key value count
	keyCount, err := e.readVarInt(reader)
	if err != nil {
		return nil, err
	}

	// 读取每个键值的类型 Read type for each key value
	types := make([]types.DataType, keyCount)
	for i := uint64(0); i < keyCount; i++ {
		typeID, err := e.readByte(reader)
		if err != nil {
			return nil, err
		}
		types[i] = types.DataType(typeID)
	}

	return &IndexKeyHeader{
		TableName: tableName,
		IndexName: indexName,
		IndexType: IndexType(indexType),
		KeyCount:  keyCount,
		Types:     types,
	}, nil
}

// encodeKeyValue 编码键值 Encode key value
func (e *BinaryEncoder) encodeKeyValue(value interface{}, dataType types.DataType) error {
	return e.encodeColumnValue(value, dataType)
}

// decodeKeyValue 解码键值 Decode key value
func (e *BinaryEncoder) decodeKeyValue(reader *bytes.Reader, dataType types.DataType) (interface{}, error) {
	return e.decodeColumnValue(reader, dataType)
}

// encodeRowID 编码行ID Encode row ID
func (e *BinaryEncoder) encodeRowID(rowID string) error {
	return e.writeString(rowID)
}

// decodeRowID 解码行ID Decode row ID
func (e *BinaryEncoder) decodeRowID(reader *bytes.Reader) (string, error) {
	return e.readString(reader)
}

// 元数据编码相关方法 Metadata encoding related methods

// writeMetadataHeader 写入元数据头部 Write metadata header
func (e *BinaryEncoder) writeMetadataHeader(metadata interface{}) error {
	// 写入魔数 Write magic number
	if err := e.writeUint32(MetadataMagic); err != nil {
		return err
	}

	// 写入元数据类型 Write metadata type
	metadataType := reflect.TypeOf(metadata).String()
	return e.writeString(metadataType)
}

// readMetadataHeader 读取元数据头部 Read metadata header
func (e *BinaryEncoder) readMetadataHeader(reader *bytes.Reader) (*MetadataHeader, error) {
	// 读取魔数 Read magic number
	magic, err := e.readUint32(reader)
	if err != nil {
		return nil, err
	}

	if magic != MetadataMagic {
		return nil, fmt.Errorf("invalid metadata magic: %x", magic)
	}

	// 读取元数据类型 Read metadata type
	metadataType, err := e.readString(reader)
	if err != nil {
		return nil, err
	}

	return &MetadataHeader{
		Type: metadataType,
	}, nil
}

// encodeDatabaseMetadata 编码数据库元数据 Encode database metadata
func (e *BinaryEncoder) encodeDatabaseMetadata(metadata *DatabaseMetadata) ([]byte, error) {
	// 编码数据库名 Encode database name
	if err := e.writeString(metadata.Name); err != nil {
		return nil, err
	}

	// 编码创建时间 Encode creation time
	if err := e.writeTime(metadata.CreatedAt); err != nil {
		return nil, err
	}

	// 编码修改时间 Encode modification time
	if err := e.writeTime(metadata.UpdatedAt); err != nil {
		return nil, err
	}

	// 编码表数量 Encode table count
	if err := e.writeVarInt(uint64(len(metadata.Tables))); err != nil {
		return nil, err
	}

	// 编码表名列表 Encode table name list
	for _, tableName := range metadata.Tables {
		if err := e.writeString(tableName); err != nil {
			return nil, err
		}
	}

	// 编码选项 Encode options
	return e.finalizeEncoding()
}

// decodeDatabaseMetadata 解码数据库元数据 Decode database metadata
func (e *BinaryEncoder) decodeDatabaseMetadata(reader *bytes.Reader, metadata *DatabaseMetadata) error {
	// 解码数据库名 Decode database name
	name, err := e.readString(reader)
	if err != nil {
		return err
	}
	metadata.Name = name

	// 解码创建时间 Decode creation time
	createdAt, err := e.readTime(reader)
	if err != nil {
		return err
	}
	metadata.CreatedAt = createdAt

	// 解码修改时间 Decode modification time
	updatedAt, err := e.readTime(reader)
	if err != nil {
		return err
	}
	metadata.UpdatedAt = updatedAt

	// 解码表数量 Decode table count
	tableCount, err := e.readVarInt(reader)
	if err != nil {
		return err
	}

	// 解码表名列表 Decode table name list
	tables := make([]string, tableCount)
	for i := uint64(0); i < tableCount; i++ {
		tableName, err := e.readString(reader)
		if err != nil {
			return err
		}
		tables[i] = tableName
	}
	metadata.Tables = tables

	return nil
}

// encodeTableMetadata 编码表元数据 Encode table metadata
func (e *BinaryEncoder) encodeTableMetadata(metadata *TableMetadata) ([]byte, error) {
	// 编码表名 Encode table name
	if err := e.writeString(metadata.Name); err != nil {
		return nil, err
	}

	// 编码数据库名 Encode database name
	if err := e.writeString(metadata.Database); err != nil {
		return nil, err
	}

	// 编码表结构 Encode table schema
	if err := e.encodeSchema(metadata.Schema); err != nil {
		return nil, err
	}

	// 编码索引信息 Encode index information
	if err := e.writeVarInt(uint64(len(metadata.Indexes))); err != nil {
		return nil, err
	}

	for _, index := range metadata.Indexes {
		if err := e.encodeIndexMetadata(index); err != nil {
			return nil, err
		}
	}

	// 编码统计信息 Encode statistics
	if err := e.writeUint64(metadata.RowCount); err != nil {
		return nil, err
	}

	if err := e.writeUint64(metadata.Size); err != nil {
		return nil, err
	}

	// 编码时间信息 Encode time information
	if err := e.writeTime(metadata.CreatedAt); err != nil {
		return nil, err
	}

	if err := e.writeTime(metadata.UpdatedAt); err != nil {
		return nil, err
	}

	return e.finalizeEncoding()
}

// decodeTableMetadata 解码表元数据 Decode table metadata
func (e *BinaryEncoder) decodeTableMetadata(reader *bytes.Reader, metadata *TableMetadata) error {
	// 解码表名 Decode table name
	name, err := e.readString(reader)
	if err != nil {
		return err
	}
	metadata.Name = name

	// 解码数据库名 Decode database name
	database, err := e.readString(reader)
	if err != nil {
		return err
	}
	metadata.Database = database

	// 解码表结构 Decode table schema
	schema, err := e.decodeSchema(reader)
	if err != nil {
		return err
	}
	metadata.Schema = schema

	// 解码索引信息 Decode index information
	indexCount, err := e.readVarInt(reader)
	if err != nil {
		return err
	}

	indexes := make([]*IndexMetadata, indexCount)
	for i := uint64(0); i < indexCount; i++ {
		index := &IndexMetadata{}
		if err := e.decodeIndexMetadata(reader, index); err != nil {
			return err
		}
		indexes[i] = index
	}
	metadata.Indexes = indexes

	// 解码统计信息 Decode statistics
	rowCount, err := e.readUint64(reader)
	if err != nil {
		return err
	}
	metadata.RowCount = rowCount

	size, err := e.readUint64(reader)
	if err != nil {
		return err
	}
	metadata.Size = size

	// 解码时间信息 Decode time information
	createdAt, err := e.readTime(reader)
	if err != nil {
		return err
	}
	metadata.CreatedAt = createdAt

	updatedAt, err := e.readTime(reader)
	if err != nil {
		return err
	}
	metadata.UpdatedAt = updatedAt

	return nil
}

// encodeIndexMetadata 编码索引元数据 Encode index metadata
func (e *BinaryEncoder) encodeIndexMetadata(metadata *IndexMetadata) error {
	// 编码索引名 Encode index name
	if err := e.writeString(metadata.Name); err != nil {
		return err
	}

	// 编码索引类型 Encode index type
	if err := e.writeByte(byte(metadata.Type)); err != nil {
		return err
	}

	// 编码列名列表 Encode column name list
	if err := e.writeVarInt(uint64(len(metadata.Columns))); err != nil {
		return err
	}

	for _, column := range metadata.Columns {
		if err := e.writeString(column); err != nil {
			return err
		}
	}

	// 编码是否唯一 Encode unique flag
	if err := e.encodeBool(metadata.Unique); err != nil {
		return err
	}

	// 编码创建时间 Encode creation time
	return e.writeTime(metadata.CreatedAt)
}

// decodeIndexMetadata 解码索引元数据 Decode index metadata
func (e *BinaryEncoder) decodeIndexMetadata(reader *bytes.Reader, metadata *IndexMetadata) error {
	// 解码索引名 Decode index name
	name, err := e.readString(reader)
	if err != nil {
		return err
	}
	metadata.Name = name

	// 解码索引类型 Decode index type
	indexType, err := e.readByte(reader)
	if err != nil {
		return err
	}
	metadata.Type = IndexType(indexType)

	// 解码列名列表 Decode column name list
	columnCount, err := e.readVarInt(reader)
	if err != nil {
		return err
	}

	columns := make([]string, columnCount)
	for i := uint64(0); i < columnCount; i++ {
		column, err := e.readString(reader)
		if err != nil {
			return err
		}
		columns[i] = column
	}
	metadata.Columns = columns

	// 解码是否唯一 Decode unique flag
	unique, err := e.decodeBool(reader)
	if err != nil {
		return err
	}
	metadata.Unique = unique

	// 解码创建时间 Decode creation time
	createdAt, err := e.readTime(reader)
	if err != nil {
		return err
	}
	metadata.CreatedAt = createdAt

	return nil
}

// encodeSchemaMetadata 编码表结构元数据 Encode schema metadata
func (e *BinaryEncoder) encodeSchemaMetadata(metadata *SchemaMetadata) ([]byte, error) {
	return e.encodeSchema(metadata.Schema)
}

// decodeSchemaMetadata 解码表结构元数据 Decode schema metadata
func (e *BinaryEncoder) decodeSchemaMetadata(reader *bytes.Reader, metadata *SchemaMetadata) error {
	schema, err := e.decodeSchema(reader)
	if err != nil {
		return err
	}
	metadata.Schema = schema
	return nil
}

// encodeGenericMetadata 编码通用元数据 Encode generic metadata
func (e *BinaryEncoder) encodeGenericMetadata(metadata interface{}) ([]byte, error) {
	// 使用JSON编码作为后备方案 Use JSON encoding as fallback
	data, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata to JSON: %w", err)
	}

	if err := e.writeBytes(data); err != nil {
		return nil, err
	}

	return e.finalizeEncoding()
}

// decodeGenericMetadata 解码通用元数据 Decode generic metadata
func (e *BinaryEncoder) decodeGenericMetadata(reader *bytes.Reader, target interface{}) error {
	// 使用JSON解码作为后备方案 Use JSON decoding as fallback
	data, err := e.readBytes(reader)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// 表结构编码相关方法 Schema encoding related methods

// encodeSchema 编码表结构 Encode schema
func (e *BinaryEncoder) encodeSchema(schema *types.Schema) error {
	// 写入魔数 Write magic number
	if err := e.writeUint32(SchemaMagic); err != nil {
		return err
	}

	// 写入版本号 Write version
	if err := e.writeUint32(schema.Version); err != nil {
		return err
	}

	// 写入列数量 Write column count
	if err := e.writeVarInt(uint64(len(schema.Columns))); err != nil {
		return err
	}

	// 编码每一列 Encode each column
	for _, column := range schema.Columns {
		if err := e.encodeColumn(column); err != nil {
			return err
		}
	}

	// 编码主键信息 Encode primary key information
	if err := e.writeVarInt(uint64(len(schema.PrimaryKey))); err != nil {
		return err
	}

	for _, pkColumn := range schema.PrimaryKey {
		if err := e.writeString(pkColumn); err != nil {
			return err
		}
	}

	return nil
}

// decodeSchema 解码表结构 Decode schema
func (e *BinaryEncoder) decodeSchema(reader *bytes.Reader) (*types.Schema, error) {
	// 读取魔数 Read magic number
	magic, err := e.readUint32(reader)
	if err != nil {
		return nil, err
	}

	if magic != SchemaMagic {
		return nil, fmt.Errorf("invalid schema magic: %x", magic)
	}

	// 读取版本号 Read version
	version, err := e.readUint32(reader)
	if err != nil {
		return nil, err
	}

	// 读取列数量 Read column count
	columnCount, err := e.readVarInt(reader)
	if err != nil {
		return nil, err
	}

	// 解码每一列 Decode each column
	columns := make([]*types.Column, columnCount)
	for i := uint64(0); i < columnCount; i++ {
		column, err := e.decodeColumn(reader)
		if err != nil {
			return nil, err
		}
		columns[i] = column
	}

	// 解码主键信息 Decode primary key information
	pkCount, err := e.readVarInt(reader)
	if err != nil {
		return nil, err
	}

	primaryKey := make([]string, pkCount)
	for i := uint64(0); i < pkCount; i++ {
		pkColumn, err := e.readString(reader)
		if err != nil {
			return nil, err
		}
		primaryKey[i] = pkColumn
	}

	return &types.Schema{
		Version:    version,
		Columns:    columns,
		PrimaryKey: primaryKey,
	}, nil
}

// encodeColumn 编码列定义 Encode column definition
func (e *BinaryEncoder) encodeColumn(column *types.Column) error {
	// 编码列名 Encode column name
	if err := e.writeString(column.Name); err != nil {
		return err
	}

	// 编码列类型 Encode column type
	if err := e.writeByte(byte(column.Type)); err != nil {
		return err
	}

	// 编码长度/精度 Encode length/precision
	if err := e.writeUint32(column.Length); err != nil {
		return err
	}

	if err := e.writeUint32(column.Precision); err != nil {
		return err
	}

	if err := e.writeUint32(column.Scale); err != nil {
		return err
	}

	// 编码约束信息 Encode constraint information
	if err := e.encodeBool(column.Nullable); err != nil {
		return err
	}

	if err := e.encodeBool(column.AutoIncrement); err != nil {
		return err
	}

	if err := e.encodeBool(column.Unique); err != nil {
		return err
	}

	// 编码默认值 Encode default value
	if column.DefaultValue != nil {
		if err := e.writeByte(ValueNotNull); err != nil {
			return err
		}
		if err := e.encodeValueInternal(column.DefaultValue); err != nil {
			return err
		}
	} else {
		if err := e.writeByte(ValueNull); err != nil {
			return err
		}
	}

	// 编码注释 Encode comment
	return e.writeString(column.Comment)
}

// decodeColumn 解码列定义 Decode column definition
func (e *BinaryEncoder) decodeColumn(reader *bytes.Reader) (*types.Column, error) {
	// 解码列名 Decode column name
	name, err := e.readString(reader)
	if err != nil {
		return nil, err
	}

	// 解码列类型 Decode column type
	typeID, err := e.readByte(reader)
	if err != nil {
		return nil, err
	}

	// 解码长度/精度 Decode length/precision
	length, err := e.readUint32(reader)
	if err != nil {
		return nil, err
	}

	precision, err := e.readUint32(reader)
	if err != nil {
		return nil, err
	}

	scale, err := e.readUint32(reader)
	if err != nil {
		return nil, err
	}

	// 解码约束信息 Decode constraint information
	nullable, err := e.decodeBool(reader)
	if err != nil {
		return nil, err
	}

	autoIncrement, err := e.decodeBool(reader)
	if err != nil {
		return nil, err
	}

	unique, err := e.decodeBool(reader)
	if err != nil {
		return nil, err
	}

	// 解码默认值 Decode default value
	var defaultValue interface{}
	hasDefault, err := e.readByte(reader)
	if err != nil {
		return nil, err
	}

	if hasDefault == ValueNotNull {
		if err := e.decodeValueInternal(reader, &defaultValue); err != nil {
			return nil, err
		}
	}

	// 解码注释 Decode comment
	comment, err := e.readString(reader)
	if err != nil {
		return nil, err
	}

	return &types.Column{
		Name:          name,
		Type:          types.DataType(typeID),
		Length:        length,
		Precision:     precision,
		Scale:         scale,
		Nullable:      nullable,
		AutoIncrement: autoIncrement,
		Unique:        unique,
		DefaultValue:  defaultValue,
		Comment:       comment,
	}, nil
}

// 特殊类型编码方法 Special type encoding methods

// encodeJSON 编码JSON Encode JSON
func (e *BinaryEncoder) encodeJSON(value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return e.writeBytes(data)
}

// decodeJSON 解码JSON Decode JSON
func (e *BinaryEncoder) decodeJSON(reader *bytes.Reader) (interface{}, error) {
	data, err := e.readBytes(reader)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

// encodeDecimal 编码十进制数 Encode decimal
func (e *BinaryEncoder) encodeDecimal(value interface{}) error {
	// 这里需要根据实际的Decimal类型实现
	// This needs to be implemented based on the actual Decimal type
	str := fmt.Sprintf("%v", value)
	return e.writeString(str)
}

// decodeDecimal 解码十进制数 Decode decimal
func (e *BinaryEncoder) decodeDecimal(reader *bytes.Reader) (interface{}, error) {
	// 这里需要根据实际的Decimal类型实现
	// This needs to be implemented based on the actual Decimal type
	str, err := e.readString(reader)
	if err != nil {
		return nil, err
	}

	// 返回字符串表示，实际使用时需要转换为具体的Decimal类型
	// Return string representation, need to convert to specific Decimal type in actual use
	return str, nil
}

// encodeUUID 编码UUID Encode UUID
func (e *BinaryEncoder) encodeUUID(value interface{}) error {
	str := fmt.Sprintf("%v", value)
	return e.writeString(str)
}

// decodeUUID 解码UUID Decode UUID
func (e *BinaryEncoder) decodeUUID(reader *bytes.Reader) (interface{}, error) {
	return e.readString(reader)
}

// encodeGeoPoint 编码地理点 Encode geo point
func (e *BinaryEncoder) encodeGeoPoint(value interface{}) error {
	// 假设GeoPoint结构包含Lat和Lng字段
	// Assume GeoPoint structure contains Lat and Lng fields
	point := value.(GeoPoint)

	if err := e.writeFloat64(point.Lat); err != nil {
		return err
	}

	return e.writeFloat64(point.Lng)
}

// decodeGeoPoint 解码地理点 Decode geo point
func (e *BinaryEncoder) decodeGeoPoint(reader *bytes.Reader) (interface{}, error) {
	lat, err := e.readFloat64(reader)
	if err != nil {
		return nil, err
	}

	lng, err := e.readFloat64(reader)
	if err != nil {
		return nil, err
	}

	return GeoPoint{Lat: lat, Lng: lng}, nil
}

// encodeVector 编码向量 Encode vector
func (e *BinaryEncoder) encodeVector(value interface{}) error {
	vector := value.([]float64)

	// 写入向量维度 Write vector dimension
	if err := e.writeVarInt(uint64(len(vector))); err != nil {
		return err
	}

	// 写入每个分量 Write each component
	for _, component := range vector {
		if err := e.writeFloat64(component); err != nil {
			return err
		}
	}

	return nil
}

// decodeVector 解码向量 Decode vector
func (e *BinaryEncoder) decodeVector(reader *bytes.Reader) (interface{}, error) {
	// 读取向量维度 Read vector dimension
	dimension, err := e.readVarInt(reader)
	if err != nil {
		return nil, err
	}

	if dimension > uint64(MaxBufferSize/8) { // 假设每个分量8字节 Assume 8 bytes per component
		return nil, fmt.Errorf("vector dimension too large: %d", dimension)
	}

	// 读取每个分量 Read each component
	vector := make([]float64, dimension)
	for i := uint64(0); i < dimension; i++ {
		component, err := e.readFloat64(reader)
		if err != nil {
			return nil, err
		}
		vector[i] = component
	}

	return vector, nil
}

// 辅助方法 Helper methods

// finalizeEncoding 完成编码 Finalize encoding
func (e *BinaryEncoder) finalizeEncoding() ([]byte, error) {
	data := e.buffer.Bytes()

	// 压缩数据 Compress data
	if e.compressor != nil {
		compressed, err := e.compressor.Compress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to compress data: %w", err)
		}
		data = compressed
	}

	// 计算校验和 Calculate checksum
	if e.checksummer != nil {
		checksum := e.checksummer.Checksum(data)
		data = append(data, checksum...)
	}

	return data, nil
}

// preprocessData 预处理数据 Preprocess data
func (e *BinaryEncoder) preprocessData(data []byte) ([]byte, error) {
	// 验证校验和 Verify checksum
	if e.checksummer != nil {
		checksumSize := e.checksummer.Size()
		if len(data) < checksumSize {
			return nil, errors.NewError(errors.ErrCodeDataCorruption, "data too short for checksum")
		}

		payload := data[:len(data)-checksumSize]
		expectedChecksum := data[len(data)-checksumSize:]
		actualChecksum := e.checksummer.Checksum(payload)

		if !bytes.Equal(expectedChecksum, actualChecksum) {
			return nil, errors.NewError(errors.ErrCodeDataCorruption, "checksum mismatch")
		}

		data = payload
	}

	// 解压数据 Decompress data
	if e.compressor != nil {
		decompressed, err := e.compressor.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress data: %w", err)
		}
		data = decompressed
	}

	return data, nil
}

// setTargetValue 设置目标值 Set target value
func (e *BinaryEncoder) setTargetValue(target interface{}, value interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetValue = targetValue.Elem()
	if !targetValue.CanSet() {
		return fmt.Errorf("target cannot be set")
	}

	if value == nil {
		targetValue.Set(reflect.Zero(targetValue.Type()))
		return nil
	}

	sourceValue := reflect.ValueOf(value)
	if sourceValue.Type().AssignableTo(targetValue.Type()) {
		targetValue.Set(sourceValue)
		return nil
	}

	// 尝试类型转换 Try type conversion
	if sourceValue.Type().ConvertibleTo(targetValue.Type()) {
		targetValue.Set(sourceValue.Convert(targetValue.Type()))
		return nil
	}

	return fmt.Errorf("cannot assign %T to %T", value, target)
}

// 缓存相关方法 Cache related methods

// getCachedEncoding 获取缓存的编码 Get cached encoding
func (e *BinaryEncoder) getCachedEncoding(value interface{}) ([]byte, bool) {
	if !e.config.EnableCaching {
		return nil, false
	}

	key := e.generateCacheKey(value)
	data, exists := e.cache[key]
	return data, exists
}

// setCachedEncoding 设置缓存的编码 Set cached encoding
func (e *BinaryEncoder) setCachedEncoding(value interface{}, data []byte) {
	if !e.config.EnableCaching {
		return
	}
	// 检查缓存大小限制 Check cache size limit
	if e.cacheSize+len(data) > e.config.MaxCacheSize {
		// 清理缓存 Clear cache
		e.clearCache()
	}

	key := e.generateCacheKey(value)
	e.cache[key] = make([]byte, len(data))
	copy(e.cache[key], data)
	e.cacheSize += len(data)
}

// generateCacheKey 生成缓存键 Generate cache key
func (e *BinaryEncoder) generateCacheKey(value interface{}) string {
	hasher := sha256.New()

	// 使用类型信息和值的字符串表示生成键
	// Use type information and string representation of value to generate key
	typeInfo := reflect.TypeOf(value).String()
	hasher.Write([]byte(typeInfo))

	valueStr := fmt.Sprintf("%v", value)
	hasher.Write([]byte(valueStr))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// clearCache 清理缓存 Clear cache
func (e *BinaryEncoder) clearCache() {
	e.cache = make(map[string][]byte)
	e.cacheSize = 0
}

// 类型注册相关方法 Type registration related methods

// TypeEncoder 类型编码器接口 Type encoder interface
type TypeEncoder interface {
	Encode(encoder *BinaryEncoder, value interface{}) error                        // 编码 Encode
	Decode(encoder *BinaryEncoder, reader *bytes.Reader, target interface{}) error // 解码 Decode
}

// RegisterTypeEncoder 注册类型编码器 Register type encoder
func (e *BinaryEncoder) RegisterTypeEncoder(typeName string, encoder TypeEncoder) {
	e.typeRegistry[typeName] = encoder
}

// UnregisterTypeEncoder 注销类型编码器 Unregister type encoder
func (e *BinaryEncoder) UnregisterTypeEncoder(typeName string) {
	delete(e.typeRegistry, typeName)
}

// registerBasicTypes 注册基础类型编码器 Register basic type encoders
func (e *BinaryEncoder) registerBasicTypes() {
	// 基础类型已经在主编码方法中处理，这里可以注册自定义类型
	// Basic types are already handled in main encoding methods, custom types can be registered here

	// 示例：注册时间类型的特殊编码器
	// Example: Register special encoder for time type
	e.RegisterTypeEncoder("time.Time", &TimeEncoder{})

	// 示例：注册UUID类型编码器
	// Example: Register UUID type encoder
	e.RegisterTypeEncoder("UUID", &UUIDEncoder{})
}

// encodeGeneric 通用编码 Generic encoding
func (e *BinaryEncoder) encodeGeneric(value interface{}) error {
	// 使用反射进行通用编码 Use reflection for generic encoding
	rv := reflect.ValueOf(value)
	rt := reflect.TypeOf(value)

	switch rv.Kind() {
	case reflect.Struct:
		return e.encodeStruct(rv, rt)
	case reflect.Slice:
		return e.encodeSlice(rv, rt)
	case reflect.Map:
		return e.encodeReflectMap(rv, rt)
	case reflect.Ptr:
		if rv.IsNil() {
			return e.writeByte(TypeNull)
		}
		return e.encodeGeneric(rv.Elem().Interface())
	default:
		// 使用Gob编码作为最后的后备方案 Use Gob encoding as last fallback
		return e.encodeWithGob(value)
	}
}

// encodeStruct 编码结构体 Encode struct
func (e *BinaryEncoder) encodeStruct(rv reflect.Value, rt reflect.Type) error {
	// 写入字段数量 Write field count
	numFields := rt.NumField()
	if err := e.writeVarInt(uint64(numFields)); err != nil {
		return err
	}

	// 编码每个字段 Encode each field
	for i := 0; i < numFields; i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// 跳过私有字段 Skip private fields
		if !fieldValue.CanInterface() {
			continue
		}

		// 写入字段名 Write field name
		if err := e.writeString(field.Name); err != nil {
			return err
		}

		// 编码字段值 Encode field value
		if err := e.encodeValueInternal(fieldValue.Interface()); err != nil {
			return err
		}
	}

	return nil
}

// encodeSlice 编码切片 Encode slice
func (e *BinaryEncoder) encodeSlice(rv reflect.Value, rt reflect.Type) error {
	// 写入切片长度 Write slice length
	length := rv.Len()
	if err := e.writeVarInt(uint64(length)); err != nil {
		return err
	}

	// 编码每个元素 Encode each element
	for i := 0; i < length; i++ {
		element := rv.Index(i)
		if err := e.encodeValueInternal(element.Interface()); err != nil {
			return err
		}
	}

	return nil
}

// encodeReflectMap 编码映射（使用反射） Encode map (using reflection)
func (e *BinaryEncoder) encodeReflectMap(rv reflect.Value, rt reflect.Type) error {
	// 写入映射长度 Write map length
	length := rv.Len()
	if err := e.writeVarInt(uint64(length)); err != nil {
		return err
	}

	// 获取所有键并排序 Get all keys and sort them
	keys := rv.MapKeys()
	sortedKeys := make([]reflect.Value, len(keys))
	copy(sortedKeys, keys)

	// 简单的字符串排序（仅适用于字符串键）
	// Simple string sorting (only for string keys)
	if rt.Key().Kind() == reflect.String {
		sort.Slice(sortedKeys, func(i, j int) bool {
			return sortedKeys[i].String() < sortedKeys[j].String()
		})
	}

	// 编码每个键值对 Encode each key-value pair
	for _, key := range sortedKeys {
		value := rv.MapIndex(key)

		// 编码键 Encode key
		if err := e.encodeValueInternal(key.Interface()); err != nil {
			return err
		}

		// 编码值 Encode value
		if err := e.encodeValueInternal(value.Interface()); err != nil {
			return err
		}
	}

	return nil
}

// encodeWithGob 使用Gob编码 Encode with Gob
func (e *BinaryEncoder) encodeWithGob(value interface{}) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("failed to encode with Gob: %w", err)
	}

	return e.writeBytes(buf.Bytes())
}

// 压缩器接口和实现 Compressor interface and implementations

// Compressor 压缩器接口 Compressor interface
type Compressor interface {
	Compress(data []byte) ([]byte, error)   // 压缩 Compress
	Decompress(data []byte) ([]byte, error) // 解压 Decompress
	Type() CompressionType                  // 类型 Type
}

// NewCompressor 创建压缩器 Create compressor
func NewCompressor(compressionType CompressionType, level int) Compressor {
	switch compressionType {
	case None:
		return &NoCompressor{}
	case Gzip:
		return &GzipCompressor{Level: level}
	case Zlib:
		return &ZlibCompressor{Level: level}
	case Snappy:
		return &SnappyCompressor{}
	case LZ4:
		return &LZ4Compressor{}
	case Zstd:
		return &ZstdCompressor{Level: level}
	default:
		return &NoCompressor{}
	}
}

// NoCompressor 无压缩器 No compressor
type NoCompressor struct{}

func (c *NoCompressor) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (c *NoCompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}

func (c *NoCompressor) Type() CompressionType {
	return None
}

// GzipCompressor Gzip压缩器 Gzip compressor
type GzipCompressor struct {
	Level int // 压缩级别 Compression level
}

func (c *GzipCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, c.Level)
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *GzipCompressor) Decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (c *GzipCompressor) Type() CompressionType {
	return Gzip
}

// ZlibCompressor Zlib压缩器 Zlib compressor
type ZlibCompressor struct {
	Level int // 压缩级别 Compression level
}

func (c *ZlibCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, c.Level)
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *ZlibCompressor) Decompress(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (c *ZlibCompressor) Type() CompressionType {
	return Zlib
}

// SnappyCompressor Snappy压缩器 Snappy compressor
type SnappyCompressor struct{}

func (c *SnappyCompressor) Compress(data []byte) ([]byte, error) {
	return snappy.Encode(nil, data), nil
}

func (c *SnappyCompressor) Decompress(data []byte) ([]byte, error) {
	return snappy.Decode(nil, data)
}

func (c *SnappyCompressor) Type() CompressionType {
	return Snappy
}

// LZ4Compressor LZ4压缩器 LZ4 compressor
type LZ4Compressor struct{}

func (c *LZ4Compressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := lz4.NewWriter(&buf)
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *LZ4Compressor) Decompress(data []byte) ([]byte, error) {
	reader := lz4.NewReader(bytes.NewReader(data))
	return io.ReadAll(reader)
}

func (c *LZ4Compressor) Type() CompressionType {
	return LZ4
}

// ZstdCompressor Zstd压缩器 Zstd compressor
type ZstdCompressor struct {
	Level int // 压缩级别 Compression level
}

func (c *ZstdCompressor) Compress(data []byte) ([]byte, error) {
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevel(c.Level)))
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	return encoder.EncodeAll(data, nil), nil
}

func (c *ZstdCompressor) Decompress(data []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	return decoder.DecodeAll(data, nil)
}

func (c *ZstdCompressor) Type() CompressionType {
	return Zstd
}

// 校验和器接口和实现 Checksummer interface and implementations

// Checksummer 校验和器接口 Checksummer interface
type Checksummer interface {
	Checksum(data []byte) []byte       // 计算校验和 Calculate checksum
	Verify(data, checksum []byte) bool // 验证校验和 Verify checksum
	Size() int                         // 校验和大小 Checksum size
	Type() ChecksumType                // 类型 Type
}

// NewChecksummer 创建校验和器 Create checksummer
func NewChecksummer(checksumType ChecksumType) Checksummer {
	switch checksumType {
	case NoChecksum:
		return &NoChecksummer{}
	case CRC32:
		return &CRC32Checksummer{}
	case MD5:
		return &MD5Checksummer{}
	case SHA256:
		return &SHA256Checksummer{}
	default:
		return &NoChecksummer{}
	}
}

// NoChecksummer 无校验和器 No checksummer
type NoChecksummer struct{}

func (c *NoChecksummer) Checksum(data []byte) []byte {
	return nil
}

func (c *NoChecksummer) Verify(data, checksum []byte) bool {
	return true
}

func (c *NoChecksummer) Size() int {
	return 0
}

func (c *NoChecksummer) Type() ChecksumType {
	return NoChecksum
}

// CRC32Checksummer CRC32校验和器 CRC32 checksummer
type CRC32Checksummer struct{}

func (c *CRC32Checksummer) Checksum(data []byte) []byte {
	checksum := crc32.ChecksumIEEE(data)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, checksum)
	return buf
}

func (c *CRC32Checksummer) Verify(data, checksum []byte) bool {
	expected := c.Checksum(data)
	return bytes.Equal(expected, checksum)
}

func (c *CRC32Checksummer) Size() int {
	return 4
}

func (c *CRC32Checksummer) Type() ChecksumType {
	return CRC32
}

// MD5Checksummer MD5校验和器 MD5 checksummer
type MD5Checksummer struct{}

func (c *MD5Checksummer) Checksum(data []byte) []byte {
	hasher := md5.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func (c *MD5Checksummer) Verify(data, checksum []byte) bool {
	expected := c.Checksum(data)
	return bytes.Equal(expected, checksum)
}

func (c *MD5Checksummer) Size() int {
	return 16
}

func (c *MD5Checksummer) Type() ChecksumType {
	return MD5
}

// SHA256Checksummer SHA256校验和器 SHA256 checksummer
type SHA256Checksummer struct{}

func (c *SHA256Checksummer) Checksum(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func (c *SHA256Checksummer) Verify(data, checksum []byte) bool {
	expected := c.Checksum(data)
	return bytes.Equal(expected, checksum)
}

func (c *SHA256Checksummer) Size() int {
	return 32
}

func (c *SHA256Checksummer) Type() ChecksumType {
	return SHA256
}

// 数据结构定义 Data structure definitions

// Row 行数据结构 Row data structure
type Row struct {
	ID      string        `json:"id"`      // 行ID Row ID
	Version uint64        `json:"version"` // 版本号 Version
	Schema  *types.Schema `json:"schema"`  // 表结构 Schema
	Values  []interface{} `json:"values"`  // 列值 Column values
}

// IndexKey 索引键结构 Index key structure
type IndexKey struct {
	TableName string           `json:"table_name"` // 表名 Table name
	IndexName string           `json:"index_name"` // 索引名 Index name
	IndexType IndexType        `json:"index_type"` // 索引类型 Index type
	Types     []types.DataType `json:"types"`      // 键值类型 Key value types
	Values    []interface{}    `json:"values"`     // 键值 Key values
	RowID     string           `json:"row_id"`     // 行ID Row ID
}

// IndexType 索引类型 Index type
type IndexType int

const (
	IndexTypePrimaryKey   IndexType = iota // 主键索引 Primary key index
	IndexTypeSecondaryKey                  // 二级索引 Secondary index
	IndexTypeUniqueKey                     // 唯一索引 Unique index
	IndexTypeCompositeKey                  // 复合索引 Composite index
	IndexTypeFullTextKey                   // 全文索引 Full-text index
	IndexTypeVectorKey                     // 向量索引 Vector index
	IndexTypeGeoKey                        // 地理索引 Geo index
)

// DatabaseMetadata 数据库元数据 Database metadata
type DatabaseMetadata struct {
	Name      string    `json:"name"`       // 数据库名 Database name
	Tables    []string  `json:"tables"`     // 表名列表 Table name list
	CreatedAt time.Time `json:"created_at"` // 创建时间 Creation time
	UpdatedAt time.Time `json:"updated_at"` // 更新时间 Update time
}

// TableMetadata 表元数据 Table metadata
type TableMetadata struct {
	Name      string           `json:"name"`       // 表名 Table name
	Database  string           `json:"database"`   // 数据库名 Database name
	Schema    *types.Schema    `json:"schema"`     // 表结构 Schema
	Indexes   []*IndexMetadata `json:"indexes"`    // 索引列表 Index list
	RowCount  uint64           `json:"row_count"`  // 行数 Row count
	Size      uint64           `json:"size"`       // 大小 Size
	CreatedAt time.Time        `json:"created_at"` // 创建时间 Creation time
	UpdatedAt time.Time        `json:"updated_at"` // 更新时间 Update time
}

// IndexMetadata 索引元数据 Index metadata
type IndexMetadata struct {
	Name      string    `json:"name"`       // 索引名 Index name
	Type      IndexType `json:"type"`       // 索引类型 Index type
	Columns   []string  `json:"columns"`    // 列名 Column names
	Unique    bool      `json:"unique"`     // 是否唯一 Is unique
	CreatedAt time.Time `json:"created_at"` // 创建时间 Creation time
}

// SchemaMetadata 表结构元数据 Schema metadata
type SchemaMetadata struct {
	Schema *types.Schema `json:"schema"` // 表结构 Schema
}

// RowHeader 行头部信息 Row header information
type RowHeader struct {
	ID          string        `json:"id"`           // 行ID Row ID
	Version     uint64        `json:"version"`      // 版本号 Version
	ColumnCount uint64        `json:"column_count"` // 列数量 Column count
	Schema      *types.Schema `json:"schema"`       // 表结构 Schema
}

// IndexKeyHeader 索引键头部信息 Index key header information
type IndexKeyHeader struct {
	TableName string           `json:"table_name"` // 表名 Table name
	IndexName string           `json:"index_name"` // 索引名 Index name
	IndexType IndexType        `json:"index_type"` // 索引类型 Index type
	KeyCount  uint64           `json:"key_count"`  // 键数量 Key count
	Types     []types.DataType `json:"types"`      // 类型列表 Type list
}

// MetadataHeader 元数据头部信息 Metadata header information
type MetadataHeader struct {
	Type string `json:"type"` // 元数据类型 Metadata type
}

// ColumnInfo 列信息 Column information
type ColumnInfo struct {
	Name     string         `json:"name"`     // 列名 Column name
	Type     types.DataType `json:"type"`     // 列类型 Column type
	Nullable bool           `json:"nullable"` // 是否可空 Is nullable
}

// GeoPoint 地理点 Geo point
type GeoPoint struct {
	Lat float64 `json:"lat"` // 纬度 Latitude
	Lng float64 `json:"lng"` // 经度 Longitude
}

// 特殊类型编码器实现 Special type encoder implementations

// TimeEncoder 时间编码器 Time encoder
type TimeEncoder struct{}

func (e *TimeEncoder) Encode(encoder *BinaryEncoder, value interface{}) error {
	t := value.(time.Time)
	return encoder.writeTime(t)
}

func (e *TimeEncoder) Decode(encoder *BinaryEncoder, reader *bytes.Reader, target interface{}) error {
	t, err := encoder.readTime(reader)
	if err != nil {
		return err
	}
	return encoder.setTargetValue(target, t)
}

// UUIDEncoder UUID编码器 UUID encoder
type UUIDEncoder struct{}

func (e *UUIDEncoder) Encode(encoder *BinaryEncoder, value interface{}) error {
	return encoder.encodeUUID(value)
}

func (e *UUIDEncoder) Decode(encoder *BinaryEncoder, reader *bytes.Reader, target interface{}) error {
	uuid, err := encoder.decodeUUID(reader)
	if err != nil {
		return err
	}
	return encoder.setTargetValue(target, uuid)
}

// 工具函数 Utility functions

// GetNativeByteOrder 获取本机字节序 Get native byte order
func GetNativeByteOrder() ByteOrder {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	if b == 0x04 {
		return LittleEndian
	}
	return BigEndian
}

// EstimateEncodingSize 估算编码大小 Estimate encoding size
func EstimateEncodingSize(value interface{}) int {
	switch v := value.(type) {
	case nil:
		return 1
	case bool:
		return 2
	case int8, uint8:
		return 2
	case int16, uint16:
		return 3
	case int32, uint32, float32:
		return 5
	case int64, uint64, float64:
		return 9
	case int, uint:
		return 9
	case string:
		return 5 + len(v) // varint + string data
	case []byte:
		return 5 + len(v) // varint + byte data
	case time.Time:
		return 13 // int64 + int32
	case []interface{}:
		size := 5 // varint for length
		for _, item := range v {
			size += EstimateEncodingSize(item)
		}
		return size
	case map[string]interface{}:
		size := 5 // varint for length
		for key, val := range v {
			size += 5 + len(key)              // key
			size += EstimateEncodingSize(val) // value
		}
		return size
	default:
		// 对于未知类型，返回一个保守的估计值
		// For unknown types, return a conservative estimate
		return 64
	}
}

// ValidateEncodingConfig 验证编码配置 Validate encoding configuration
func ValidateEncodingConfig(config *EncodingConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrCodeInvalidParameter, "encoding config is nil")
	}

	if config.Version < 1 || config.Version > CurrentVersion {
		return errors.NewError(errors.ErrCodeInvalidParameter,
			fmt.Sprintf("unsupported encoding version: %d", config.Version))
	}

	if config.BufferSize < MinBufferSize || config.BufferSize > MaxBufferSize {
		return errors.NewError(errors.ErrCodeInvalidParameter,
			fmt.Sprintf("invalid buffer size: %d", config.BufferSize))
	}

	if config.EnableCaching && config.MaxCacheSize <= 0 {
		return errors.NewError(errors.ErrCodeInvalidParameter,
			"max cache size must be positive when caching is enabled")
	}

	return nil
}

// CreateOptimizedConfig 创建优化配置 Create optimized configuration
func CreateOptimizedConfig(dataType string) *EncodingConfig {
	config := &EncodingConfig{
		Version:          CurrentVersion,
		ByteOrder:        GetNativeByteOrder(),
		BufferSize:       DefaultBufferSize,
		EnableCaching:    true,
		MaxCacheSize:     1024 * 1024, // 1MB
		CompressionLevel: 6,
	}

	// 根据数据类型优化配置 Optimize configuration based on data type
	switch strings.ToLower(dataType) {
	case "text", "json", "string":
		config.Compression = Zstd
		config.Checksum = CRC32
		config.CompressionLevel = 3

	case "binary", "blob":
		config.Compression = LZ4
		config.Checksum = SHA256
		config.CompressionLevel = 1

	case "numeric", "integer":
		config.Compression = Snappy
		config.Checksum = CRC32
		config.CompressionLevel = 6

	case "time", "timestamp":
		config.Compression = Snappy
		config.Checksum = CRC32
		config.CompressionLevel = 6

	default:
		config.Compression = Snappy
		config.Checksum = CRC32
		config.CompressionLevel = 6
	}

	return config
}
