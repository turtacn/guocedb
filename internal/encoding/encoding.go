package encoding

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"

	"github.com/turtacn/guocedb/common/errors"
)

// ByteOrder is the standard byte order used for encoding numerical types.
// We use BigEndian for network compatibility and often better sortability
// when keys are compared lexicographically.
// ByteOrder 是用于编码数值类型的标准字节顺序。
// 我们使用 BigEndian 以实现网络兼容性，并且在按字典顺序比较键时通常具有更好的可排序性。
var ByteOrder = binary.BigEndian

const (
	// Escape byte used in string encoding to handle zero bytes.
	// 在字符串编码中用于处理零字节的转义字节。
	escapeByte byte = 0x01
	// Marker for an escaped zero byte (escapeByte + 1).
	// 转义零字节的标记 (escapeByte + 1)。
	escapedZero byte = 0x01 // escapeByte + 0 -> invalid sequence, use escapeByte + 1 instead
	// Marker for an escaped escape byte (escapeByte + 2).
	// 转义转义字节的标记 (escapeByte + 2)。
	escapedEscape byte = 0x02 // escapeByte + 1 -> escapedZero, use escapeByte + 2 instead
	// Terminator byte for encoded strings.
	// 编码字符串的终止符字节。
	stringTerminator byte = 0x00
)

// --- Integer Encoding ---

// EncodeUint64 appends the BigEndian binary representation of a uint64 to the buffer.
// This encoding is sortable for unsigned integers.
// EncodeUint64 将 uint64 的 BigEndian 二进制表示形式附加到缓冲区。
// 此编码对于无符号整数是可排序的。
func EncodeUint64(buf []byte, v uint64) []byte {
	var encoded [8]byte
	ByteOrder.PutUint64(encoded[:], v)
	return append(buf, encoded[:]...)
}

// DecodeUint64 decodes a uint64 from the beginning of the buffer using BigEndian.
// It returns the decoded value, the remaining buffer, and any error.
// DecodeUint64 使用 BigEndian 从缓冲区开头解码 uint64。
// 它返回解码后的值、剩余的缓冲区和任何错误。
func DecodeUint64(buf []byte) (uint64, []byte, error) {
	if len(buf) < 8 {
		return 0, buf, errors.Newf(errors.ErrCodeSerialization, "buffer too short to decode uint64: need 8, got %d", len(buf))
	}
	val := ByteOrder.Uint64(buf[:8])
	return val, buf[8:], nil
}

// EncodeInt64 appends a sortable BigEndian binary representation of an int64 to the buffer.
// It flips the sign bit to ensure negative numbers sort before positive numbers.
// EncodeInt64 将 int64 的可排序 BigEndian 二进制表示形式附加到缓冲区。
// 它翻转符号位以确保负数排在正数之前。
func EncodeInt64(buf []byte, v int64) []byte {
	// Flip the sign bit (XOR with 1 << 63) to make it sortable
	// 0x8000000000000000 is 1 << 63
	// 翻转符号位（与 1 << 63 进行异或）使其可排序
	// 0x8000000000000000 是 1 << 63
	uv := uint64(v) ^ (1 << 63)
	return EncodeUint64(buf, uv)
}

// DecodeInt64 decodes a sortable int64 from the beginning of the buffer.
// It reads a uint64 and flips the sign bit back.
// DecodeInt64 从缓冲区开头解码可排序的 int64。
// 它读取一个 uint64 并将符号位翻转回来。
func DecodeInt64(buf []byte) (int64, []byte, error) {
	uv, remainder, err := DecodeUint64(buf)
	if err != nil {
		// Don't wrap here, DecodeUint64 already created a suitable error
		// 这里不要包装，DecodeUint64 已经创建了一个合适的错误
		return 0, buf, err
	}
	// Flip the sign bit back
	// 将符号位翻转回来
	v := int64(uv ^ (1 << 63))
	return v, remainder, nil
}

// --- Float Encoding ---

