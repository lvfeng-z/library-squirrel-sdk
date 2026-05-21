package pluginsdk

import "errors"

// WindowHandle 窗口句柄
type WindowHandle interface {
	Show()
	Close()
	SetTitle(title string)
}

// WindowOptions 创建窗口的选项
type WindowOptions struct {
	Title  string
	Width  int
	Height int
	URL    string
}

// TaskCreateResult 任务创建结果
type TaskCreateResult struct {
	Succeed       bool
	AddedQuantity int
	Msg           string
}

var (
	ErrPluginCrashed = errors.New("plugin process crashed")
)
