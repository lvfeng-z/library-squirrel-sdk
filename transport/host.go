package transport

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
)

// HostDeps 主程序侧提供给 HostService 的依赖
type HostDeps struct {
	dto.StorageProvider
	dto.PluginRootProvider
	dto.WorkSetQueryProvider
	dto.SiteSaveProvider
	dto.TaskCreateProvider
	dto.UrlListenerRegistry
	dto.FrontendEventProvider
	LogFunc                 func(level int32, template string, args []string, loggerName string)
	OnRegisterTaskHandler   func(contributionId, name, description string) error
	OnRegisterSiteBrowser   func(contributionId, name, description string) error
	OnUnregisterSiteBrowser func(contributionId string) error
}

// HostServiceServer HostService 的 gRPC 服务端实现
type HostServiceServer struct {
	gen.UnimplementedHostServiceServer
	deps HostDeps
}

// NewHostServiceServer 创建 HostService gRPC 服务端
func NewHostServiceServer(deps HostDeps) *HostServiceServer {
	return &HostServiceServer{deps: deps}
}

// RegisterHostService 将 HostService 注册到 gRPC server
func RegisterHostService(s *grpc.Server, deps HostDeps) {
	gen.RegisterHostServiceServer(s, NewHostServiceServer(deps))
}

func (s *HostServiceServer) RegisterTaskHandler(ctx context.Context, req *gen.RegisterExtensionRequest) (*gen.Empty, error) {
	if s.deps.OnRegisterTaskHandler != nil {
		return &gen.Empty{}, s.deps.OnRegisterTaskHandler(req.ContributionId, req.Name, req.Description)
	}
	return &gen.Empty{}, nil
}

func (s *HostServiceServer) RegisterSiteBrowser(ctx context.Context, req *gen.RegisterExtensionRequest) (*gen.Empty, error) {
	if s.deps.OnRegisterSiteBrowser != nil {
		return &gen.Empty{}, s.deps.OnRegisterSiteBrowser(req.ContributionId, req.Name, req.Description)
	}
	return &gen.Empty{}, nil
}

func (s *HostServiceServer) UnregisterSiteBrowser(ctx context.Context, req *gen.UnregisterRequest) (*gen.Empty, error) {
	if s.deps.OnUnregisterSiteBrowser != nil {
		return &gen.Empty{}, s.deps.OnUnregisterSiteBrowser(req.ContributionId)
	}
	return &gen.Empty{}, nil
}

