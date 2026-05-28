package transport

import "fmt"

// FormatLogArgs 格式化日志参数（供 HostService 日志使用）
func FormatLogArgs(template string, args ...any) string {
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}
