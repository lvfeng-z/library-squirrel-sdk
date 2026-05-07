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

// PluginContextInit 插件初始化参数（由主进程通过 activate 通知传入）
type PluginContextInit struct {
	PluginPublicID string `json:"pluginPublicId"`
	PluginData     string `json:"pluginData"`
	RootPath       string `json:"rootPath"`
}

// StartResponse Start 方法的响应体
type StartResponse struct {
	WorkResponse *WorkResponse `json:"workResponse"`
	StreamID     string        `json:"streamId"`
}

// StreamEnd 流结束信号
type StreamEnd struct {
	StreamID string
	Error   error
}

// PingRequest 心跳请求
type PingRequest struct{}

// PingResponse 心跳响应
type PingResponse struct {
	Timestamp int64
}

// Error codes for RPC
var (
	ErrStreamNotFound = errors.New("stream not found")
	ErrPluginCrashed  = errors.New("plugin process crashed")
)
