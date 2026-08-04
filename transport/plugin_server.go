package transport

import (
	"context"
	"fmt"
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
					Task: resp,
				},
			}); err != nil {
				return err
			}
		}
	} else {
		for _, resp := range result.Array() {
			if err := stream.Send(&gen.CreateChunk{
				Payload: &gen.CreateChunk_Task{
					Task: resp,
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *taskHandlerServer) CreateWorkInfo(ctx context.Context, req *gen.CreateWorkInfoRequest) (*gen.WorkResponse, error) {
	task := req.Task
	workResp, err := s.handler.CreateWorkInfo(task)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "createWorkInfo failed: %v", err)
	}
	return workResp, nil
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
	task := startReq.Task
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
	task := req.Task
	workResp, err := s.handler.Retry(task)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "retry failed: %v", err)
	}
	return workResp, nil
}

func (s *taskHandlerServer) Pause(ctx context.Context, req *gen.TaskResParamMessage) (*gen.Empty, error) {
	param := req.Param
	if err := s.handler.Pause(param); err != nil {
		return nil, status.Errorf(codes.Internal, "pause failed: %v", err)
	}
	return &gen.Empty{}, nil
}

func (s *taskHandlerServer) Stop(ctx context.Context, req *gen.TaskResParamMessage) (*gen.Empty, error) {
	param := req.Param
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

// QueryWorkSetOrder 查询作品集内作品原站顺序（主程序作品入库后拉取，仅写 site_sort_order）
// 可选能力：插件未实现 dto.WorkOrderQuerier 时返回空响应（site_sort_order 保持空，仅本地序）
func (s *taskHandlerServer) QueryWorkSetOrder(ctx context.Context, req *gen.QueryWorkSetOrderRequest) (*gen.QueryWorkSetOrderResponse, error) {
	querier, ok := s.handler.(dto.WorkOrderQuerier)
	if !ok {
		return &gen.QueryWorkSetOrderResponse{}, nil
	}
	entries, err := querier.QueryWorkSetOrder(req.GetSiteId(), req.GetSiteWorkSetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "queryWorkSetOrder failed: %v", err)
	}
	return &gen.QueryWorkSetOrderResponse{Entries: entries}, nil
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
		if err := send(&gen.StreamChunk{Payload: &gen.StreamChunk_WorkResponse{WorkResponse: workResp}}); err != nil {
			return err
		}
	}
	if err := send(&gen.StreamChunk{Payload: &gen.StreamChunk_Specs{Specs: storeSpecsToProto(specs)}}); err != nil {
		return err
	}

	// readers 按 spec 顺序(specIndex 索引);同 role 多 spec 各自独立 reader,替代旧 map[role] 同 role 覆盖
	readers := make([]io.ReadCloser, len(specs))
	for i, sp := range specs {
		if sp != nil && sp.ReadCloser != nil {
			readers[i] = sp.ReadCloser
		}
	}
	completed := make([]bool, len(readers))
	completedCount := 0
	totalReaders := 0
	for _, r := range readers {
		if r != nil {
			totalReaders++
		}
	}
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
		// role 字段编码 "role#specIndex"(主程序 recvSpecsAndPull 编码),解析回纯 role + spec 索引
		encodedRole := pull.GetRole()
		role, specIdx := DecodePullRole(encodedRole)
		if specIdx < 0 || specIdx >= len(readers) || readers[specIdx] == nil {
			if e := send(&gen.StreamChunk{Role: encodedRole, Payload: &gen.StreamChunk_Error{Error: fmt.Sprintf("未知 spec 索引: %d (role=%s)", specIdx, role)}}); e != nil {
				return e
			}
			continue
		}
		if completed[specIdx] {
			if e := send(&gen.StreamChunk{Role: encodedRole, Payload: &gen.StreamChunk_Eof{Eof: true}}); e != nil {
				return e
			}
			continue
		}
		reader := readers[specIdx]
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
				if e := send(&gen.StreamChunk{Role: encodedRole, Payload: &gen.StreamChunk_Data{Data: append([]byte(nil), buf[:n]...)}}); e != nil {
					return e
				}
			}
			if readErr == io.EOF {
				if e := send(&gen.StreamChunk{Role: encodedRole, Payload: &gen.StreamChunk_Eof{Eof: true}}); e != nil {
					return e
				}
				completed[specIdx] = true
				completedCount++
				if completedCount == totalReaders {
					return nil // 全部 reader EOF
				}
			} else if readErr != nil {
				if e := send(&gen.StreamChunk{Role: encodedRole, Payload: &gen.StreamChunk_Error{Error: readErr.Error()}}); e != nil {
					return e
				}
				completed[specIdx] = true
				completedCount++
				if completedCount == totalReaders {
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

// ========== 转换函数 ==========
// 别名类型(dto=gen,如 TaskDTO/WorkDTO/WorkResponse/TaskCreateResponse 等)在 gRPC 边界直传,无需转换。
// 仅 TaskResumeParam(含 OffsetForRole 方法)/StoreResumeOffset(含 String 方法)保留为 struct,
// 及 StoreSpec(手写 io.ReadCloser)需手工 proto↔DTO 转换。

// protoToTaskResumeParam 续传参数 proto → DTO(TaskResumeParam 非别名:含 OffsetForRole 方法)
func protoToTaskResumeParam(pb *gen.TaskResumeParam) *dto.TaskResumeParam {
	if pb == nil {
		return nil
	}
	offsets := make([]*dto.StoreResumeOffset, 0, len(pb.StreamOffsets))
	for _, o := range pb.StreamOffsets {
		offsets = append(offsets, protoToStoreResumeOffset(o))
	}
	return &dto.TaskResumeParam{
		Task:          pb.Task,
		StreamOffsets: offsets,
	}
}

// protoToStoreResumeOffset 把续传偏移从 proto 转为 DTO(身份化:role+store_seq)
func protoToStoreResumeOffset(pb *gen.StoreResumeOffset) *dto.StoreResumeOffset {
	if pb == nil {
		return nil
	}
	return &dto.StoreResumeOffset{
		Role:     pb.Role,
		StoreSeq: pb.StoreSeq,
		Offset:   pb.Offset,
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
			Description: sp.Description,
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
