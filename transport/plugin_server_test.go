package transport

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"google.golang.org/grpc"
)

// fakeCreateStream 收集 Create 流发出的全部块。Create 只调用 Send，
// 嵌入的 grpc.ServerStream 仅供满足接口，不会被触达。
type fakeCreateStream struct {
	grpc.ServerStream
	chunks []*gen.CreateChunk
}

func (f *fakeCreateStream) Send(chunk *gen.CreateChunk) error {
	f.chunks = append(f.chunks, chunk)
	return nil
}

// fakeCreateHandler 仅实现 Create，嵌入 nil dto.WorkFetcher 占位接口其余方法
type fakeCreateHandler struct {
	dto.WorkFetcher
	result *dto.TaskCreateResult
	err    error
}

func (f *fakeCreateHandler) Create(url string) (*dto.TaskCreateResult, error) {
	return f.result, f.err
}

// runCreate 驱动 workFetchServer.Create 并断言流正常结束（无 gRPC 错误）
func runCreate(t *testing.T, handler dto.WorkFetcher) []*gen.CreateChunk {
	t.Helper()
	stream := &fakeCreateStream{}
	server := &workFetchServer{handler: handler}
	if err := server.Create(&gen.CreateRequest{Url: "https://example.com/work"}, stream); err != nil {
		t.Fatalf("Create 返回 gRPC 错误: %v", err)
	}
	return stream.chunks
}

// TestCreateHandlerErrorSingleErrorChunk 插件 Create 错误返回形态：
// 仅发单个 error 块承载原因文本，无 mode 块，流以正常结束收尾（非 gRPC 错误）
func TestCreateHandlerErrorSingleErrorChunk(t *testing.T) {
	chunks := runCreate(t, &fakeCreateHandler{err: errors.New("获取作品详情失败")})
	if len(chunks) != 1 {
		t.Fatalf("期望恰 1 块, 实得 %d 块", len(chunks))
	}
	if e := chunks[0].GetError(); e != "获取作品详情失败" {
		t.Fatalf("error 块文本 = %q, 期望 %q", e, "获取作品详情失败")
	}
	if chunks[0].GetMode() != nil {
		t.Fatal("错误返回形态不应发 mode 块")
	}
}

// TestCreateBatchReasonAppendErrorChunk 批量 + SetReason 形态：
// 块序为 mode → task… → error，error 块必为最后一块且文本即 reason
func TestCreateBatchReasonAppendErrorChunk(t *testing.T) {
	result := dto.BatchResult([]*dto.TaskCreateResponse{{}, {}})
	result.SetReason("未发现可导入的文件")
	chunks := runCreate(t, &fakeCreateHandler{result: result})
	if len(chunks) != 4 {
		t.Fatalf("期望 4 块(mode+2 task+error), 实得 %d 块", len(chunks))
	}
	if mode := chunks[0].GetMode(); mode == nil || mode.GetIsStream() {
		t.Fatal("首块应为批量 mode 块(is_stream=false)")
	}
	for i := 1; i <= 2; i++ {
		if chunks[i].GetTask() == nil {
			t.Fatalf("第 %d 块应为 task 块", i)
		}
	}
	if e := chunks[3].GetError(); e != "未发现可导入的文件" {
		t.Fatalf("末块 error 文本 = %q, 期望 %q", e, "未发现可导入的文件")
	}
}

// TestCreateStreamReasonAppendErrorChunk 流式 + SetReason 形态：
// 生产侧在 close channel 前声明 reason，服务端在流消费完毕后读取，
// 跨 goroutine 依赖 channel close 的 happens-before 顺序；error 块为末块
func TestCreateStreamReasonAppendErrorChunk(t *testing.T) {
	producer := make(chan *dto.TaskCreateResponse)
	result := dto.StreamResult(producer)
	go func() {
		producer <- &dto.TaskCreateResponse{}
		result.SetReason("部分任务创建失败")
		close(producer)
	}()
	chunks := runCreate(t, &fakeCreateHandler{result: result})
	if len(chunks) != 3 {
		t.Fatalf("期望 3 块(mode+task+error), 实得 %d 块", len(chunks))
	}
	if mode := chunks[0].GetMode(); mode == nil || !mode.GetIsStream() {
		t.Fatal("首块应为流式 mode 块(is_stream=true)")
	}
	if chunks[1].GetTask() == nil {
		t.Fatal("第 2 块应为 task 块")
	}
	if e := chunks[2].GetError(); e != "部分任务创建失败" {
		t.Fatalf("末块 error 文本 = %q, 期望 %q", e, "部分任务创建失败")
	}
}

