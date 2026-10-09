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
	handler            dto.WorkFetcher
	browser            dto.SiteBrowser
	siteAuthorFetchers map[string]dto.SiteAuthorFetcher
	onActivate         func(dto.PluginContext)
	onShutdown         func()
	onSettingChanged   dto.SettingChangeHandler
}

// WithWorkFetcher 注册 WorkFetcher 扩展点（作品拉取型插件）。未设置本选项时插件进程不注册
// WorkFetchService，主程序侧任务相关 RPC 得到 gRPC Unimplemented——工具型插件
// （仅库查询/前端扩展等宿主能力）无需注册作品拉取扩展
func WithWorkFetcher(handler dto.WorkFetcher) ServeOption {
	return func(c *serveConfig) { c.handler = handler }
}

// WithBrowser 注册 SiteBrowser 扩展点
func WithBrowser(browser dto.SiteBrowser) ServeOption {
	return func(c *serveConfig) { c.browser = browser }
}

// WithSiteAuthorFetcher 注册站点作者信息拉取扩展点的一个条目实现（「站点实体元数据+资源
// 拉取」契约家族首成员）。可多次调用，每次注册一个条目：id 为 manifest 清单 siteAuthorFetch
// 条目 id（插件内唯一，同 id 后调覆盖）；拉取请求按请求内 extensionId 分派到对应实现，
// 未命中条目 id 得 codes.InvalidArgument。单实例插件注册一个条目即等价单实现形态。
// 一次不调用时插件进程不注册 SiteAuthorFetchService，主程序侧拉取 RPC 得到 gRPC
// Unimplemented；主程序仅对 manifest 声明 siteAuthorFetch 能力的插件调用
func WithSiteAuthorFetcher(id string, fetcher dto.SiteAuthorFetcher) ServeOption {
	return func(c *serveConfig) {
		if c.siteAuthorFetchers == nil {
			c.siteAuthorFetchers = make(map[string]dto.SiteAuthorFetcher)
		}
		c.siteAuthorFetchers[id] = fetcher
	}
}

// WithActivate 设置 Activate 回调（在此回调中注册扩展点和 URL 监听器）
func WithActivate(fn func(dto.PluginContext)) ServeOption {
	return func(c *serveConfig) { c.onActivate = fn }
}

// WithShutdown 设置 Shutdown 回调（插件进程被关闭前调用，用于清理资源）
func WithShutdown(fn func()) ServeOption {
	return func(c *serveConfig) { c.onShutdown = fn }
}

// WithSettingChangeHandler 注册设置变更处置函数：宿主在插件设置落库（保存/重置）
// 成功后异步推送变更通知（来源 + 键），插件在此自行反应（如重载配置热生效——
// 激活时快照的设置改后不重启进程即生效）。未设置本选项时通知为空操作（debug 日志）；
// 处置函数在 RPC goroutine 执行，须快速返回、并发安全自理，错误自行记日志（无返回值）
func WithSettingChangeHandler(handler dto.SettingChangeHandler) ServeOption {
	return func(c *serveConfig) { c.onSettingChanged = handler }
}

// Serve 启动插件进程，由插件开发者调用。全部能力以选项提供，无必填参数：
// 作品拉取型插件经 WithWorkFetcher 注册作品拉取扩展，工具型插件省略之
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
		Handler:            cfg.handler,
		Browser:            cfg.browser,
		SiteAuthorFetchers: cfg.siteAuthorFetchers,
		OnActivate:         cfg.onActivate,
		OnShutdown:         cfg.onShutdown,
		OnSettingChanged:   cfg.onSettingChanged,
	}
}
