// Package encoding provides utilities for encoding and decoding data within Guocedb.
// This includes serialization formats for storing data persistently,
// as well as methods for transforming data for network transmission or inter-component communication.
//
// 此包提供了 Guocedb 内部数据编码和解码的工具。
// 这包括用于持久存储数据的序列化格式，
// 以及用于网络传输或组件间通信的数据转换方法。
package encoding

import (
	"bytes"
	"encoding/binary" // For encoding/decoding fixed-size integers
	"fmt"
	"time" // For time.Time serialization
)

// The following constants define byte prefixes used for encoding different data types.
// These prefixes help in distinguishing the type of data when deserializing,
// enabling a self-describing binary format.
//
// 以下常量定义了用于编码不同数据类型的字节前缀。
// 这些前缀有助于在反序列化时区分数据类型，
// 从而实现自描述的二进制格式。
const (
	// TypePrefixNull indicates a NULL value.
	TypePrefixNull byte = 0x00
	// TypePrefixBooleanFalse indicates a boolean false value.
	TypePrefixBooleanFalse byte = 0x01
	// TypePrefixBooleanTrue indicates a boolean true value.
	TypePrefixBooleanTrue byte = 0x02
	// TypePrefixTinyInt indicates an 8-bit signed integer.
	TypePrefixTinyInt byte = 0x03
	// TypePrefixSmallInt indicates a 16-bit signed integer.
	TypePrefixSmallInt byte = 0x04
	// TypePrefixInt indicates a 32-bit signed integer.
	TypePrefixInt byte = 0x05
	// TypePrefixBigInt indicates a 64-bit signed integer.
	TypePrefixBigInt byte = 0x06
	// TypePrefixFloat indicates a 32-bit floating-point number.
	TypePrefixFloat byte = 0x07
	// TypePrefixDouble indicates a 64-bit floating-point number.
	TypePrefixDouble byte = 0x08
	// TypePrefixText indicates a string value.
	TypePrefixText byte = 0x09
	// TypePrefixBlob indicates binary large object data.
	TypePrefixBlob byte = 0x0A
	// TypePrefixDateTime indicates a time.Time value representing date and time.
	TypePrefixDateTime byte = 0x0B
	// TypePrefixDate indicates a time.Time value representing only the date.
	TypePrefixDate byte = 0x0C
	// TypePrefixTime indicates a time.Time value representing only the time of day.
	TypePrefixTime byte = 0x0D
	// TypePrefixReservedStart is a placeholder for future type prefixes.
	TypePrefixReservedStart byte = 0x10
	// TypePrefixReservedEnd marks the end of reserved type prefixes.
	TypePrefixReservedEnd byte = 0xFE
	// TypePrefixInvalid indicates an invalid or unknown type.
	TypePrefixInvalid byte = 0xFF
)

// EncodeNull encodes a NULL value. It returns a single byte prefix.
// 编码 NULL 值。它返回一个单字节前缀。
func EncodeNull() []byte {
	return []byte{TypePrefixNull}
}

// DecodeNull decodes a NULL value. It checks if the prefix matches.
// 解码 NULL 值。它检查前缀是否匹配。
func DecodeNull(data []byte) error {
	if len(data) != 1 || data[0] != TypePrefixNull {
		return fmt.Errorf("invalid data for NULL decoding")
	}
	return nil
}

// EncodeBoolean encodes a boolean value. It returns a single byte prefix.
// 编码布尔值。它返回一个单字节前缀。
func EncodeBoolean(v bool) []byte {
	if v {
		return []byte{TypePrefixBooleanTrue}
	}
	return []byte{TypePrefixBooleanFalse}
}

// DecodeBoolean decodes a boolean value from the given byte slice.
// It returns the boolean value and any error encountered.
// 从给定字节切片中解码布尔值。
// 它返回布尔值和遇到的任何错误。
func DecodeBoolean(data []byte) (bool, error) {
	if len(data) != 1 {
		return false, fmt.Errorf("invalid data length for boolean decoding: expected 1 byte, got %d", len(data))
	}
	switch data[0] {
	case TypePrefixBooleanTrue:
		return true, nil
	case TypePrefixBooleanFalse:
		return false, nil
	default:
		return false, fmt.Errorf("invalid prefix for boolean decoding: 0x%02x", data[0])
	}
}

// EncodeTinyInt encodes an int8 value into a byte slice, prefixed with TypePrefixTinyInt.
// 编码 int8 值到字节切片，前缀为 TypePrefixTinyInt。
func EncodeTinyInt(v int8) []byte {
	buf := make([]byte, 1+binary.Size(v))
	buf[0] = TypePrefixTinyInt
	binary.BigEndian.PutUint8(buf[1:], uint8(v)) // int8 to uint8 for PutUint8
	return buf
}