// TestCreateWithoutReasonNoErrorChunk 未声明 reason 时不追加 error 块
func TestCreateWithoutReasonNoErrorChunk(t *testing.T) {
	result := dto.BatchResult([]*dto.TaskCreateResponse{{}})
	chunks := runCreate(t, &fakeCreateHandler{result: result})
	if len(chunks) != 2 {
		t.Fatalf("期望 2 块(mode+task), 实得 %d 块", len(chunks))
	}
	for i, c := range chunks {
		if c.GetError() != "" {
			t.Fatalf("第 %d 块不应为 error 块", i)
		}
	}
}

// ========== Create 流等待期心跳接入面 ==========

// fakeHeartbeatCreateHandler 实现 CreateWithHeartbeat 的替身：等待点行为经 onCreate
// 回调注入；createCnt 计数用于断言双路径分派（实现新接口时原 Create 不被调用）
type fakeHeartbeatCreateHandler struct {
	dto.WorkFetcher
	result    *dto.TaskCreateResult
	err       error
	createCnt int
	onCreate  func(reporter dto.HeartbeatReporter)
}

func (f *fakeHeartbeatCreateHandler) Create(url string) (*dto.TaskCreateResult, error) {
	f.createCnt++
	return f.result, f.err
}

func (f *fakeHeartbeatCreateHandler) CreateWithHeartbeat(url string, reporter dto.HeartbeatReporter) (*dto.TaskCreateResult, error) {
	if f.onCreate != nil {
		f.onCreate(reporter)
	}
	return f.result, f.err
}

// TestCreateWithHeartbeatSendsHeartbeatBeforeResult 实现 CreateWithHeartbeat 的
// handler：等待期上报的心跳块先于 mode/task 上线，原 Create 方法不被调用
func TestCreateWithHeartbeatSendsHeartbeatBeforeResult(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	h := &fakeHeartbeatCreateHandler{result: dto.BatchResult([]*dto.TaskCreateResponse{{}})}
	h.onCreate = func(reporter dto.HeartbeatReporter) { reporter.Heartbeat() }
	chunks := runCreate(t, h)
	if len(chunks) != 3 {
		t.Fatalf("期望 3 块(heartbeat+mode+task), 实得 %d 块", len(chunks))
	}
	if chunks[0].GetHeartbeat() == nil {
		t.Fatal("首块应为心跳块")
	}
	if chunks[1].GetMode() == nil || chunks[2].GetTask() == nil {
		t.Fatalf("块序应为 heartbeat→mode→task")
	}
	if h.createCnt != 0 {
		t.Fatal("实现 CreateWithHeartbeat 时不应再调用原 Create")
	}
}

// TestCreateHeartbeatThrottledOnLiveInterval 服务端默认限频（liveness.HeartbeatInterval）：
// 等待点在限频窗内连续多次调用恰产生一块心跳
func TestCreateHeartbeatThrottledOnLiveInterval(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	h := &fakeHeartbeatCreateHandler{result: dto.BatchResult([]*dto.TaskCreateResponse{{}})}
	h.onCreate = func(reporter dto.HeartbeatReporter) {
		reporter.Heartbeat()
		reporter.Heartbeat()
		reporter.Heartbeat()
	}
	chunks := runCreate(t, h)
	hb := 0
	for _, c := range chunks {
		if c.GetHeartbeat() != nil {
			hb++
		}
	}
	if hb != 1 {
		t.Fatalf("限频窗内 3 次调用应恰 1 块心跳, 实得 %d 块", hb)
	}
}

