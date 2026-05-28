package plugin

import (
	goPlugin "github.com/hashicorp/go-plugin"

	"github.com/lvfeng-z/library-squirrel-plugin-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-plugin-sdk/transport"
)

// ServeOption 配置插件启动选项
type ServeOption func(*serveConfig)

type serveConfig struct {
	browser    dto.SiteBrowser
	onActivate func(dto.PluginContext)
}

// WithBrowser 注册 SiteBrowser 扩展点
func WithBrowser(browser dto.SiteBrowser) ServeOption {
	return func(c *serveConfig) { c.browser = browser }
}

// WithActivate 设置 Activate 回调（在此回调中注册扩展点和 URL 监听器）
func WithActivate(fn func(dto.PluginContext)) ServeOption {
	return func(c *serveConfig) { c.onActivate = fn }
}

// Serve 启动插件进程，由插件开发者调用
func Serve(handler dto.TaskHandler, opts ...ServeOption) {
	cfg := &serveConfig{}
	for _, o := range opts {
		o(cfg)
	}

	lsPlugin := &transport.LSPlugin{
		Handler:    handler,
		Browser:    cfg.browser,
		OnActivate: cfg.onActivate,
	}

	goPlugin.Serve(&goPlugin.ServeConfig{
		HandshakeConfig: transport.Handshake,
		Plugins: map[string]goPlugin.Plugin{
			"library_squirrel": lsPlugin,
		},
		GRPCServer: goPlugin.DefaultGRPCServer,
	})
}
