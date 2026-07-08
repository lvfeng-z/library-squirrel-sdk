package transport

import (
	"context"
	"io"
	"log"

	"github.com/hashicorp/go-plugin"
	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ========== PluginLifecycleServer ==========

type lifecycleServer struct {
	gen.UnimplementedPluginLifecycleServer
	onActivate func(pluginCtx dto.PluginContext)
	onShutdown func()
	broker     *plugin.GRPCBroker
}

func (s *lifecycleServer) Activate(ctx context.Context, req *gen.ActivateRequest) (*gen.ActivateResponse, error) {
	if s.onActivate != nil && req.HostServiceId != 0 {
		conn, err := s.broker.Dial(req.HostServiceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "dial host service: %v", err)
		}
		pluginCtx := NewPluginContextClient(conn)
		pluginCtx.SetMainWindowHandle(uintptr(req.MainWindowHandle))
		s.onActivate(pluginCtx)
	}
	return &gen.ActivateResponse{}, nil
}

func (s *lifecycleServer) Shutdown(ctx context.Context, req *gen.Empty) (*gen.Empty, error) {
	if s.onShutdown != nil {
		s.onShutdown()
	}
	return &gen.Empty{}, nil
}

// ========== TaskHandlerServiceServer ==========

type taskHandlerServer struct {
	gen.UnimplementedTaskHandlerServiceServer
	handler dto.TaskHandler
}

func (s *taskHandlerServer) Create(req *gen.CreateRequest, stream grpc.ServerStreamingServer[gen.CreateChunk]) error {
	result, err := s.handler.Create(req.Url)
	if err != nil {
		return status.Errorf(codes.Internal, "create failed: %v", err)
	}

	if err := stream.Send(&gen.CreateChunk{
		Payload: &gen.CreateChunk_Mode{
			Mode: &gen.CreateMode{IsStream: result.IsStream()},
		},
	}); err != nil {
		return err
	}

	if result.IsStream() {
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

func (s *taskHandlerServer) Start(stream gen.TaskHandlerService_StartServer) error {
	ctx := stream.Context()
	// 首帧:StartRequest
	frame, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "start: 收首帧失败: %v", err)
	}
	startReq := frame.GetStart()
	if startReq == nil {
		return status.Errorf(codes.InvalidArgument, "start: 首帧必须为 StartRequest")
	}
	task := protoToTask(startReq.Task)
	specs, workResp, err := s.handler.Start(ctx, task, startReq.StoreRoles)
	if err != nil {
		return status.Errorf(codes.Internal, "start failed: %v", err)
	}
	return serveSpecsPull(ctx, stream.Send, func() (*gen.PullRequest, error) {
		f, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		return f.GetPull(), nil
	}, specs, workResp)
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

func (s *taskHandlerServer) Resume(stream gen.TaskHandlerService_ResumeServer) error {
	ctx := stream.Context()
	// 首帧:TaskResumeParamMessage
	frame, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "resume: 收首帧失败: %v", err)
	}
	resumeReq := frame.GetResume()
	if resumeReq == nil {
		return status.Errorf(codes.InvalidArgument, "resume: 首帧必须为 TaskResumeParamMessage")
	}
	param := protoToTaskResumeParam(resumeReq.Param)
	specs, workResp, err := s.handler.Resume(ctx, param)
	if err != nil {
		return status.Errorf(codes.Internal, "resume failed: %v", err)
	}
	return serveSpecsPull(ctx, stream.Send, func() (*gen.PullRequest, error) {
		f, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		return f.GetPull(), nil
	}, specs, workResp)
}

// closeSpecReaders 关闭所有 spec 的 reader(忽略 nil)
func closeSpecReaders(specs []*dto.StoreSpec) {
	for _, sp := range specs {
		if sp != nil && sp.ReadCloser != nil {
			_ = sp.ReadCloser.Close()
		}
	}
}

// pullReadResult 单次 reader.Read 的结果,供 goroutine + select 把阻塞读的结果传回主循环
type pullReadResult struct {
	n   int
	err error
}