// TestCreateHeartbeatNoopBelowMinHostVersion 协商版本不足时：走 CreateWithHeartbeat
// 路径但心跳 no-op，块序与未接入心跳完全一致（无心跳块）
func TestCreateHeartbeatNoopBelowMinHostVersion(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion-1)
	h := &fakeHeartbeatCreateHandler{result: dto.BatchResult([]*dto.TaskCreateResponse{{}})}
	h.onCreate = func(reporter dto.HeartbeatReporter) { reporter.Heartbeat() }
	chunks := runCreate(t, h)
	if len(chunks) != 2 {
		t.Fatalf("期望 2 块(mode+task), 实得 %d 块", len(chunks))
	}
	if chunks[0].GetMode() == nil || chunks[1].GetTask() == nil {
		t.Fatal("块序应为 mode→task")
	}
}

// TestCreateHeartbeatClosedAfterHandlerReturn 关闭舞步：handler 返回后上报器已
// 关闭——handler 保存的上报器引用再调用不再上块，结果块发送独占流
//（模拟消费方后台 ticker 未停的违约场景）
func TestCreateHeartbeatClosedAfterHandlerReturn(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	var saved dto.HeartbeatReporter
	h := &fakeHeartbeatCreateHandler{result: dto.BatchResult([]*dto.TaskCreateResponse{{}})}
	h.onCreate = func(reporter dto.HeartbeatReporter) {
		saved = reporter
		reporter.Heartbeat()
	}
	chunks := runCreate(t, h)
	saved.Heartbeat()
	saved.Heartbeat()
	hb := 0
	for _, c := range chunks {
		if c.GetHeartbeat() != nil {
			hb++
		}
	}
	if hb != 1 || len(chunks) != 3 {
		t.Fatalf("handler 返回后心跳不应上流, 期望恰 3 块含 1 心跳, 实得 %d 块含 %d 心跳", len(chunks), hb)
	}
}

// TestConcurrentCreateFlowsIsolatedReporters 并发多流程隔离：两条并发 Create 流
// 在对方等待期间各自上报心跳并返回结果，块序互不串扰——每条流独立上报器
//（流程级生命周期，无跨流锁与共享限频计时），-race 锚定
func TestConcurrentCreateFlowsIsolatedReporters(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	const flows = 2
	// 两流程各上报一次心跳后互相等待：制造真实并发窗再放行
	entered := make(chan struct{}, flows)
	release := make(chan struct{})

	results := make([][]*gen.CreateChunk, flows)
	errs := make([]error, flows)
	var wg sync.WaitGroup
	for i := 0; i < flows; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h := &fakeHeartbeatCreateHandler{result: dto.BatchResult([]*dto.TaskCreateResponse{{}})}
			h.onCreate = func(reporter dto.HeartbeatReporter) {
				reporter.Heartbeat()
				entered <- struct{}{}
				<-release
			}
			stream := &fakeCreateStream{}
			server := &workFetchServer{handler: h}
			errs[i] = server.Create(&gen.CreateRequest{Url: "https://example.com/work"}, stream)
			results[i] = stream.chunks
		}(i)
	}
	for i := 0; i < flows; i++ {
		<-entered
	}
	close(release)
	wg.Wait()

	for i := 0; i < flows; i++ {
		if errs[i] != nil {
			t.Fatalf("流程 %d Create 返回错误: %v", i, errs[i])
		}
		chunks := results[i]
		if len(chunks) != 3 || chunks[0].GetHeartbeat() == nil || chunks[1].GetMode() == nil || chunks[2].GetTask() == nil {
			t.Fatalf("流程 %d 块序应为 heartbeat→mode→task, 实得 %d 块", i, len(chunks))
		}
	}
}

// ========== Start/Resume 流等待期心跳接入面 ==========

// fakeStartStream Start 双向流替身：Recv 回放预置帧（耗尽后 io.EOF 结束 pull
// 循环），Send 收集全部块；服务端入口即读 Context()，须显式提供
type fakeStartStream struct {
	grpc.ServerStream
	ctx    context.Context
	frames []*gen.StartFrame
	chunks []*gen.StreamChunk
}

