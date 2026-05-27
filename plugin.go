package pluginsdk

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lvfeng-z/library-squirrel-plugin-sdk/gen"
)

// Handshake 握手配置，主程序和插件必须使用相同的值
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "LIBRARY_SQUIRREL_PLUGIN",
	MagicCookieValue: "d9a7f2b1e5c3846021de7baf0953c8e7",
}

// PluginMap 插件类型映射
var PluginMap = map[string]plugin.Plugin{
	"library_squirrel": &LSPlugin{},
}

// ServeOption 配置插件启动选项
type ServeOption func(*serveConfig)

type serveConfig struct {
	browser    SiteBrowser
	onActivate func(PluginContext)
}

// WithBrowser 注册 SiteBrowser 扩展点
func WithBrowser(browser SiteBrowser) ServeOption {
	return func(c *serveConfig) { c.browser = browser }
}

// WithActivate 设置 Activate 回调（在此回调中注册扩展点和 URL 监听器）
func WithActivate(fn func(PluginContext)) ServeOption {
	return func(c *serveConfig) { c.onActivate = fn }
}

// Serve 启动插件进程，由插件开发者调用
func Serve(handler TaskHandler, opts ...ServeOption) {
	cfg := &serveConfig{}
	for _, o := range opts {
		o(cfg)
	}

	lsPlugin := &LSPlugin{
		handler:    handler,
		browser:    cfg.browser,
		onActivate: cfg.onActivate,
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"library_squirrel": lsPlugin,
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// LSPlugin 实现 hashicorp/go-plugin 的 GRPCPlugin 接口
type LSPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	handler    TaskHandler
	browser    SiteBrowser
	onActivate func(PluginContext)

	// HostDeps 主程序侧注入的依赖，用于 GRPCClient 在主程序侧注册 HostService
	HostDeps *HostDeps
}

func (p *LSPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	gen.RegisterPluginLifecycleServer(s, &lifecycleServer{
		onActivate: p.onActivate,
		broker:     broker,
	})
	gen.RegisterTaskHandlerServiceServer(s, &taskHandlerServer{
		handler: p.handler,
	})
	if p.browser != nil {
		gen.RegisterSiteBrowserServiceServer(s, &siteBrowserServer{
			browser: p.browser,
		})
	}
	return nil
}

func (p *LSPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	// 主程序侧：在 GRPCBroker 上注册 HostService，供插件回调
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

// ========== PluginLifecycleServer ==========

type lifecycleServer struct {
	gen.UnimplementedPluginLifecycleServer
	onActivate func(PluginContext)
	broker     *plugin.GRPCBroker
}

func (s *lifecycleServer) Activate(ctx context.Context, req *gen.ActivateRequest) (*gen.ActivateResponse, error) {
	// 通过 GRPCBroker 拨号回主程序的 HostService
	if s.onActivate != nil && req.HostServiceId != 0 {
		conn, err := s.broker.Dial(req.HostServiceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "dial host service: %v", err)
		}
		pluginCtx := NewPluginContextClient(conn)
		pluginCtx.mainWindowHWND = uintptr(req.MainWindowHandle)
		s.onActivate(pluginCtx)
	}
	return &gen.ActivateResponse{}, nil
}

func (s *lifecycleServer) Shutdown(ctx context.Context, req *gen.Empty) (*gen.Empty, error) {
	return &gen.Empty{}, nil
}

// ========== TaskHandlerServiceServer ==========

type taskHandlerServer struct {
	gen.UnimplementedTaskHandlerServiceServer
	handler TaskHandler
}

func (s *taskHandlerServer) Create(req *gen.CreateRequest, stream grpc.ServerStreamingServer[gen.CreateChunk]) error {
	result, err := s.handler.Create(req.Url)
	if err != nil {
		return status.Errorf(codes.Internal, "create failed: %v", err)
	}

	// 发送模式标记
	if err := stream.Send(&gen.CreateChunk{
		Payload: &gen.CreateChunk_Mode{
			Mode: &gen.CreateMode{IsStream: result.IsStream()},
		},
	}); err != nil {
		return err
	}

	if result.IsStream() {
		// 流式模式：从 channel 逐条发送
		for resp := range result.Stream() {
			if err := stream.Send(&gen.CreateChunk{
				Payload: &gen.CreateChunk_Task{
					Task: taskCreateResponseToProto(resp),
				},
			}); err != nil {
				return err
			}
		}
	} else {
		// 批量模式：遍历数组逐条发送
		for _, resp := range result.Array() {
			if err := stream.Send(&gen.CreateChunk{
				Payload: &gen.CreateChunk_Task{
					Task: taskCreateResponseToProto(resp),
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *taskHandlerServer) CreateWorkInfo(ctx context.Context, req *gen.CreateWorkInfoRequest) (*gen.WorkResponse, error) {
	task := protoToTask(req.Task)
	workResp, err := s.handler.CreateWorkInfo(task)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "createWorkInfo failed: %v", err)
	}
	return workResponseToProto(workResp), nil
}

func (s *taskHandlerServer) Start(req *gen.StartRequest, stream gen.TaskHandlerService_StartServer) error {
	task := protoToTask(req.Task)
	reader, workResp, err := s.handler.Start(task)
	if err != nil {
		return status.Errorf(codes.Internal, "start failed: %v", err)
	}
	defer reader.Close()

	if err := stream.Send(&gen.StreamChunk{
		Payload: &gen.StreamChunk_WorkResponse{
			WorkResponse: workResponseToProto(workResp),
		},
	}); err != nil {
		return err
	}

	buf := make([]byte, 32*1024)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			if err := stream.Send(&gen.StreamChunk{
				Payload: &gen.StreamChunk_Data{Data: buf[:n]},
			}); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = stream.Send(&gen.StreamChunk{
				Payload: &gen.StreamChunk_Error{Error: readErr.Error()},
			})
			return status.Errorf(codes.Internal, "read error: %v", readErr)
		}
	}

	return stream.Send(&gen.StreamChunk{Payload: &gen.StreamChunk_Eof{Eof: true}})
}

func (s *taskHandlerServer) Retry(ctx context.Context, req *gen.RetryRequest) (*gen.WorkResponse, error) {
	task := protoToTask(req.Task)
	workResp, err := s.handler.Retry(task)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "retry failed: %v", err)
	}
	return workResponseToProto(workResp), nil
}

