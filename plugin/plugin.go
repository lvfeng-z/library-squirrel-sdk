package plugin

import (
	goPlugin "github.com/hashicorp/go-plugin"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/liveness"
	"github.com/lvfeng-z/library-squirrel-sdk/transport"
)

// ServeOption 配置插件启动选项
type ServeOption func(*serveConfig)

type serveConfig struct {
	handler    dto.TaskHandler
	browser    dto.SiteBrowser
	onActivate func(dto.PluginContext)
	onShutdown func()
}

// WithTaskHandler 注册 TaskHandler 扩展点（下载型插件）。未设置本选项时插件进程不注册
// TaskHandlerService，主程序侧任务相关 RPC 得到 gRPC Unimplemented——工具型插件
// （仅库查询/前端扩展等宿主能力）无需注册任务处理器
func WithTaskHandler(handler dto.TaskHandler) ServeOption {
	return func(c *serveConfig) { c.handler = handler }
}

// WithBrowser 注册 SiteBrowser 扩展点
func WithBrowser(browser dto.SiteBrowser) ServeOption {
	return func(c *serveConfig) { c.browser = browser }
}

// WithActivate 设置 Activate 回调（在此回调中注册扩展点和 URL 监听器）
func WithActivate(fn func(dto.PluginContext)) ServeOption {
	return func(c *serveConfig) { c.onActivate = fn }
}

// WithShutdown 设置 Shutdown 回调（插件进程被关闭前调用，用于清理资源）
func WithShutdown(fn func()) ServeOption {
	return func(c *serveConfig) { c.onShutdown = fn }
}

// Serve 启动插件进程，由插件开发者调用。全部能力以选项提供，无必填参数：
// 下载型插件经 WithTaskHandler 注册任务处理器，工具型插件省略之
func Serve(opts ...ServeOption) {
	goPlugin.Serve(&goPlugin.ServeConfig{
		HandshakeConfig: transport.Handshake,
		Plugins: map[string]goPlugin.Plugin{
			"library_squirrel": newLSPlugin(opts...),
		},
		GRPCServer: liveness.GRPCServerFactory,
	})
}

// newLSPlugin 累积启动选项并组装插件进程的 gRPC 服务描述（Serve 的可测内核）
func newLSPlugin(opts ...ServeOption) *transport.LSPlugin {
	cfg := &serveConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return &transport.LSPlugin{
		Handler:    cfg.handler,
		Browser:    cfg.browser,
		OnActivate: cfg.onActivate,
		OnShutdown: cfg.onShutdown,
	}
}
