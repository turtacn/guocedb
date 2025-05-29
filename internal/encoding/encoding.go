// Package encoding defines internal tools for common data encoding and decoding within the Guocedb project.
// This file provides efficient conversion methods between various data structures (Go native types,
// custom structs) and byte slices or specific formats (e.g., Protobuf, gob, etc.). It is a core
// component relied upon by the storage layer (especially badger/encoding.go), compute/catalog/persistent/persistent_catalog.go,
// and potentially protocol/mysql for data transmission.
package encoding

import (
	"bytes"           // For byte buffer operations.
	"encoding/binary" // For binary encoding of primitive types.
	"fmt"             // For error formatting.
	"io"              // For io.ReadFull or io.Writer.

	"github.com/turtacn/guocedb/common/errors"      // For unified error handling.
	"github.com/turtacn/guocedb/common/types/enum"  // For component types and SQL types.
	"github.com/turtacn/guocedb/common/types/value" // For Guocedb's internal Value types.

	"google.golang.org/protobuf/proto" // For Protobuf serialization (if used for complex structs).
	"gopkg.in/yaml.v3"                 // For YAML encoding/decoding of configuration-like data.
)

// Encoder provides an interface for encoding various data types into a byte slice.
// This interface can be extended to support different encoding mechanisms (e.g., gob, protobuf).
type Encoder interface {
	Encode(v interface{}) ([]byte, error)
}

// Decoder provides an interface for decoding byte slices into various data types.
type Decoder interface {
	Decode(data []byte, v interface{}) error
}

// BinaryEncoder implements the Encoder interface for binary encoding of primitive types.
type BinaryEncoder struct{}

// NewBinaryEncoder creates a new BinaryEncoder.
func NewBinaryEncoder() *BinaryEncoder {
	return &BinaryEncoder{}
}

// Encode encodes a primitive Go type (e.g., int64, float64, bool) or Guocedb's Value types
// into a byte slice using binary.BigEndian.
// For complex types, consider using ProtobufEncoder or GobEncoder.
func (e *BinaryEncoder) Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	// Handle Guocedb's internal Value types directly
	if val, ok := v.(value.Value); ok {
		return val.Bytes() // Use the Value's own Bytes() method for serialization
	}

	// Handle primitive Go types
	switch val := v.(type) {
	case int64:
		err = binary.Write(&buf, binary.BigEndian, val)
	case uint64:
		err = binary.Write(&buf, binary.BigEndian, val)
	case int32:
		err = binary.Write(&buf, binary.BigEndian, val)
	case uint32:
		err = binary.Write(&buf, binary.BigEndian, val)
	case int:
		err = binary.Write(&buf, binary.BigEndian, int64(val)) // Encode int as int64
	case uint:
		err = binary.Write(&buf, binary.BigEndian, uint64(val)) // Encode uint as uint64
	case float64:
		err = binary.Write(&buf, binary.BigEndian, val)
	case bool:
		if val {
			err = buf.WriteByte(0x01)
		} else {
			err = buf.WriteByte(0x00)
		}
	case string:
		// Length prefix the string for robust deserialization
		lenBytes := make([]byte, 4) // Max string length up to 2^32 - 1 bytes
		binary.BigEndian.PutUint32(lenBytes, uint32(len(val)))
		buf.Write(lenBytes)
		_, err = buf.WriteString(val)
	case []byte:
		// Directly write byte slice, potentially length-prefixed if context requires
		// For now, assume it's raw bytes that don't need length prefixing here
		_, err = buf.Write(val)
	default:
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("unsupported type for binary encoding: %T", v), nil)
	}

	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to binary encode value of type %T", v), err)
	}
	return buf.Bytes(), nil
}

// BinaryDecoder implements the Decoder interface for binary decoding of primitive types.
type BinaryDecoder struct{}

// NewBinaryDecoder creates a new BinaryDecoder.
func NewBinaryDecoder() *BinaryDecoder {
	return &BinaryDecoder{}
}

