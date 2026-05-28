package transport

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/lvfeng-z/library-squirrel-plugin-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-plugin-sdk/gen"
)

// LSPlugin 实现 hashicorp/go-plugin 的 GRPCPlugin 接口
type LSPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Handler    dto.TaskHandler
	Browser    dto.SiteBrowser
	OnActivate func(dto.PluginContext)
	HostDeps   *HostDeps
}

func (p *LSPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	gen.RegisterPluginLifecycleServer(s, &lifecycleServer{
		onActivate: p.OnActivate,
		broker:     broker,
	})
	gen.RegisterTaskHandlerServiceServer(s, &taskHandlerServer{
		handler: p.Handler,
	})
	if p.Browser != nil {
		gen.RegisterSiteBrowserServiceServer(s, &siteBrowserServer{
			browser: p.Browser,
		})
	}
	return nil
}

func (p *LSPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	var hostServiceId uint32
	if p.HostDeps != nil {
		deps := *p.HostDeps
		hostServiceId = broker.NextId()
		go broker.AcceptAndServe(hostServiceId, func(opts []grpc.ServerOption) *grpc.Server {
			s := grpc.NewServer(opts...)
			RegisterHostService(s, deps)
			return s
		})
	}

	return &GRPCPluginClient{
		Lifecycle:     gen.NewPluginLifecycleClient(c),
		Task:          gen.NewTaskHandlerServiceClient(c),
		Browser:       gen.NewSiteBrowserServiceClient(c),
		HostServiceId: hostServiceId,
	}, nil
}