// EncodeFloat64 appends a sortable BigEndian binary representation of a float64 to the buffer.
// It handles the sign bit correctly for sorting positive and negative numbers, including Inf and NaN.
// EncodeFloat64 将 float64 的可排序 BigEndian 二进制表示形式附加到缓冲区。
// 它正确处理符号位以对正数和负数进行排序，包括 Inf 和 NaN。
func EncodeFloat64(buf []byte, v float64) []byte {
	uv := math.Float64bits(v)
	// If negative, invert all bits
	// 如果为负，则反转所有位
	if v < 0 {
		uv = ^uv
	} else {
		// If positive or zero, flip the sign bit (MSB)
		// 如果为正或零，则翻转符号位 (MSB)
		uv = uv ^ (1 << 63)
	}
	return EncodeUint64(buf, uv)
}

// DecodeFloat64 decodes a sortable float64 from the beginning of the buffer.
// DecodeFloat64 从缓冲区开头解码可排序的 float64。
func DecodeFloat64(buf []byte) (float64, []byte, error) {
	uv, remainder, err := DecodeUint64(buf)
	if err != nil {
		return 0, buf, err
	}

	// Check the original sign bit (MSB of the encoded uint64)
	// 检查原始符号位（编码后 uint64 的 MSB）
	if uv&(1<<63) == 0 { // Original was negative (encoded MSB is 0 after inversion)
		// 如果原始值为负（反转后编码的 MSB 为 0）
		uv = ^uv // Invert back // 反转回来
	} else { // Original was non-negative (encoded MSB is 1 after sign flip)
		// 如果原始值为非负（符号翻转后编码的 MSB 为 1）
		uv = uv ^ (1 << 63) // Flip sign bit back // 将符号位翻转回来
	}

	v := math.Float64frombits(uv)
	return v, remainder, nil
}

// --- String Encoding ---

// EncodeString appends a sortable representation of a string to the buffer.
// It escapes the stringTerminator (0x00) and escapeByte (0x01) within the string
// and appends a stringTerminator (0x00) at the end.
// This allows strings containing null bytes to be encoded correctly and ensures
// that lexicographical comparison of the encoded bytes matches string comparison.
// Example: "a" -> ['a', 0x00]
//
//	"a\x00b" -> ['a', 0x01, 0x01, 'b', 0x00] (0x01, 0x01 represents escaped 0x00)
//	"a\x01b" -> ['a', 0x01, 0x02, 'b', 0x00] (0x01, 0x02 represents escaped 0x01)
//
// EncodeString 将字符串的可排序表示形式附加到缓冲区。
// 它对字符串中的 stringTerminator (0x00) 和 escapeByte (0x01) 进行转义，
// 并在末尾附加一个 stringTerminator (0x00)。
// 这允许正确编码包含空字节的字符串，并确保
// 编码字节的字典比较与字符串比较匹配。
// 示例： "a" -> ['a', 0x00]
//
//	"a\x00b" -> ['a', 0x01, 0x01, 'b', 0x00] (0x01, 0x01 表示转义的 0x00)
//	"a\x01b" -> ['a', 0x01, 0x02, 'b', 0x00] (0x01, 0x02 表示转义的 0x01)
func EncodeString(buf []byte, v string) []byte {
	// Pre-allocate buffer assuming minimal escaping for efficiency
	// 预分配缓冲区，假设最少的转义以提高效率
	encoded := make([]byte, 0, len(v)+1) // +1 for terminator // +1 用于终止符

	for i := 0; i < len(v); i++ {
		b := v[i]
		if b == stringTerminator { // escape 0x00
			encoded = append(encoded, escapeByte, escapedZero)
		} else if b == escapeByte { // escape 0x01
			encoded = append(encoded, escapeByte, escapedEscape)
		} else {
			encoded = append(encoded, b)
		}
	}
	encoded = append(encoded, stringTerminator) // Add terminator // 添加终止符
	return append(buf, encoded...)
}

