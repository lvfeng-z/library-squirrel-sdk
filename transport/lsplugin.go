package transport

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"github.com/lvfeng-z/library-squirrel-sdk/liveness"
)

// LSPlugin 实现 hashicorp/go-plugin 的 GRPCPlugin 接口
type LSPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Handler            dto.WorkFetcher
	Browser            dto.SiteBrowser
	SiteAuthorFetchers map[string]dto.SiteAuthorFetcher
	OnActivate         func(dto.PluginContext)
	OnShutdown         func()
	OnSettingChanged   dto.SettingChangeHandler
	HostDeps           *HostDeps
}

func (p *LSPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	gen.RegisterPluginLifecycleServer(s, &lifecycleServer{
		onActivate:       p.OnActivate,
		onShutdown:       p.OnShutdown,
		onSettingChanged: p.OnSettingChanged,
		broker:           broker,
	})
	// 未提供 WorkFetcher 时不注册 WorkFetchService：gRPC 对未注册服务的调用返回
	// codes.Unimplemented，工具型插件（仅库查询/前端扩展等宿主能力）无需注册假作品拉取扩展
	if p.Handler != nil {
		gen.RegisterWorkFetchServiceServer(s, &workFetchServer{
			handler: p.Handler,
		})
	}
	if p.Browser != nil {
		gen.RegisterSiteBrowserServiceServer(s, &siteBrowserServer{
			browser: p.Browser,
		})
	}
	if len(p.SiteAuthorFetchers) > 0 {
		gen.RegisterSiteAuthorFetchServiceServer(s, &siteAuthorFetchServer{
			fetchers: p.SiteAuthorFetchers,
		})
	}
	return nil
}

func (p *LSPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (any, error) {
	var hostServiceId uint32
	if p.HostDeps != nil {
		deps := *p.HostDeps
		hostServiceId = broker.NextId()
		go broker.AcceptAndServe(hostServiceId, func(opts []grpc.ServerOption) *grpc.Server {
			// opts 含 go-plugin 注入的框架必需项(TLS/拦截器等),必须追加而非覆盖;
			// 追加消息接收上限覆盖插件→宿主方向的大消息(如大库会话快照写库)
			s := grpc.NewServer(append(opts, grpc.MaxRecvMsgSize(liveness.MaxRecvMsgSize))...)
			RegisterHostService(s, deps)
			return s
		})
	}

	return &GRPCPluginClient{
		Lifecycle:       gen.NewPluginLifecycleClient(c),
		WorkFetch:       gen.NewWorkFetchServiceClient(c),
		Browser:         gen.NewSiteBrowserServiceClient(c),
		SiteAuthorFetch: gen.NewSiteAuthorFetchServiceClient(c),
		HostServiceId:   hostServiceId,
	}, nil
}