// Decode decodes a byte slice into a pointer to a primitive Go type or Guocedb's Value types.
// The target 'v' must be a pointer to the type you expect to decode.
func (d *BinaryDecoder) Decode(data []byte, v interface{}) error {
	buf := bytes.NewReader(data)
	var err error

	// Handle Guocedb's internal Value types (requires knowing the SQLType)
	// This would typically be handled by a specialized function like value.ValueFromBytes
	// For this generic decoder, it's harder unless type info is embedded in 'data'.
	// This part is illustrative; actual usage in storage would pass SQLType.
	// For now, if 'v' is a Value, we assume it's already a specific concrete type and we attempt to populate it.
	if val, ok := v.(value.Value); ok {
		// This path is tricky without knowing the target SQLType from 'data'.
		// A more robust system would involve a factory method or type info in data.
		// For simplicity, let's assume 'data' contains enough info or is used with value.ValueFromBytes directly.
		// If 'v' is a pointer to a concrete value.Value implementation, we can try to unmarshal.
		// E.g., if v is *value.Integer, we try to decode an integer.
		// This requires more complex logic for each value type, or embedding type information.
		// For now, this generic BinaryDecoder won't directly populate a 'value.Value' interface type,
		// but rather the underlying Go types that 'value.Value' wraps.
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			"decoding into common/types/value.Value interface directly is not supported by generic BinaryDecoder; use value.ValueFromBytes", nil)
	}

	switch val := v.(type) {
	case *int64:
		err = binary.Read(buf, binary.BigEndian, val)
	case *uint64:
		err = binary.Read(buf, binary.BigEndian, val)
	case *int32:
		err = binary.Read(buf, binary.BigEndian, val)
	case *uint32:
		err = binary.Read(buf, binary.BigEndian, val)
	case *int:
		var i64 int64
		err = binary.Read(buf, binary.BigEndian, &i64)
		if err == nil {
			*val = int(i64)
		}
	case *uint:
		var u64 uint64
		err = binary.Read(buf, binary.BigEndian, &u64)
		if err == nil {
			*val = uint(u64)
		}
	case *float64:
		err = binary.Read(buf, binary.BigEndian, val)
	case *bool:
		var b byte
		err = binary.Read(buf, binary.BigEndian, &b)
		if err == nil {
			*val = (b == 0x01)
		}
	case *string:
		// Read length prefix then the string bytes
		var strLen uint32
		if err = binary.Read(buf, binary.BigEndian, &strLen); err != nil {
			break // Exit switch on error
		}
		strBytes := make([]byte, strLen)
		if _, err = io.ReadFull(buf, strBytes); err != nil {
			break
		}
		*val = string(strBytes)
	case *[]byte:
		// Read remaining bytes into the slice
		*val = make([]byte, buf.Len())
		_, err = io.ReadFull(buf, *val)
	default:
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("unsupported type for binary decoding: %T", v), nil)
	}

	if err != nil {
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to binary decode value of type %T", v), err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Protobuf Encoding (Example)
// -----------------------------------------------------------------------------

// ProtobufEncoder implements the Encoder interface for Protobuf serialization.
// Requires defining `.proto` files and generating Go code.
type ProtobufEncoder struct{}

// NewProtobufEncoder creates a new ProtobufEncoder.
func NewProtobufEncoder() *ProtobufEncoder {
	return &ProtobufEncoder{}
}

// Encode encodes a proto.Message into a byte slice.
func (e *ProtobufEncoder) Encode(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("ProtobufEncoder requires proto.Message, got %T", v), nil)
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to marshal protobuf message of type %T", msg), err)
	}
	return data, nil
}

// ProtobufDecoder implements the Decoder interface for Protobuf deserialization.
type ProtobufDecoder struct{}

// NewProtobufDecoder creates a new ProtobufDecoder.
func NewProtobufDecoder() *ProtobufDecoder {
	return &ProtobufDecoder{}
}

// Decode decodes a byte slice into a proto.Message.
// The target 'v' must be a pointer to a proto.Message type (e.g., &MyProtoMessage{}).
func (d *ProtobufDecoder) Decode(data []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("ProtobufDecoder requires proto.Message pointer, got %T", v), nil)
	}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to unmarshal protobuf message of type %T", msg), err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// YAML Encoding (Example for configuration or metadata)
// -----------------------------------------------------------------------------

// YAMLEncoder implements the Encoder interface for YAML serialization.
// Useful for human-readable configuration or metadata.
type YAMLEncoder struct{}

// NewYAMLEncoder creates a new YAMLEncoder.
func NewYAMLEncoder() *YAMLEncoder {
	return &YAMLEncoder{}
}

// Encode encodes a Go struct into a YAML byte slice.
func (e *YAMLEncoder) Encode(v interface{}) ([]byte, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, errors.NewGuocedbError(enum.ErrEncoding, errors.CodeSerializationFailed,
			fmt.Sprintf("failed to marshal YAML for type %T", v), err)
	}
	return data, nil
}

// YAMLDecoder implements the Decoder interface for YAML deserialization.
type YAMLDecoder struct{}

// NewYAMLDecoder creates a new YAMLDecoder.
func NewYAMLDecoder() *YAMLDecoder {
	return &YAMLDecoder{}
}

// Decode decodes a YAML byte slice into a Go struct.
// The target 'v' must be a pointer to the struct type.
func (d *YAMLDecoder) Decode(data []byte, v interface{}) error {
	err := yaml.Unmarshal(data, v)
	if err != nil {
		return errors.NewGuocedbError(enum.ErrEncoding, errors.CodeDeserializationFailed,
			fmt.Sprintf("failed to unmarshal YAML into type %T", v), err)
	}
	return nil
}

//Personal.AI order the ending