// DecodeTinyInt decodes an int8 value from a byte slice.
// 从字节切片中解码 int8 值。
func DecodeTinyInt(data []byte) (int8, error) {
	if len(data) != 1+binary.Size(int8(0)) || data[0] != TypePrefixTinyInt {
		return 0, fmt.Errorf("invalid data for TinyInt decoding")
	}
	return int8(binary.BigEndian.Uint8(data[1:])), nil
}

// EncodeSmallInt encodes an int16 value into a byte slice, prefixed with TypePrefixSmallInt.
// 编码 int16 值到字节切片，前缀为 TypePrefixSmallInt。
func EncodeSmallInt(v int16) []byte {
	buf := make([]byte, 1+binary.Size(v))
	buf[0] = TypePrefixSmallInt
	binary.BigEndian.PutUint16(buf[1:], uint16(v))
	return buf
}

// DecodeSmallInt decodes an int16 value from a byte slice.
// 从字节切片中解码 int16 值。
func DecodeSmallInt(data []byte) (int16, error) {
	if len(data) != 1+binary.Size(int16(0)) || data[0] != TypePrefixSmallInt {
		return 0, fmt.Errorf("invalid data for SmallInt decoding")
	}
	return int16(binary.BigEndian.Uint16(data[1:])), nil
}

// EncodeInt encodes an int32 value into a byte slice, prefixed with TypePrefixInt.
// 编码 int32 值到字节切片，前缀为 TypePrefixInt。
func EncodeInt(v int32) []byte {
	buf := make([]byte, 1+binary.Size(v))
	buf[0] = TypePrefixInt
	binary.BigEndian.PutUint32(buf[1:], uint32(v))
	return buf
}

// DecodeInt decodes an int32 value from a byte slice.
// 从字节切片中解码 int32 值。
func DecodeInt(data []byte) (int32, error) {
	if len(data) != 1+binary.Size(int32(0)) || data[0] != TypePrefixInt {
		return 0, fmt.Errorf("invalid data for Int decoding")
	}
	return int32(binary.BigEndian.Uint32(data[1:])), nil
}

// EncodeBigInt encodes an int64 value into a byte slice, prefixed with TypePrefixBigInt.
// 编码 int64 值到字节切片，前缀为 TypePrefixBigInt。
func EncodeBigInt(v int64) []byte {
	buf := make([]byte, 1+binary.Size(v))
	buf[0] = TypePrefixBigInt
	binary.BigEndian.PutUint64(buf[1:], uint64(v))
	return buf
}

// DecodeBigInt decodes an int64 value from a byte slice.
// 从字节切片中解码 int64 值。
func DecodeBigInt(data []byte) (int64, error) {
	if len(data) != 1+binary.Size(int64(0)) || data[0] != TypePrefixBigInt {
		return 0, fmt.Errorf("invalid data for BigInt decoding")
	}
	return int64(binary.BigEndian.Uint64(data[1:])), nil
}

// EncodeFloat encodes a float32 value into a byte slice, prefixed with TypePrefixFloat.
// 编码 float32 值到字节切片，前缀为 TypePrefixFloat。
func EncodeFloat(v float32) []byte {
	buf := make([]byte, 1+binary.Size(v))
	buf[0] = TypePrefixFloat
	binary.BigEndian.PutUint32(buf[1:], uint32(v)) // Assuming float32 fits into uint32 bits
	return buf
}

// DecodeFloat decodes a float32 value from a byte slice.
// 从字节切片中解码 float32 值。
func DecodeFloat(data []byte) (float32, error) {
	if len(data) != 1+binary.Size(float32(0)) || data[0] != TypePrefixFloat {
		return 0, fmt.Errorf("invalid data for Float decoding")
	}
	var f float32
	// binary.Read requires an io.Reader and a pointer to the value
	buf := bytes.NewReader(data[1:])
	err := binary.Read(buf, binary.BigEndian, &f)
	if err != nil {
		return 0, fmt.Errorf("failed to read float32: %w", err)
	}
	return f, nil
}

// EncodeDouble encodes a float64 value into a byte slice, prefixed with TypePrefixDouble.
// 编码 float64 值到字节切片，前缀为 TypePrefixDouble。
func EncodeDouble(v float64) []byte {
	buf := make([]byte, 1+binary.Size(v))
	buf[0] = TypePrefixDouble
	binary.BigEndian.PutUint64(buf[1:], uint64(v)) // Assuming float64 fits into uint64 bits
	return buf
}

