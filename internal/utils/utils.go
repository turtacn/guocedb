package utils

import (
	"fmt"
	"os"
	"reflect" // Used for generic slice contains check, use with caution
	"strconv"
	"strings"
)

// --- File System Utilities ---
// --- 文件系统实用程序 ---

// FileExists checks if a file or directory exists at the given path.
// FileExists 检查给定路径的文件或目录是否存在。
func FileExists(path string) bool {
	_, err := os.Stat(path)
	// Check if error is "not exist"
	// 检查错误是否为“不存在”
	// return !os.IsNotExist(err)
	// A more explicit check: nil means exists, IsNotExist means doesn't exist.
	// 更明确的检查：nil 表示存在，IsNotExist 表示不存在。
	if err == nil {
		return true // File or directory exists // 文件或目录存在
	}
	if os.IsNotExist(err) {
		return false // File or directory does not exist // 文件或目录不存在
	}
	// Other errors (e.g., permission denied) are treated as "doesn't exist"
	// for simplicity in many checks, but could be handled differently if needed.
	// 为简化许多检查，其他错误（例如权限被拒绝）被视为“不存在”，
	// 但如果需要，可以不同地处理。
	// Consider logging err here if unexpected errors are important.
	// 如果意外错误很重要，请考虑在此处记录错误。
	// log.Warnf("Unexpected error checking file existence for '%s': %v", path, err)
	return false
}

// IsDir checks if the given path exists and is a directory.
// IsDir 检查给定路径是否存在并且是目录。
func IsDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Doesn't exist, so not a directory // 不存在，因此不是目录
		}
		return false, fmt.Errorf("failed to stat path '%s': %w", path, err) // Other error // 其他错误
	}
	return stat.IsDir(), nil
}

// IsFile checks if the given path exists and is a regular file.
// IsFile 检查给定路径是否存在并且是常规文件。
func IsFile(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Doesn't exist, so not a file // 不存在，因此不是文件
		}
		return false, fmt.Errorf("failed to stat path '%s': %w", path, err) // Other error // 其他错误
	}
	// Check if it's not a directory and not other special types (symlink, etc.)
	// 检查它是否不是目录，也不是其他特殊类型（符号链接等）
	return !stat.IsDir() && stat.Mode().IsRegular(), nil
}

// EnsureDir creates a directory at the given path if it doesn't already exist.
// It creates parent directories as needed (like mkdir -p).
// EnsureDir 如果给定路径的目录尚不存在，则创建该目录。
// 它根据需要创建父目录（类似于 mkdir -p）。
func EnsureDir(path string) error {
	// os.MkdirAll returns nil if the directory already exists.
	// 如果目录已存在，os.MkdirAll 返回 nil。
	// 0755 provides read/write/execute for owner, read/execute for group/others. Common default.
	// 0755 为所有者提供读/写/执行权限，为组/其他人提供读/执行权限。常见的默认设置。
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", path, err)
	}
	return nil
}

// --- String Utilities ---
// --- 字符串实用程序 ---

