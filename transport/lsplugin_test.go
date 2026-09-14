package transport

import (
	"context"
	"net"
	"testing"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// serveGRPC 在环回监听上启动 gRPC server 并返回接通的客户端连接（测试结束自动清理）
func serveGRPC(t *testing.T, register func(*grpc.Server)) *grpc.ClientConn {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听环回端口失败: %v", err)
	}
	s := grpc.NewServer()
	register(s)
	go s.Serve(lis)
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.Stop()
		t.Fatalf("连接测试 server 失败: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
		s.Stop()
	})
	return conn
}

// pauseRecordingHandler 仅实现 Pause，嵌入 nil dto.TaskHandler 占位接口其余方法
type pauseRecordingHandler struct {
	dto.TaskHandler
	paused int
}

func (f *pauseRecordingHandler) Pause(param *dto.TaskResParam) error {
	f.paused++
	return nil
}

// TestGRPCServerWithoutTaskHandlerUnimplemented 无 TaskHandler 形态（工具型插件：
// 仅库查询/前端扩展等宿主能力）：TaskHandlerService 未注册，主程序侧任务 RPC
// 调用得 codes.Unimplemented（不 panic、不空响应）
func TestGRPCServerWithoutTaskHandlerUnimplemented(t *testing.T) {
	conn := serveGRPC(t, func(s *grpc.Server) {
		// broker 仅被 lifecycleServer 持有、Activate 调用时才解引用，注册路径传 nil 安全
		if err := (&LSPlugin{}).GRPCServer(nil, s); err != nil {
			t.Errorf("GRPCServer 注册失败: %v", err)
		}
	})
	taskClient := gen.NewTaskHandlerServiceClient(conn)
	unaryCalls := map[string]func() error{
		"Pause": func() error {
			_, err := taskClient.Pause(context.Background(), &gen.TaskResParamMessage{})
			return err
		},
		"CreateWorkInfo": func() error {
			_, err := taskClient.CreateWorkInfo(context.Background(), &gen.CreateWorkInfoRequest{})
			return err
		},
		"Retry": func() error {
			_, err := taskClient.Retry(context.Background(), &gen.RetryRequest{})
			return err
		},
	}
	for name, call := range unaryCalls {
		if err := call(); status.Code(err) != codes.Unimplemented {
			t.Errorf("%s 错误码 = %v, 期望 codes.Unimplemented (err=%v)", name, status.Code(err), err)
		}
	}
}

// TestGRPCServerWithTaskHandlerServesPause 对照：注册 TaskHandler 后同一 RPC 到达插件处理器
func TestGRPCServerWithTaskHandlerServesPause(t *testing.T) {
	handler := &pauseRecordingHandler{}
	conn := serveGRPC(t, func(s *grpc.Server) {
		if err := (&LSPlugin{Handler: handler}).GRPCServer(nil, s); err != nil {
			t.Errorf("GRPCServer 注册失败: %v", err)
		}
	})
	taskClient := gen.NewTaskHandlerServiceClient(conn)
	if _, err := taskClient.Pause(context.Background(), &gen.TaskResParamMessage{}); err != nil {
		t.Fatalf("Pause 应到达插件处理器: %v", err)
	}
	if handler.paused != 1 {
		t.Fatalf("处理器应被调到 1 次, 实得 %d", handler.paused)
	}
}

// queryProviderStub 仅实现被测方法，嵌入 nil dto.LibraryQueryProvider 占位接口其余方法
type queryProviderStub struct {
	dto.LibraryQueryProvider
	works      *gen.QueryWorksResponse
	workDirErr error
}

func (f *queryProviderStub) QueryWorks(ctx context.Context, req *gen.QueryWorksRequest) (*gen.QueryWorksResponse, error) {
	return f.works, nil
}

func (f *queryProviderStub) GetWorkDir(ctx context.Context, req *gen.Empty) (*gen.GetWorkDirResponse, error) {
	return nil, f.workDirErr
}

// TestPluginContextQueryReachesProvider 插件侧查询方法组全链：PluginContextClient →
// LibraryQuery service → HostDeps 注入的 LibraryQueryProvider；响应直通、status 错误原样透传
func TestPluginContextQueryReachesProvider(t *testing.T) {
	provider := &queryProviderStub{
		works:      &gen.QueryWorksResponse{Page: &gen.PageInfo{Total: 3}},
		workDirErr: status.Error(codes.FailedPrecondition, "工作目录未配置"),
	}
	conn := serveGRPC(t, func(s *grpc.Server) {
		RegisterHostService(s, HostDeps{LibraryQueryProvider: provider})
	})
	ctx := NewPluginContextClient(conn)

	resp, err := ctx.QueryWorks(&dto.QueryWorksRequest{})
	if err != nil {
		t.Fatalf("QueryWorks 失败: %v", err)
	}
	if resp.GetPage().GetTotal() != 3 {
		t.Fatalf("QueryWorks 响应总数 = %d, 期望 3（provider 返回值应直通）", resp.GetPage().GetTotal())
	}

	if _, err := ctx.GetWorkDir(); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("GetWorkDir 错误码 = %v, 期望 codes.FailedPrecondition (err=%v)", status.Code(err), err)
	}
}

// TestPluginContextQueryWithoutProviderUnimplemented 宿主未注入 LibraryQueryProvider：
// 查询调用得 codes.Unimplemented——新插件调旧宿主的优雅降级契约
func TestPluginContextQueryWithoutProviderUnimplemented(t *testing.T) {
	conn := serveGRPC(t, func(s *grpc.Server) {
		RegisterHostService(s, HostDeps{})
	})
	ctx := NewPluginContextClient(conn)
	if _, err := ctx.QueryWorks(&dto.QueryWorksRequest{}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("QueryWorks 错误码 = %v, 期望 codes.Unimplemented (err=%v)", status.Code(err), err)
	}
}
