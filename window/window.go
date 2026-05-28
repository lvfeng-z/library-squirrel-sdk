package window

import (
	"github.com/lvfeng-z/library-squirrel-plugin-sdk/dto"
)

// OpenWindow 创建一个 WebView2 弹窗。
// ownerHWND 为主窗口的原生句柄，用于设置弹窗的 Owner 属性（保持模态关系），可为 0。
func OpenWindow(options dto.WindowOptions, ownerHWND uintptr) (dto.WindowHandle, error) {
	return openWindow(options, ownerHWND)
}