// ToLowerCamelCase converts a snake_case or kebab-case string to lowerCamelCase.
// Example: "user_id" -> "userId", "log-level" -> "logLevel"
// ToLowerCamelCase 将蛇形命名法（snake_case）或短横线命名法（kebab-case）字符串转换为小驼峰式命名法（lowerCamelCase）。
// 示例："user_id" -> "userId", "log-level" -> "logLevel"
func ToLowerCamelCase(s string) string {
	var result strings.Builder
	nextUpper := false
	for i, r := range s {
		if r == '_' || r == '-' {
			nextUpper = true
		} else if nextUpper {
			result.WriteRune(strings.ToUpper(string(r))[0]) // Convert rune to uppercase // 将 rune 转换为大写
			nextUpper = false
		} else {
			// First character should be lowercase // 第一个字符应为小写
			if i == 0 {
				result.WriteRune(strings.ToLower(string(r))[0])
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

// --- Pointer Dereferencing Helpers ---
// These helpers safely dereference pointers, returning a default value if the pointer is nil.
// --- 指针解引用辅助函数 ---
// 这些辅助函数安全地解引用指针，如果指针为 nil，则返回默认值。

// PointerToString safely dereferences a *string, returning "" if nil.
// PointerToString 安全地解引用 *string，如果为 nil 则返回 ""。
func PointerToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// PointerToStringDef safely dereferences a *string, returning a default value if nil.
// PointerToStringDef 安全地解引用 *string，如果为 nil 则返回默认值。
func PointerToStringDef(ptr *string, def string) string {
	if ptr == nil {
		return def
	}
	return *ptr
}

// PointerToInt64 safely dereferences an *int64, returning 0 if nil.
// PointerToInt64 安全地解引用 *int64，如果为 nil 则返回 0。
func PointerToInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// PointerToInt64Def safely dereferences an *int64, returning a default value if nil.
// PointerToInt64Def 安全地解引用 *int64，如果为 nil 则返回默认值。
func PointerToInt64Def(ptr *int64, def int64) int64 {
	if ptr == nil {
		return def
	}
	return *ptr
}

// PointerToBool safely dereferences a *bool, returning false if nil.
// PointerToBool 安全地解引用 *bool，如果为 nil 则返回 false。
func PointerToBool(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

// PointerToBoolDef safely dereferences a *bool, returning a default value if nil.
// PointerToBoolDef 安全地解引用 *bool，如果为 nil 则返回默认值。
func PointerToBoolDef(ptr *bool, def bool) bool {
	if ptr == nil {
		return def
	}
	return *ptr
}

// --- Slice Utilities ---
// --- 切片实用程序 ---

// StringInSlice checks if a string exists within a slice of strings.
// StringInSlice 检查字符串是否存在于字符串切片中。
func StringInSlice(target string, slice []string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// Int64InSlice checks if an int64 exists within a slice of int64s.
// Int64InSlice 检查 int64 是否存在于 int64 切片中。
func Int64InSlice(target int64, slice []int64) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// Contains checks if a slice contains a specific element.
// Uses reflection, so it's less performant than type-specific versions like StringInSlice.
// Use sparingly or when generics are not available/suitable.
// Contains 检查切片是否包含特定元素。
// 使用反射，因此性能低于类型特定版本（如 StringInSlice）。
// 请谨慎使用，或在泛型不可用/不适用时使用。
func Contains(slice interface{}, element interface{}) bool {
	sliceValue := reflect.ValueOf(slice)

	if sliceValue.Kind() != reflect.Slice {
		// Or return an error? For simplicity, return false.
		// 或者返回错误？为简单起见，返回 false。
		return false
	}

	for i := 0; i < sliceValue.Len(); i++ {
		// Use reflect.DeepEqual for comparison to handle complex types,
		// but simple equality check might be faster for basic types if guaranteed.
		// 使用 reflect.DeepEqual 进行比较以处理复杂类型，
		// 但如果保证是基本类型，简单的相等性检查可能更快。
		if reflect.DeepEqual(sliceValue.Index(i).Interface(), element) {
			// if sliceValue.Index(i).Interface() == element { // Faster for comparable types // 对于可比较类型更快
			return true
		}
	}

	return false
}

// RemoveStringFromSlice creates a new slice excluding all occurrences of the target string.
// Preserves the order of the remaining elements.
// RemoveStringFromSlice 创建一个新切片，排除目标字符串的所有出现。
// 保留剩余元素的顺序。
func RemoveStringFromSlice(slice []string, target string) []string {
	result := make([]string, 0, len(slice)) // Pre-allocate capacity // 预分配容量
	for _, item := range slice {
		if item != target {
			result = append(result, item)
		}
	}
	// If the length hasn't changed, return the original slice to avoid allocation?
	// Requires tracking if any element was actually removed.
	// For simplicity, always return the new slice (which might be the same underlying array initially).
	// 如果长度没有改变，是否返回原始切片以避免分配？
	// 需要跟踪是否有任何元素实际被移除。
	// 为简单起见，始终返回新切片（最初可能是相同的底层数组）。
	return result
}

// --- Type Conversion Utilities ---
// --- 类型转换实用程序 ---

// StringToInt converts a string to an int, returning an error on failure.
// StringToInt 将字符串转换为 int，失败时返回错误。
func StringToInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string '%s' to int: %w", s, err)
	}
	return i, nil
}

// StringToIntDef converts a string to an int, returning a default value on failure.
// StringToIntDef 将字符串转换为 int，失败时返回默认值。
func StringToIntDef(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

// StringToInt64 converts a string to an int64, returning an error on failure.
// StringToInt64 将字符串转换为 int64，失败时返回错误。
func StringToInt64(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64) // Base 10, 64-bit // 10 进制，64 位
	if err != nil {
		return 0, fmt.Errorf("failed to convert string '%s' to int64: %w", s, err)
	}
	return i, nil
}

// StringToInt64Def converts a string to an int64, returning a default value on failure.
// StringToInt64Def 将字符串转换为 int64，失败时返回默认值。
func StringToInt64Def(s string, def int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return i
}

// StringToBool converts a string to a bool, returning an error on failure.
// Recognizes "true", "false", "1", "0". Case-insensitive.
// StringToBool 将字符串转换为 bool，失败时返回错误。
// 识别 "true", "false", "1", "0"。不区分大小写。
func StringToBool(s string) (bool, error) {
	sLower := strings.ToLower(strings.TrimSpace(s))
	switch sLower {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		b, err := strconv.ParseBool(s) // Standard library handles "true", "false" case-insensitively
		if err != nil {
			return false, fmt.Errorf("failed to convert string '%s' to bool: %w", s, err)
		}
		return b, nil // Should not be reached due to switch, but for completeness // 由于 switch 不应到达，但为了完整性
	}
}

// StringToBoolDef converts a string to a bool, returning a default value on failure.
// StringToBoolDef 将字符串转换为 bool，失败时返回默认值。
func StringToBoolDef(s string, def bool) bool {
	val, err := StringToBool(s)
	if err != nil {
		return def
	}
	return val
}
