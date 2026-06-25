package transport

import (
	"context"
	"io"

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
	reader, workResp, err := s.handler.Resume(param)
	if err != nil {
		return status.Errorf(codes.Internal, "resume failed: %v", err)
	}
	if reader != nil {
		defer reader.Close()
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

	if reader != nil {
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
				return status.Errorf(codes.Internal, "resume read error: %v", readErr)
			}
		}
	}

	return stream.Send(&gen.StreamChunk{Payload: &gen.StreamChunk_Eof{Eof: true}})
}

func (s *taskHandlerServer) GetThumbnail(ctx context.Context, req *gen.GetThumbnailRequest) (*gen.GetThumbnailResponse, error) {
	if s.handler == nil {
		return nil, status.Error(codes.Unimplemented, "handler not registered")
	}
	resp, err := s.handler.GetThumbnail(req.TaskData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getThumbnail failed: %v", err)
	}
	if resp == nil {
		return &gen.GetThumbnailResponse{}, nil
	}
	return &gen.GetThumbnailResponse{
		Data:   resp.Data,
		Format: resp.Format,
	}, nil
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
	if r.Resource != nil {
		pb.Resource = &gen.TaskResourceDTO{
			Format:      r.Resource.Format,
			Size:        r.Resource.Size,
			SuggestName: r.Resource.SuggestName,
			Continuable: r.Resource.Continuable,
		}
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
		PluginExtensionID: pb.PluginExtensionId,
		PluginData:           pb.PluginData,
		ErrorMessage:         pb.ErrorMessage,
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