// DecodeDouble decodes a float64 value from a byte slice.
// 从字节切片中解码 float64 值。
func DecodeDouble(data []byte) (float64, error) {
	if len(data) != 1+binary.Size(float64(0)) || data[0] != TypePrefixDouble {
		return 0, fmt.Errorf("invalid data for Double decoding")
	}
	var f float64
	buf := bytes.NewReader(data[1:])
	err := binary.Read(buf, binary.BigEndian, &f)
	if err != nil {
		return 0, fmt.Errorf("failed to read float64: %w", err)
	}
	return f, nil
}

// EncodeText encodes a string value into a byte slice. It prefixes the data with TypePrefixText
// followed by a length prefix (variable length encoded) and then the string bytes.
// 编码字符串值到字节切片。它以 TypePrefixText 为前缀，
// 紧接着是长度前缀（变长编码），然后是字符串字节。
func EncodeText(s string) []byte {
	// First byte is the type prefix
	// Then, length prefix (variable length encoding for length)
	// Finally, the string bytes
	var buf bytes.Buffer
	buf.WriteByte(TypePrefixText)
	// Write length prefix (e.g., Uvarint)
	// For simplicity, let's just use fixed 4 bytes for length for now, or consider binary.PutUvarint
	// For now, let's stick to a simple length-prefixed format.
	// This would require a decision on the max string length or a more complex varint.
	// For demonstration, let's assume simple length prefix (e.g., 4 bytes for length)
	// A more robust solution would use a variable-length integer encoding for length.
	// For this example, let's just write the string directly and prepend the length.
	// This approach avoids the need for a separate length prefix if we always read until EOF or know the exact length.
	// A common pattern is Type + Length + Data. Let's use that.
	length := uint64(len(s))
	lenBuf := make([]byte, binary.MaxVarintLen64) // Max size for Uvarint
	n := binary.PutUvarint(lenBuf, length)        // Encode length as Uvarint
	buf.Write(lenBuf[:n])
	buf.WriteString(s)
	return buf.Bytes()
}

// DecodeText decodes a string value from a byte slice.
// It expects the byte slice to start with TypePrefixText, followed by a variable-length
// encoded length, and then the string bytes.
// 从字节切片中解码字符串值。
// 它期望字节切片以 TypePrefixText 开头，
// 紧接着是变长编码的长度，然后是字符串字节。
func DecodeText(data []byte) (string, error) {
	if len(data) == 0 || data[0] != TypePrefixText {
		return "", fmt.Errorf("invalid prefix for Text decoding")
	}
	reader := bytes.NewReader(data[1:]) // Skip type prefix
	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read text length: %w", err)
	}
	if length > uint64(reader.Len()) {
		return "", fmt.Errorf("text length %d exceeds remaining data length %d", length, reader.Len())
	}
	textBytes := make([]byte, length)
	_, err = reader.Read(textBytes)
	if err != nil {
		return "", fmt.Errorf("failed to read text bytes: %w", err)
	}
	return string(textBytes), nil
}

// EncodeBlob encodes a byte slice (BLOB) value into a byte slice. It prefixes the data with TypePrefixBlob
// followed by a length prefix (variable length encoded) and then the blob bytes.
// 编码字节切片（BLOB）值到字节切片。它以 TypePrefixBlob 为前缀，
// 紧接着是长度前缀（变长编码），然后是 BLOB 字节。
func EncodeBlob(b []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(TypePrefixBlob)
	length := uint64(len(b))
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, length)
	buf.Write(lenBuf[:n])
	buf.Write(b)
	return buf.Bytes()
}

// DecodeBlob decodes a byte slice (BLOB) value from a byte slice.
// It expects the byte slice to start with TypePrefixBlob, followed by a variable-length
// encoded length, and then the blob bytes.
// 从字节切片中解码字节切片（BLOB）值。
// 它期望字节切片以 TypePrefixBlob 开头，
// 紧接着是变长编码的长度，然后是 BLOB 字节。
func DecodeBlob(data []byte) ([]byte, error) {
	if len(data) == 0 || data[0] != TypePrefixBlob {
		return nil, fmt.Errorf("invalid prefix for Blob decoding")
	}
	reader := bytes.NewReader(data[1:])
	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob length: %w", err)
	}
	if length > uint64(reader.Len()) {
		return nil, fmt.Errorf("blob length %d exceeds remaining data length %d", length, reader.Len())
	}
	blobBytes := make([]byte, length)
	_, err = reader.Read(blobBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob bytes: %w", err)
	}
	return blobBytes, nil
}

// EncodeDateTime encodes a time.Time value representing a DATETIME into a byte slice.
// It uses RFC3339Nano format for consistency and precision.
// 编码表示 DATETIME 的 time.Time 值到字节切片。
// 它使用 RFC3339Nano 格式以确保一致性和精度。
func EncodeDateTime(t time.Time) []byte {
	s := t.Format(time.RFC3339Nano)
	data := []byte(s)
	// Similar to text, but with TypePrefixDateTime
	var buf bytes.Buffer
	buf.WriteByte(TypePrefixDateTime)
	length := uint64(len(data))
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, length)
	buf.Write(lenBuf[:n])
	buf.Write(data)
	return buf.Bytes()
}

