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

// TaskCreateResult 任务创建结果
type TaskCreateResult struct {
	Succeed       bool
	AddedQuantity int
	Msg           string
}

var (
	ErrPluginCrashed = errors.New("plugin process crashed")
)