func (f *fakeStartStream) Context() context.Context { return f.ctx }

func (f *fakeStartStream) Send(c *gen.StreamChunk) error {
	f.chunks = append(f.chunks, c)
	return nil
}

func (f *fakeStartStream) Recv() (*gen.StartFrame, error) {
	if len(f.frames) == 0 {
		return nil, io.EOF
	}
	fr := f.frames[0]
	f.frames = f.frames[1:]
	return fr, nil
}

// fakeResumeStream Resume 双向流替身（同 fakeStartStream）
type fakeResumeStream struct {
	grpc.ServerStream
	ctx    context.Context
	frames []*gen.ResumeFrame
	chunks []*gen.StreamChunk
}

func (f *fakeResumeStream) Context() context.Context { return f.ctx }

func (f *fakeResumeStream) Send(c *gen.StreamChunk) error {
	f.chunks = append(f.chunks, c)
	return nil
}

func (f *fakeResumeStream) Recv() (*gen.ResumeFrame, error) {
	if len(f.frames) == 0 {
		return nil, io.EOF
	}
	fr := f.frames[0]
	f.frames = f.frames[1:]
	return fr, nil
}

// fakeStreamHandler Start/Resume handler 替身：行为经 startFn/resumeFn 注入
type fakeStreamHandler struct {
	dto.WorkFetcher
	startFn  func(ctx context.Context, task *dto.TaskDTO, storeRoles []string) ([]*dto.StoreSpec, *dto.WorkResponse, error)
	resumeFn func(ctx context.Context, param *dto.TaskResumeParam) ([]*dto.StoreSpec, *dto.WorkResponse, error)
}

func (f *fakeStreamHandler) Start(ctx context.Context, task *dto.TaskDTO, storeRoles []string) ([]*dto.StoreSpec, *dto.WorkResponse, error) {
	return f.startFn(ctx, task, storeRoles)
}

func (f *fakeStreamHandler) Resume(ctx context.Context, param *dto.TaskResumeParam) ([]*dto.StoreSpec, *dto.WorkResponse, error) {
	return f.resumeFn(ctx, param)
}

// TestStartHeartbeatBeforeFirstResponse Start handler 经 HeartbeatFromContext 取得
// 上报器：心跳块先于 WorkResponse/Specs 首响应上线；pull 循环收 io.EOF 正常收尾
func TestStartHeartbeatBeforeFirstResponse(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	h := &fakeStreamHandler{
		startFn: func(ctx context.Context, task *dto.TaskDTO, storeRoles []string) ([]*dto.StoreSpec, *dto.WorkResponse, error) {
			dto.HeartbeatFromContext(ctx).Heartbeat()
			return nil, &dto.WorkResponse{}, nil
		},
	}
	server := &workFetchServer{handler: h}
	stream := &fakeStartStream{
		ctx:    context.Background(),
		frames: []*gen.StartFrame{{Frame: &gen.StartFrame_Start{Start: &gen.StartRequest{}}}},
	}
	if err := server.Start(stream); err != nil {
		t.Fatalf("Start 返回错误: %v", err)
	}
	if len(stream.chunks) != 3 {
		t.Fatalf("期望 3 块(heartbeat+workResponse+specs), 实得 %d 块", len(stream.chunks))
	}
	if stream.chunks[0].GetHeartbeat() == nil {
		t.Fatal("首块应为心跳块")
	}
	if stream.chunks[1].GetWorkResponse() == nil || stream.chunks[2].GetSpecs() == nil {
		t.Fatal("块序应为 heartbeat→workResponse→specs")
	}
}