// DecodeDateTime decodes a time.Time value from a byte slice, expecting RFC3339Nano format.
// 解码表示 DATETIME 的 time.Time 值，期望 RFC3339Nano 格式。
func DecodeDateTime(data []byte) (time.Time, error) {
	if len(data) == 0 || data[0] != TypePrefixDateTime {
		return time.Time{}, fmt.Errorf("invalid prefix for DateTime decoding")
	}
	reader := bytes.NewReader(data[1:])
	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read datetime string length: %w", err)
	}
	if length > uint64(reader.Len()) {
		return time.Time{}, fmt.Errorf("datetime string length %d exceeds remaining data length %d", length, reader.Len())
	}
	sBytes := make([]byte, length)
	_, err = reader.Read(sBytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read datetime string bytes: %w", err)
	}
	t, err := time.Parse(time.RFC3339Nano, string(sBytes))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse datetime string '%s': %w", string(sBytes), err)
	}
	return t, nil
}

// EncodeDate encodes a time.Time value representing only a DATE into a byte slice.
// It formats the date as "YYYY-MM-DD".
// 编码表示 DATE 的 time.Time 值到字节切片。
// 它将日期格式化为“YYYY-MM-DD”。
func EncodeDate(t time.Time) []byte {
	s := t.Format("2006-01-02") // Standard SQL date format
	data := []byte(s)
	var buf bytes.Buffer
	buf.WriteByte(TypePrefixDate)
	length := uint64(len(data))
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, length)
	buf.Write(lenBuf[:n])
	buf.Write(data)
	return buf.Bytes()
}

// DecodeDate decodes a time.Time value representing only a DATE from a byte slice.
// It expects the format "YYYY-MM-DD".
// 解码表示 DATE 的 time.Time 值。
// 它期望格式为“YYYY-MM-DD”。
func DecodeDate(data []byte) (time.Time, error) {
	if len(data) == 0 || data[0] != TypePrefixDate {
		return time.Time{}, fmt.Errorf("invalid prefix for Date decoding")
	}
	reader := bytes.NewReader(data[1:])
	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read date string length: %w", err)
	}
	if length > uint64(reader.Len()) {
		return time.Time{}, fmt.Errorf("date string length %d exceeds remaining data length %d", length, reader.Len())
	}
	sBytes := make([]byte, length)
	_, err = reader.Read(sBytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read date string bytes: %w", err)
	}
	t, err := time.Parse("2006-01-02", string(sBytes))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date string '%s': %w", string(sBytes), err)
	}
	return t, nil
}

// EncodeTime encodes a time.Time value representing only a TIME into a byte slice.
// It formats the time as "HH:MM:SS".
// 编码表示 TIME 的 time.Time 值到字节切片。
// 它将时间格式化为“HH:MM:SS”。
func EncodeTime(t time.Time) []byte {
	s := t.Format("15:04:05") // Standard SQL time format
	data := []byte(s)
	var buf bytes.Buffer
	buf.WriteByte(TypePrefixTime)
	length := uint64(len(data))
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, length)
	buf.Write(lenBuf[:n])
	buf.Write(data)
	return buf.Bytes()
}

// DecodeTime decodes a time.Time value representing only a TIME from a byte slice.
// It expects the format "HH:MM:SS".
// 解码表示 TIME 的 time.Time 值。
// 它期望格式为“HH:MM:SS”。
func DecodeTime(data []byte) (time.Time, error) {
	if len(data) == 0 || data[0] != TypePrefixTime {
		return time.Time{}, fmt.Errorf("invalid prefix for Time decoding")
	}
	reader := bytes.NewReader(data[1:])
	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read time string length: %w", err)
	}
	if length > uint64(reader.Len()) {
		return time.Time{}, fmt.Errorf("time string length %d exceeds remaining data length %d", length, reader.Len())
	}
	sBytes := make([]byte, length)
	_, err = reader.Read(sBytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read time string bytes: %w", err)
	}
	t, err := time.Parse("15:04:05", string(sBytes))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time string '%s': %w", string(sBytes), err)
	}
	return t, nil
}

// PeekTypePrefix reads the first byte of a byte slice to determine the data type.
// This is useful for dispatching to the correct decoding function without knowing the full type beforehand.
// 读取字节切片的第一个字节以确定数据类型。
// 这对于在不知道完整类型的情况下分派到正确的解码函数很有用。
func PeekTypePrefix(data []byte) byte {
	if len(data) == 0 {
		return TypePrefixInvalid
	}
	return data[0]
}