// DecodeString decodes a sortable string from the beginning of the buffer.
// It reads bytes until the stringTerminator (0x00) is found, handling escaped sequences.
// Returns the decoded string, the remaining buffer, and any error.
// DecodeString 从缓冲区开头解码可排序的字符串。
// 它读取字节直到找到 stringTerminator (0x00)，并处理转义序列。
// 返回解码后的字符串、剩余的缓冲区和任何错误。
func DecodeString(buf []byte) (string, []byte, error) {
	// Find the first occurrence of the non-escaped terminator
	// 查找第一个未转义终止符的出现位置
	idx := -1
	for i := 0; i < len(buf); i++ {
		if buf[i] == stringTerminator {
			idx = i
			break
		} else if buf[i] == escapeByte {
			// Skip the next byte as it's part of the escape sequence
			// 跳过下一个字节，因为它是转义序列的一部分
			i++
			if i >= len(buf) {
				// Unterminated escape sequence at the end of the buffer
				// 缓冲区末尾未终止的转义序列
				return "", buf, errors.New(errors.ErrCodeSerialization, "malformed string encoding: unterminated escape sequence")
			}
		}
	}

	if idx == -1 {
		// Terminator not found
		// 未找到终止符
		return "", buf, errors.New(errors.ErrCodeSerialization, "malformed string encoding: missing terminator")
	}

	encodedBytes := buf[:idx]
	remainder := buf[idx+1:] // Skip the terminator // 跳过终止符

	// Pre-allocate assuming minimal escaping
	// 预分配，假设最少的转义
	decoded := make([]byte, 0, len(encodedBytes))
	i := 0
	for i < len(encodedBytes) {
		b := encodedBytes[i]
		if b == escapeByte {
			i++ // Move to the escaped value // 移动到转义后的值
			if i >= len(encodedBytes) {
				// Should have been caught by the initial scan, but double-check
				// 应该在初始扫描中被捕获，但再次检查
				return "", buf, errors.New(errors.ErrCodeSerialization, "malformed string encoding: invalid escape sequence at end")
			}
			escapedVal := encodedBytes[i]
			if escapedVal == escapedZero {
				decoded = append(decoded, stringTerminator) // 0x01 0x01 -> 0x00
			} else if escapedVal == escapedEscape {
				decoded = append(decoded, escapeByte) // 0x01 0x02 -> 0x01
			} else {
				// Invalid escape sequence
				// 无效的转义序列
				return "", buf, errors.Newf(errors.ErrCodeSerialization, "malformed string encoding: invalid escape sequence %x%x", escapeByte, escapedVal)
			}
		} else {
			decoded = append(decoded, b)
		}
		i++
	}

	return string(decoded), remainder, nil
}

// --- Time Encoding ---

// EncodeTime appends a sortable representation of a time.Time value to the buffer.
// It encodes the time as nanoseconds since the Unix epoch (int64).
// EncodeTime 将 time.Time 值的可排序表示形式附加到缓冲区。
// 它将时间编码为自 Unix 纪元以来的纳秒数 (int64)。
func EncodeTime(buf []byte, t time.Time) []byte {
	// Convert to nanoseconds since epoch for high precision
	// 转换为自纪元以来的纳秒数以获得高精度
	nanos := t.UnixNano()
	return EncodeInt64(buf, nanos)
}

// DecodeTime decodes a sortable time.Time value from the beginning of the buffer.
// It decodes an int64 representing nanoseconds since the Unix epoch.
// DecodeTime 从缓冲区开头解码可排序的 time.Time 值。
// 它解码表示自 Unix 纪元以来纳秒数的 int64。
func DecodeTime(buf []byte) (time.Time, []byte, error) {
	nanos, remainder, err := DecodeInt64(buf)
	if err != nil {
		return time.Time{}, buf, err // Propagate error // 传播错误
	}
	// Handle potential zero time representation if needed, though UnixNano handles zero time correctly.
	// 如果需要，处理可能的零时间表示，尽管 UnixNano 可以正确处理零时间。
	t := time.Unix(0, nanos).UTC() // Reconstruct time, store in UTC for consistency // 重构时间，以 UTC 存储以保持一致性
	return t, remainder, nil
}

