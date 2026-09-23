package transport

import (
	"errors"
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
func runCreate(t *testing.T, handler *fakeCreateHandler) []*gen.CreateChunk {
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