func (s *HostServiceServer) GetValue(ctx context.Context, req *gen.StorageKeyRequest) (*gen.StorageValueResponse, error) {
	value, err := s.deps.GetValue(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	return &gen.StorageValueResponse{Value: value}, nil
}

func (s *HostServiceServer) SetValue(ctx context.Context, req *gen.StorageEntryRequest) (*gen.Empty, error) {
	return &gen.Empty{}, s.deps.SetValue(ctx, req.Key, req.Value)
}

func (s *HostServiceServer) SetValueEncrypted(ctx context.Context, req *gen.StorageEntryRequest) (*gen.Empty, error) {
	return &gen.Empty{}, s.deps.SetValueEncrypted(ctx, req.Key, req.Value)
}

func (s *HostServiceServer) DeleteValue(ctx context.Context, req *gen.StorageKeyRequest) (*gen.Empty, error) {
	return &gen.Empty{}, s.deps.DeleteValue(ctx, req.Key)
}

func (s *HostServiceServer) GetAllValues(ctx context.Context, req *gen.Empty) (*gen.AllStorageValuesResponse, error) {
	values, err := s.deps.GetAllValues(ctx)
	if err != nil {
		return nil, err
	}
	return &gen.AllStorageValuesResponse{Values: values}, nil
}

func (s *HostServiceServer) GetWorkSetBySiteWorkSetId(ctx context.Context, req *gen.WorkSetQueryRequest) (*gen.WorkSetQueryResponse, error) {
	ws, err := s.deps.GetWorkSetBySiteWorkSetId(ctx, req.SiteWorkSetId, req.SiteName)
	if err != nil {
		return nil, err
	}
	return &gen.WorkSetQueryResponse{WorkSet: workSetToProto(ws)}, nil
}

func (s *HostServiceServer) AddSite(ctx context.Context, req *gen.AddSiteRequest) (*gen.Empty, error) {
	sites := make([]*dto.SiteDTO, len(req.Sites))
	for i, ps := range req.Sites {
		sites[i] = protoToSite(ps)
	}
	return &gen.Empty{}, s.deps.AddSite(ctx, sites)
}

func (s *HostServiceServer) RegisterUrlListener(ctx context.Context, req *gen.UrlListenerRequest) (*gen.Empty, error) {
	return &gen.Empty{}, s.deps.RegisterUrlListener(ctx, req.ContributionId, req.Patterns)
}

func (s *HostServiceServer) UnregisterUrlListener(ctx context.Context, req *gen.Empty) (*gen.Empty, error) {
	return &gen.Empty{}, s.deps.UnregisterUrlListener(ctx)
}

func (s *HostServiceServer) CreateTask(ctx context.Context, req *gen.CreateTaskRequest) (*gen.CreateTaskResponse, error) {
	result, err := s.deps.CreateTask(ctx, req.Url)
	if err != nil {
		return nil, err
	}
	return &gen.CreateTaskResponse{
		Succeed:       result.Succeed,
		AddedQuantity: int32(result.AddedQuantity),
		Msg:           result.Msg,
	}, nil
}

func (s *HostServiceServer) GetPluginRoot(ctx context.Context, req *gen.GetPluginRootRequest) (*gen.GetPluginRootResponse, error) {
	path := s.deps.GetPluginRoot(ctx, req.IsRelative)
	return &gen.GetPluginRootResponse{Path: path}, nil
}

func (s *HostServiceServer) Log(ctx context.Context, req *gen.LogRequest) (*gen.Empty, error) {
	if s.deps.LogFunc != nil {
		s.deps.LogFunc(req.Level, req.Template, req.Args, req.LoggerName)
	}
	return &gen.Empty{}, nil
}

func (s *HostServiceServer) PublishToFrontend(ctx context.Context, req *gen.PublishToFrontendRequest) (*gen.Empty, error) {
	if err := s.deps.PublishToFrontend(req.Topic, req.Data); err != nil {
		return nil, err
	}
	return &gen.Empty{}, nil
}

func (s *HostServiceServer) SubscribeFrontend(req *gen.SubscribeFrontendRequest, stream grpc.ServerStreamingServer[gen.FrontendMessage]) error {
	ch := make(chan []byte, 16)
	cancel, err := s.deps.SubscribeFrontend(req.Topic, func(data []byte) {
		ch <- data
	})
	if err != nil {
		return err
	}
	defer cancel()

	for {
		select {
		case data := <-ch:
			if err := stream.Send(&gen.FrontendMessage{Topic: req.Topic, Data: data}); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *HostServiceServer) UnsubscribeFrontend(ctx context.Context, req *gen.UnsubscribeFrontendRequest) (*gen.Empty, error) {
	return &gen.Empty{}, s.deps.UnsubscribeFrontend(req.Topic)
}

// ========== Provider 接口（定义在 dto 包中，HostDeps 在此引用）==========

// ========== GRPCPluginClient ==========

// GRPCPluginClient 封装插件侧的 gRPC 客户端接口
type GRPCPluginClient struct {
	Lifecycle     gen.PluginLifecycleClient
	Task          gen.TaskHandlerServiceClient
	Browser       gen.SiteBrowserServiceClient
	HostServiceId uint32 // GRPCBroker 上 HostService 的 ID，传递给插件 Activate
}

// DiscoverPluginServices 通过 gRPC 连接发现插件服务
func DiscoverPluginServices(conn *grpc.ClientConn) *GRPCPluginClient {
	return &GRPCPluginClient{
		Lifecycle: gen.NewPluginLifecycleClient(conn),
		Task:      gen.NewTaskHandlerServiceClient(conn),
		Browser:   gen.NewSiteBrowserServiceClient(conn),
	}
}

// GetGRPCConn 从 hashicorp/go-plugin client 获取 gRPC 连接
func GetGRPCConn(pluginClient *plugin.Client) (*grpc.ClientConn, error) {
	rpcClient, err := pluginClient.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin client: %w", err)
	}
	_, err = rpcClient.Dispense("library_squirrel")
	if err != nil {
		return nil, fmt.Errorf("failed to dispense plugin: %w", err)
	}
	return nil, fmt.Errorf("GetGRPCConn: use rpcClient.Conn() instead")
}

// ========== 转换函数（host 侧）==========

func workSetToProto(ws *dto.WorkSetDTO) *gen.WorkSet {
	if ws == nil {
		return nil
	}
	return &gen.WorkSet{
		Id:                     ws.ID,
		CreateTime:             ws.CreateTime,
		UpdateTime:             ws.UpdateTime,
		SiteId:                 ws.SiteID,
		SiteWorkSetId:          ws.SiteWorkSetID,
		SiteWorkSetName:        ws.SiteWorkSetName,
		SiteAuthorId:           ws.SiteAuthorID,
		SiteWorkSetDescription: ws.SiteWorkSetDescription,
		SiteUploadTime:         ws.SiteUploadTime,
		SiteUpdateTime:         ws.SiteUpdateTime,
		NickName:               ws.NickName,
		LastView:               ws.LastView,
	}
}

func protoToSite(pb *gen.Site) *dto.SiteDTO {
	if pb == nil {
		return nil
	}
	return &dto.SiteDTO{
		ID:              pb.Id,
		CreateTime:      pb.CreateTime,
		UpdateTime:      pb.UpdateTime,
		SiteName:        pb.SiteName,
		SiteDescription: pb.SiteDescription,
		Homepage:        pb.Homepage,
	}
}