func (s *taskHandlerServer) Pause(ctx context.Context, req *gen.TaskResParamMessage) (*gen.Empty, error) {
	param := protoToTaskResParam(req.Param)
	if err := s.handler.Pause(param); err != nil {
		return nil, status.Errorf(codes.Internal, "pause failed: %v", err)
	}
	return &gen.Empty{}, nil
}

func (s *taskHandlerServer) Stop(ctx context.Context, req *gen.TaskResParamMessage) (*gen.Empty, error) {
	param := protoToTaskResParam(req.Param)
	if err := s.handler.Stop(param); err != nil {
		return nil, status.Errorf(codes.Internal, "stop failed: %v", err)
	}
	return &gen.Empty{}, nil
}

func (s *taskHandlerServer) Resume(req *gen.TaskResParamMessage, stream gen.TaskHandlerService_ResumeServer) error {
	param := protoToTaskResParam(req.Param)
	workResp, err := s.handler.Resume(param)
	if err != nil {
		return status.Errorf(codes.Internal, "resume failed: %v", err)
	}

	if workResp != nil {
		if err := stream.Send(&gen.StreamChunk{
			Payload: &gen.StreamChunk_WorkResponse{
				WorkResponse: workResponseToProto(workResp),
			},
		}); err != nil {
			return err
		}
	}

	return stream.Send(&gen.StreamChunk{Payload: &gen.StreamChunk_Eof{Eof: true}})
}

// ========== SiteBrowserServiceServer ==========

type siteBrowserServer struct {
	gen.UnimplementedSiteBrowserServiceServer
	browser SiteBrowser
}

func (s *siteBrowserServer) Open(ctx context.Context, req *gen.BrowserRequest) (*gen.Empty, error) {
	if err := s.browser.Open(); err != nil {
		return nil, status.Errorf(codes.Internal, "open browser failed: %v", err)
	}
	return &gen.Empty{}, nil
}

func (s *siteBrowserServer) Close(ctx context.Context, req *gen.BrowserRequest) (*gen.Empty, error) {
	if err := s.browser.Close(); err != nil {
		return nil, status.Errorf(codes.Internal, "close browser failed: %v", err)
	}
	return &gen.Empty{}, nil
}

// ========== 转换函数：SDK 类型 → Proto ==========

func taskCreateResponseToProto(r *TaskCreateResponse) *gen.TaskCreateResponse {
	pb := &gen.TaskCreateResponse{
		PluginTaskId: r.PluginTaskID,
		TaskName:     r.TaskName,
		SiteWorkId:   r.SiteWorkID,
		Url:          r.URL,
		PluginData:   r.PluginData,
		SiteName:     r.SiteName,
	}
	for _, c := range r.Children {
		pb.Children = append(pb.Children, &gen.TaskCreateChildResponse{
			TaskName:   c.TaskName,
			SiteWorkId: c.SiteWorkID,
			Url:        c.URL,
			PluginData: c.PluginData,
			SiteName:   c.SiteName,
		})
	}
	return pb
}

