package pluginsdk

import "errors"

// WindowHandle 窗口句柄
type WindowHandle interface {
	Close()
	SetTitle(title string)
	// WaitForNavigation 阻塞等待导航到匹配 urlPrefix 的 URL，返回被拦截的完整 URL
	WaitForNavigation(urlPrefix string, timeoutMs int64) (string, error)
	// ExecuteScript 在窗口中执行 JavaScript 并返回 JSON 结果
	ExecuteScript(js string) (string, error)
	// Done 返回窗口关闭信号 channel
	Done() <-chan struct{}
}

// WindowOptions 创建窗口的选项
type WindowOptions struct {
	Title                string
	Width                int
	Height               int
	URL                  string
	DataPath             string               // WebView2 用户数据目录
	OnNavigationStarting func(uri string) bool // 返回 false 阻止导航
}

// CreateTaskResult 主程序 CreateTask 的返回结果
type CreateTaskResult struct {
	Succeed       bool
	AddedQuantity int
	Msg           string
}

// TaskCreateResult 插件 Create 方法的返回结果，支持批量或流式
type TaskCreateResult struct {
	array  []*TaskCreateResponse
	stream <-chan *TaskCreateResponse
	isStream bool
}

// BatchResult 创建批量模式的结果
func BatchResult(responses []*TaskCreateResponse) *TaskCreateResult {
	return &TaskCreateResult{array: responses}
}

// StreamResult 创建流式模式的结果
func StreamResult(ch <-chan *TaskCreateResponse) *TaskCreateResult {
	return &TaskCreateResult{stream: ch, isStream: true}
}

// IsStream 是否为流式模式
func (r *TaskCreateResult) IsStream() bool {
	return r.isStream
}

// Array 获取批量结果（仅批量模式有效）
func (r *TaskCreateResult) Array() []*TaskCreateResponse {
	return r.array
}

// Stream 获取流式 channel（仅流式模式有效）
func (r *TaskCreateResult) Stream() <-chan *TaskCreateResponse {
	return r.stream
}

var (
	ErrPluginCrashed = errors.New("plugin process crashed")
)