// TestStartWithoutHeartbeatCallBlockOrderUnchanged handler 不取上报器（未接入）：
// 块序恰 workResponse→specs，无心跳块，与接入前一致
func TestStartWithoutHeartbeatCallBlockOrderUnchanged(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	h := &fakeStreamHandler{
		startFn: func(ctx context.Context, task *dto.TaskDTO, storeRoles []string) ([]*dto.StoreSpec, *dto.WorkResponse, error) {
			return nil, &dto.WorkResponse{}, nil
		},
	}
	server := &workFetchServer{handler: h}
	stream := &fakeStartStream{
		ctx:    context.Background(),
		frames: []*gen.StartFrame{{Frame: &gen.StartFrame_Start{Start: &gen.StartRequest{}}}},
	}
	if err := server.Start(stream); err != nil {
		t.Fatalf("Start 返回错误: %v", err)
	}
	if len(stream.chunks) != 2 {
		t.Fatalf("期望 2 块(workResponse+specs), 实得 %d 块", len(stream.chunks))
	}
	if stream.chunks[0].GetWorkResponse() == nil || stream.chunks[1].GetSpecs() == nil {
		t.Fatal("块序应为 workResponse→specs")
	}
}

// TestResumeHeartbeatBeforeFirstResponse Resume 同构：心跳块先于首响应上线
func TestResumeHeartbeatBeforeFirstResponse(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	h := &fakeStreamHandler{
		resumeFn: func(ctx context.Context, param *dto.TaskResumeParam) ([]*dto.StoreSpec, *dto.WorkResponse, error) {
			dto.HeartbeatFromContext(ctx).Heartbeat()
			return nil, &dto.WorkResponse{}, nil
		},
	}
	server := &workFetchServer{handler: h}
	stream := &fakeResumeStream{
		ctx:    context.Background(),
		frames: []*gen.ResumeFrame{{Frame: &gen.ResumeFrame_Resume{Resume: &gen.TaskResumeParamMessage{}}}},
	}
	if err := server.Resume(stream); err != nil {
		t.Fatalf("Resume 返回错误: %v", err)
	}
	if len(stream.chunks) != 3 {
		t.Fatalf("期望 3 块(heartbeat+workResponse+specs), 实得 %d 块", len(stream.chunks))
	}
	if stream.chunks[0].GetHeartbeat() == nil {
		t.Fatal("首块应为心跳块")
	}
	if stream.chunks[1].GetWorkResponse() == nil || stream.chunks[2].GetSpecs() == nil {
		t.Fatal("块序应为 heartbeat→workResponse→specs")
	}
}

// TestResumeWithoutHeartbeatCallBlockOrderUnchanged handler 不取上报器（未接入）：
// 块序恰 workResponse→specs，与接入前一致
func TestResumeWithoutHeartbeatCallBlockOrderUnchanged(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	h := &fakeStreamHandler{
		resumeFn: func(ctx context.Context, param *dto.TaskResumeParam) ([]*dto.StoreSpec, *dto.WorkResponse, error) {
			return nil, &dto.WorkResponse{}, nil
		},
	}
	server := &workFetchServer{handler: h}
	stream := &fakeResumeStream{
		ctx:    context.Background(),
		frames: []*gen.ResumeFrame{{Frame: &gen.ResumeFrame_Resume{Resume: &gen.TaskResumeParamMessage{}}}},
	}
	if err := server.Resume(stream); err != nil {
		t.Fatalf("Resume 返回错误: %v", err)
	}
	if len(stream.chunks) != 2 {
		t.Fatalf("期望 2 块(workResponse+specs), 实得 %d 块", len(stream.chunks))
	}
	if stream.chunks[0].GetWorkResponse() == nil || stream.chunks[1].GetSpecs() == nil {
		t.Fatal("块序应为 workResponse→specs")
	}
}

// TestActivateStoresHostContractVersion Activate 时宿主契约版本落进程级原子值
//（HostServiceId=0 时不 dial broker，仅记录版本），经 HostContractVersion 可读
func TestActivateStoresHostContractVersion(t *testing.T) {
	srv := &lifecycleServer{}
	if _, err := srv.Activate(t.Context(), &gen.ActivateRequest{HostContractVersion: 14}); err != nil {
		t.Fatalf("Activate 失败: %v", err)
	}
	t.Cleanup(func() { hostContractVersion.Store(0) })
	if got := HostContractVersion(); got != 14 {
		t.Fatalf("Activate 后协商版本应为 14, 实得 %d", got)
	}
}