// --- Boolean Encoding ---

// EncodeBool appends a sortable representation of a boolean value to the buffer.
// false -> 0x00, true -> 0x01
// EncodeBool 将布尔值的可排序表示形式附加到缓冲区。
// false -> 0x00, true -> 0x01
func EncodeBool(buf []byte, v bool) []byte {
	if v {
		return append(buf, 0x01)
	}
	return append(buf, 0x00)
}

// DecodeBool decodes a sortable boolean value from the beginning of the buffer.
// Reads one byte: 0x00 -> false, 0x01 -> true. Other values are errors.
// DecodeBool 从缓冲区开头解码可排序的布尔值。
// 读取一个字节：0x00 -> false, 0x01 -> true。其他值为错误。
func DecodeBool(buf []byte) (bool, []byte, error) {
	if len(buf) < 1 {
		return false, buf, errors.New(errors.ErrCodeSerialization, "buffer too short to decode bool: need 1 byte")
	}
	b := buf[0]
	remainder := buf[1:]
	if b == 0x00 {
		return false, remainder, nil
	}
	if b == 0x01 {
		return true, remainder, nil
	}
	return false, buf, errors.Newf(errors.ErrCodeSerialization, "invalid byte value for bool: expected 0x00 or 0x01, got 0x%02x", b)
}

// --- Composite Key Encoding Helpers ---

// EncodeBytes appends byte slice data directly, suitable for parts of keys
// where sortability is handled by preceding encoded types or where direct
// byte comparison is desired. It does NOT add terminators or perform escaping.
// Use with caution, primarily for raw byte segments within a larger structured key.
// EncodeBytes 直接附加字节切片数据，适用于键的某些部分，
// 其中可排序性由前面的编码类型处理，或者需要直接字节比较。
// 它不添加终止符或执行转义。
// 请谨慎使用，主要用于较大结构化键中的原始字节段。
func EncodeBytes(buf []byte, data []byte) []byte {
	return append(buf, data...)
}

// DecodeBytes reads a fixed number of bytes from the buffer.
// DecodeBytes 从缓冲区读取固定数量的字节。
func DecodeBytes(buf []byte, length int) ([]byte, []byte, error) {
	if len(buf) < length {
		return nil, buf, errors.Newf(errors.ErrCodeSerialization, "buffer too short to decode bytes: need %d, got %d", length, len(buf))
	}
	data := buf[:length]
	remainder := buf[length:]
	// Return a copy to prevent modification issues if the original buffer is reused.
	// 返回一个副本以防止在原始缓冲区被重用时出现修改问题。
	dataCopy := make([]byte, length)
	copy(dataCopy, data)
	return dataCopy, remainder, nil
}

// DecodeRemainingBytes returns all remaining bytes in the buffer.
// Useful for the last component of a key or value.
// DecodeRemainingBytes 返回缓冲区中所有剩余的字节。
// 对于键或值的最后一个组件很有用。
func DecodeRemainingBytes(buf []byte) ([]byte, []byte, error) {
	if len(buf) == 0 {
		// Return nil slice for data, empty slice for remainder
		// 为数据返回 nil 切片，为剩余部分返回空切片
		return nil, buf, nil
	}
	// Return a copy
	// 返回一个副本
	dataCopy := make([]byte, len(buf))
	copy(dataCopy, buf)
	// Remainder is empty
	// 剩余部分为空
	return dataCopy, buf[len(buf):], nil
}

// CompareEncoded compares two byte slices assumed to be encoded using the
// sortable methods in this package. It provides a standard comparison function.
// Returns:
//
//	-1 if a < b
//	 0 if a == b
//	+1 if a > b
//
// CompareEncoded 比较两个假定使用此包中可排序方法编码的字节切片。
// 它提供了一个标准的比较函数。
// 返回：
//
//	-1 如果 a < b
//	 0 如果 a == b
//	+1 如果 a > b
func CompareEncoded(a, b []byte) int {
	return bytes.Compare(a, b)
}