func workResponseToProto(r *WorkResponse) *gen.WorkResponse {
	if r == nil {
		return nil
	}
	pb := &gen.WorkResponse{}
	if r.Work != nil {
		pb.Work = workToProto(r.Work)
	}
	if r.Site != nil {
		pb.Site = siteDTOToProto(r.Site)
	}
	for _, a := range r.LocalAuthors {
		pb.LocalAuthors = append(pb.LocalAuthors, localAuthorDTOToProto(a))
	}
	for _, t := range r.LocalTags {
		pb.LocalTags = append(pb.LocalTags, localTagDTOToProto(t))
	}
	for _, a := range r.SiteAuthors {
		pb.SiteAuthors = append(pb.SiteAuthors, &gen.TaskSiteAuthorDTO{
			SiteAuthorId:    a.SiteAuthorID,
			AuthorName:      a.AuthorName,
			Homepage:        a.Homepage,
			FixedAuthorName: a.FixedAuthorName,
			Introduce:       a.Introduce,
		})
	}
	for _, t := range r.SiteTags {
		pb.SiteTags = append(pb.SiteTags, &gen.TaskSiteTagDTO{
			SiteTagId:   t.SiteTagID,
			TagName:     t.TagName,
			Description: t.Description,
		})
	}
	for _, ws := range r.WorkSets {
		pb.WorkSets = append(pb.WorkSets, &gen.TaskWorkSetDTO{
			SiteWorkSetId: ws.SiteWorkSetID,
			WorkSetName:   ws.WorkSetName,
		})
	}
	if r.Resource != nil {
		pb.Resource = &gen.TaskResourceDTO{
			ResourceId:   r.Resource.ResourceID,
			Url:          r.Resource.URL,
			Type:         r.Resource.Type,
			Format:       r.Resource.Format,
			LocalPath:    r.Resource.LocalPath,
			RemotePath:   r.Resource.RemotePath,
			Size:         r.Resource.Size,
			Completeness: int32(r.Resource.Completeness),
		}
	}
	return pb
}

func workToProto(w *Work) *gen.Work {
	if w == nil {
		return nil
	}
	return &gen.Work{
		Id:                  w.ID,
		CreateTime:          w.CreateTime,
		UpdateTime:          w.UpdateTime,
		SiteId:              w.SiteID,
		SiteWorkId:          w.SiteWorkID,
		SiteWorkName:        w.SiteWorkName,
		SiteAuthorId:        w.SiteAuthorID,
		SiteWorkDescription: w.SiteWorkDescription,
		SiteUploadTime:      w.SiteUploadTime,
		SiteUpdateTime:      w.SiteUpdateTime,
		NickName:            w.NickName,
		LocalAuthorId:       w.LocalAuthorID,
		LastView:            w.LastView,
	}
}

func siteDTOToProto(s *SiteDTO) *gen.SiteDTO {
	if s == nil {
		return nil
	}
	return &gen.SiteDTO{
		Id:              s.ID,
		SiteName:        s.SiteName,
		SiteDescription: s.SiteDescription,
		Homepage:        s.Homepage,
		CreateTime:      s.CreateTime,
		UpdateTime:      s.UpdateTime,
	}
}

func localAuthorDTOToProto(a *LocalAuthorDTO) *gen.LocalAuthorDTO {
	if a == nil {
		return nil
	}
	return &gen.LocalAuthorDTO{
		Id:         a.ID,
		AuthorName: a.AuthorName,
		Introduce:  a.Introduce,
		LastUse:    a.LastUse,
		CreateTime: a.CreateTime,
		UpdateTime: a.UpdateTime,
	}
}

func localTagDTOToProto(t *LocalTagDTO) *gen.LocalTagDTO {
	if t == nil {
		return nil
	}
	return &gen.LocalTagDTO{
		Id:             t.ID,
		LocalTagName:   t.LocalTagName,
		BaseLocalTagId: t.BaseLocalTagID,
		Description:    t.Description,
		LastUse:        t.LastUse,
		CreateTime:     t.CreateTime,
		UpdateTime:     t.UpdateTime,
	}
}

// ========== 转换函数：Proto → SDK 类型 ==========

func protoToTask(pb *gen.Task) *Task {
	if pb == nil {
		return nil
	}
	return &Task{
		ID:                   pb.Id,
		CreateTime:           pb.CreateTime,
		UpdateTime:           pb.UpdateTime,
		HasChild:             pb.HasChild,
		Pid:                  pb.Pid,
		TaskName:             pb.TaskName,
		SiteID:               pb.SiteId,
		SiteWorkID:           pb.SiteWorkId,
		URL:                  pb.Url,
		Status:               int(pb.Status),
		PendingResourceID:    pb.PendingResourceId,
		Continuable:          pb.Continuable,
		PluginPublicID:       pb.PluginPublicId,
		PluginContributionID: pb.PluginContributionId,
		PluginData:           pb.PluginData,
		ErrorMessage:         pb.ErrorMessage,
	}
}

func protoToTaskResParam(pb *gen.TaskResParam) *TaskResParam {
	if pb == nil {
		return nil
	}
	return &TaskResParam{
		Task:         protoToTask(pb.Task),
		ResourceID:   pb.ResourceId,
		ResourcePath: pb.ResourcePath,
	}
}

// FormatLogArgs 格式化日志参数（供 HostService 日志使用）
func FormatLogArgs(template string, args ...any) string {
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}
