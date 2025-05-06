// Package utils contains general-purpose internal utility functions.
// These functions are used across different packages within Guocedb.
//
// utils 包包含通用的内部工具函数。
// 这些函数在 Guocedb 的不同包中使用。
package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/turtacn/guocedb/common/log"
	// Consider adding imports for common libraries if needed (e.g., sync, io).
	// 如果需要，考虑添加常用库的导入（例如，sync, io）。
)

// TODO: Add general utility functions here as they are identified during development.
// TODO: 在此处添加通用工具函数，它们在开发过程中会确定。

// Example: A helper function to check if a context is cancelled.
// 示例：一个检查 context 是否被取消的辅助函数。
func IsContextCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		log.Debug("Context is cancelled.") // Context 已取消。
		return true
	default:
		return false
	}
}

// Example: A helper function for basic error wrapping with a message.
// 示例：一个用于使用消息进行基本错误包装的辅助函数。
func WrapError(err error, message string) error {
	if err == nil {
		return fmt.Errorf(message) // Create a new error if original is nil
	}
	return fmt.Errorf("%s: %w", message, err) // Wrap the original error
}

// Example: A helper function for timing operations (for logging or metrics).
// 示例：一个用于计时操作的辅助函数（用于日志记录或指标）。
// Use defer utils.TimeOperation("operation_name")()
func TimeOperation(name string) func() {
	startTime := time.Now()
	log.Debug("Operation '%s' started.", name) // 操作 '%s' 已开始。
	return func() {
		duration := time.Since(startTime)
		log.Debug("Operation '%s' finished in %s.", name, duration) // 操作 '%s' 在 %s 中完成。
	}
}


// TODO: Add more utility functions as needed, e.g.:
// - String manipulation helpers.
// - Slice/map manipulation helpers.
// - Concurrent programming helpers.
// - I/O helpers.
// TODO: 根据需要添加更多工具函数，例如：
// - 字符串操作辅助函数。
// - 切片/ map 操作辅助函数。
// - 并发编程辅助函数。
// - I/O 辅助函数。