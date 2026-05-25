package pluginsdk

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/lvfeng-z/library-squirrel-plugin-sdk/gen"
)

// HostDeps 主程序侧提供给 HostService 的依赖
type HostDeps struct {
	PluginDataProvider
	SecureStorageProvider
	WorkSetQueryProvider
	SiteSaveProvider
	TaskCreateProvider
	UrlListenerRegistry
	FrontendEventProvider
	LogFunc func(level int32, template string, args []string, loggerName string)
	// 注册回调：插件通过 HostService 注册扩展点时调用
	OnRegisterTaskHandler  func(contributionId, name, description string) error
	OnRegisterSiteBrowser  func(contributionId, name, description string) error
	OnUnregisterSiteBrowser func(contributionId string) error
}

// NewClientConfig 创建 hashicorp/go-plugin 的 Client 配置
func NewClientConfig(cmd *exec.Cmd) *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		Plugins:          PluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	}
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

func (s *HostServiceServer) GetPluginData(ctx context.Context, req *gen.Empty) (*gen.PluginDataResponse, error) {
	data, err := s.deps.GetPluginData(ctx)
	if err != nil {
		return nil, err
	}
	return &gen.PluginDataResponse{Data: data}, nil
}

func (s *HostServiceServer) SetPluginData(ctx context.Context, req *gen.PluginDataRequest) (*gen.Empty, error) {
	return &gen.Empty{}, s.deps.SetPluginData(ctx, req.Data)
}

func (s *HostServiceServer) StoreEncryptedValue(ctx context.Context, req *gen.EncryptRequest) (*gen.EncryptResponse, error) {
	key, err := s.deps.StoreEncryptedValue(ctx, req.PlainValue, req.Description)
	if err != nil {
		return nil, err
	}
	return &gen.EncryptResponse{Key: key}, nil
}

func (s *HostServiceServer) GetDecryptedValue(ctx context.Context, req *gen.DecryptRequest) (*gen.DecryptResponse, error) {
	value, err := s.deps.GetDecryptedValue(ctx, req.StorageKey)
	if err != nil {
		return nil, err
	}
	return &gen.DecryptResponse{Value: value}, nil
}

func (s *HostServiceServer) RemoveEncryptedValue(ctx context.Context, req *gen.DecryptRequest) (*gen.Empty, error) {
	return &gen.Empty{}, s.deps.RemoveEncryptedValue(ctx, req.StorageKey)
}

func (s *HostServiceServer) GetWorkSetBySiteWorkSetId(ctx context.Context, req *gen.WorkSetQueryRequest) (*gen.WorkSetQueryResponse, error) {
	ws, err := s.deps.GetWorkSetBySiteWorkSetId(ctx, req.SiteWorkSetId, req.SiteName)
	if err != nil {
		return nil, err
	}
	return &gen.WorkSetQueryResponse{WorkSet: workSetToProto(ws)}, nil
}

func (s *HostServiceServer) AddSite(ctx context.Context, req *gen.AddSiteRequest) (*gen.Empty, error) {
	sites := make([]*Site, len(req.Sites))
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

// ========== Provider 接口 ==========

// PluginDataProvider 插件数据持久化
type PluginDataProvider interface {
	GetPluginData(ctx context.Context) (string, error)
	SetPluginData(ctx context.Context, data string) error
	GetPluginRoot(ctx context.Context, isRelative bool) string
}

// SecureStorageProvider 加密存储
type SecureStorageProvider interface {
	StoreEncryptedValue(ctx context.Context, plainValue, description string) (string, error)
	GetDecryptedValue(ctx context.Context, storageKey string) (string, error)
	RemoveEncryptedValue(ctx context.Context, storageKey string) error
}

// WorkSetQueryProvider 作品集查询
type WorkSetQueryProvider interface {
	GetWorkSetBySiteWorkSetId(ctx context.Context, siteWorkSetId, siteName string) (*WorkSet, error)
}

// SiteSaveProvider 站点保存
type SiteSaveProvider interface {
	AddSite(ctx context.Context, sites []*Site) error
}

// TaskCreateProvider 任务创建
type TaskCreateProvider interface {
	CreateTask(ctx context.Context, url string) (*CreateTaskResult, error)
}

// UrlListenerRegistry URL 监听器注册
type UrlListenerRegistry interface {
	RegisterUrlListener(ctx context.Context, contributionId string, patterns []string) error
	UnregisterUrlListener(ctx context.Context) error
}

// FrontendEventProvider 前后端事件桥接
type FrontendEventProvider interface {
	PublishToFrontend(topic string, data []byte) error
	SubscribeFrontend(topic string, pushCh func([]byte)) (cancel func(), err error)
	UnsubscribeFrontend(topic string) error
}

// ========== Site/WorkSet 转换（HostService 专用）==========

func workSetToProto(ws *WorkSet) *gen.WorkSet {
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

func protoToSite(pb *gen.Site) *Site {
	if pb == nil {
		return nil
	}
	return &Site{
		ID:              pb.Id,
		CreateTime:      pb.CreateTime,
		UpdateTime:      pb.UpdateTime,
		SiteName:        pb.SiteName,
		SiteDescription: pb.SiteDescription,
		Homepage:        pb.Homepage,
	}
}

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