// serveSpecsPull 发送 WorkResponse(可选)+ Specs 声明,随后进入 pull 循环:
// Recv(PullRequest) → 按 role 选 reader → reader.Read(max_bytes) 一批 → Send(data/eof/error)。
// reader.Read 由主程序按需驱动,reader 不领先主程序落盘(主程序保证持久化的前提)。
// ctx 为 gRPC stream context:主程序取消任务时 ctx Done,若 reader.Read 正阻塞在网络,
// 通过 Close reader 令其返回(合规 reader 的 Close 可中断在途 Read),使 serveSpecsPull 及时退出。
func serveSpecsPull(
	ctx context.Context,
	send func(*gen.StreamChunk) error,
	recvPull func() (*gen.PullRequest, error),
	specs []*dto.StoreSpec, workResp *dto.WorkResponse,
) error {
	defer closeSpecReaders(specs)
	if workResp != nil {
		if err := send(&gen.StreamChunk{Payload: &gen.StreamChunk_WorkResponse{WorkResponse: workResponseToProto(workResp)}}); err != nil {
			return err
		}
	}
	if err := send(&gen.StreamChunk{Payload: &gen.StreamChunk_Specs{Specs: storeSpecsToProto(specs)}}); err != nil {
		return err
	}

	readers := make(map[string]io.ReadCloser, len(specs))
	for _, sp := range specs {
		if sp != nil && sp.ReadCloser != nil {
			readers[sp.Role] = sp.ReadCloser
		}
	}
	completed := make(map[string]struct{}, len(readers))
	buf := make([]byte, 32*1024)
	for {
		pull, err := recvPull()
		if err == io.EOF {
			return nil // 主程序关闭发送侧
		}
		if err != nil {
			return err
		}
		if pull == nil {
			continue
		}
		role := pull.GetRole()
		reader, ok := readers[role]
		if !ok {
			if e := send(&gen.StreamChunk{Role: role, Payload: &gen.StreamChunk_Error{Error: "未知 role: " + role}}); e != nil {
				return e
			}
			continue
		}
		if _, done := completed[role]; done {
			if e := send(&gen.StreamChunk{Role: role, Payload: &gen.StreamChunk_Eof{Eof: true}}); e != nil {
				return e
			}
			continue
		}
		maxN := int(pull.GetMaxBytes())
		if maxN <= 0 || maxN > len(buf) {
			maxN = len(buf)
		}
		// reader.Read 可能阻塞在网络、不响应 ctx;用 goroutine + select 让 ctx 取消可中断:
		// ctx Done 时 Close reader,合规 reader 的 Close 会让阻塞的 Read 返回,
		// goroutine 随后写入 buffered channel 退出,无泄漏。
		ch := make(chan pullReadResult, 1)
		go func() {
			n, err := reader.Read(buf[:maxN])
			ch <- pullReadResult{n, err}
		}()
		select {
		case <-ctx.Done():
			// ctx 取消时 reader.Read 可能恰好完成,chunk 被丢弃(reader 内部 offset 已推进而主程序未落盘)。
			// 仅记录丢包窗口命中(n>0)作长期哨兵;常态(Read 阻塞或 n=0)静默。
			select {
			case res := <-ch:
				if res.n > 0 {
					log.Printf("[serveSpecsPull] 丢包窗口命中: ctx 取消时 reader.Read 已完成, 丢弃 %d 字节 role=%s err=%v", res.n, role, res.err)
				}
			default:
			}
			reader.Close()
			return ctx.Err()
		case res := <-ch:
			n, readErr := res.n, res.err
			if n > 0 {
				if e := send(&gen.StreamChunk{Role: role, Payload: &gen.StreamChunk_Data{Data: append([]byte(nil), buf[:n]...)}}); e != nil {
					return e
				}
			}
			if readErr == io.EOF {
				if e := send(&gen.StreamChunk{Role: role, Payload: &gen.StreamChunk_Eof{Eof: true}}); e != nil {
					return e
				}
				completed[role] = struct{}{}
				if len(completed) == len(readers) {
					return nil // 全部 role EOF
				}
			} else if readErr != nil {
				if e := send(&gen.StreamChunk{Role: role, Payload: &gen.StreamChunk_Error{Error: readErr.Error()}}); e != nil {
					return e
				}
				completed[role] = struct{}{}
				if len(completed) == len(readers) {
					return nil
				}
			}
		}
	}
}

// ========== SiteBrowserServiceServer ==========

type siteBrowserServer struct {
	gen.UnimplementedSiteBrowserServiceServer
	browser dto.SiteBrowser
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

// ========== 转换函数：DTO → Proto ==========

func taskCreateResponseToProto(r *dto.TaskCreateResponse) *gen.TaskCreateResponse {
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

func workResponseToProto(r *dto.WorkResponse) *gen.WorkResponse {
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
	return pb
}

func workToProto(w *dto.WorkDTO) *gen.Work {
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

func siteDTOToProto(s *dto.SiteDTO) *gen.SiteDTO {
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

func localAuthorDTOToProto(a *dto.LocalAuthorDTO) *gen.LocalAuthorDTO {
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

func localTagDTOToProto(t *dto.LocalTagDTO) *gen.LocalTagDTO {
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

// ========== 转换函数：Proto → DTO ==========

func protoToTask(pb *gen.Task) *dto.TaskDTO {
	if pb == nil {
		return nil
	}
	return &dto.TaskDTO{
		ID:                pb.Id,
		CreateTime:        pb.CreateTime,
		UpdateTime:        pb.UpdateTime,
		HasChild:          pb.HasChild,
		Pid:               pb.Pid,
		TaskName:          pb.TaskName,
		SiteID:            pb.SiteId,
		SiteWorkID:        pb.SiteWorkId,
		URL:               pb.Url,
		Status:            int(pb.Status),
		PendingResourceID: pb.PendingResourceId,
		Continuable:       pb.Continuable,
		PluginPublicID:    pb.PluginPublicId,
		PluginExtensionID: pb.PluginExtensionId,
		PluginData:        pb.PluginData,
		ErrorMessage:      pb.ErrorMessage,
	}
}

func protoToTaskResParam(pb *gen.TaskResParam) *dto.TaskResParam {
	if pb == nil {
		return nil
	}
	return &dto.TaskResParam{
		Task:            protoToTask(pb.Task),
		ResourceID:      pb.ResourceId,
		ResourcePath:    pb.ResourcePath,
		DownloadedBytes: pb.DownloadedBytes,
	}
}

func protoToTaskResumeParam(pb *gen.TaskResumeParam) *dto.TaskResumeParam {
	if pb == nil {
		return nil
	}
	return &dto.TaskResumeParam{
		Task:          protoToTask(pb.Task),
		StreamOffsets: pb.StreamOffsets,
	}
}

func storeSpecsToProto(specs []*dto.StoreSpec) *gen.StoreSpecs {
	pb := &gen.StoreSpecs{}
	for _, sp := range specs {
		if sp == nil {
			continue
		}
		meta := &gen.StoreSpecMeta{
			Role:        sp.Role,
			Generation:  sp.Generation,
			Format:      sp.Format,
			Size:        sp.Size,
			SuggestName: sp.SuggestName,
		}
		if sp.Continuable != nil {
			meta.Continuable = sp.Continuable
		}
		if sp.ResumeWriteOffset != nil {
			meta.ResumeWriteOffset = sp.ResumeWriteOffset
		}
		pb.Items = append(pb.Items, meta)
	}
	return pb
}
