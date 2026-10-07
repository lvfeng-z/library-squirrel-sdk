package transport

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lvfeng-z/library-squirrel-sdk/gen"
)

// HeartbeatMinHostVersion 流等待期心跳所需的最低宿主契约版本：宿主经 Activate
// 协商字段下发的版本低于此值（含未 Activate、字段缺省 0）时，心跳上报为 no-op
//——旧宿主不识别心跳块，发送会破坏其流解析
const HeartbeatMinHostVersion uint32 = 14

// hostContractVersion 宿主契约版本（进程级）：宿主随 Activate 下发，一次协商
// 覆盖全部流。原子访问：Activate 与各流式 RPC 的心跳上报并发
var hostContractVersion atomic.Uint32

// HostContractVersion 返回宿主经 Activate 下发的契约版本，未 Activate 时为 0。
// 供插件按宿主能力自适应（如长等待上限按宿主是否容忍心跳分档）
func HostContractVersion() uint32 {
	return hostContractVersion.Load()
}

// heartbeatReporter 流等待期心跳上报器。生命周期粒度=每次流式 RPC 调用（流程级）：
// 服务端为每条流各构造一个实例，各绑各的 Send、各自互斥锁与限频计时——并发
// 多流程（Create×2、Create+Start 并行等）互不干扰，限频按流程独立计。
//
// 并发安全：互斥锁串行化全部心跳发送与关闭（gRPC stream.Send 非并发安全，
// 跨协程调用合法）。内置限频：距上次实际发送不足间隔时静默吞掉，消费方可以
// 任意 tick 频率调用而不产生线上噪音。Close 后任何在途/后续 Heartbeat 不再
// 发送——handler 返回后由服务端调用，随后的结果块/首响应发送独占流。
type heartbeatReporter struct {
	send     func() error // 向绑定流发送一个心跳块
	interval time.Duration

	mu       sync.Mutex
	lastSend time.Time
	closed   bool
}

// newCreateHeartbeatReporter 构造向 Create 流发送心跳块的上报器；
// interval 为限频间隔，测试可注入缩短值
func newCreateHeartbeatReporter(send func(*gen.CreateChunk) error, interval time.Duration) *heartbeatReporter {
	return &heartbeatReporter{
		send: func() error {
			return send(&gen.CreateChunk{Payload: &gen.CreateChunk_Heartbeat{Heartbeat: &gen.Heartbeat{}}})
		},
		interval: interval,
	}
}

// newStreamHeartbeatReporter 构造向 Start/Resume 流发送心跳块的上报器；
// interval 为限频间隔，测试可注入缩短值
func newStreamHeartbeatReporter(send func(*gen.StreamChunk) error, interval time.Duration) *heartbeatReporter {
	return &heartbeatReporter{
		send: func() error {
			return send(&gen.StreamChunk{Payload: &gen.StreamChunk_Heartbeat{Heartbeat: &gen.Heartbeat{}}})
		},
		interval: interval,
	}
}

// Heartbeat 上报一次心跳。宿主契约版本低于心跳能力版本时 no-op；限频窗内与
// 已关闭时静默跳过；发送失败（流已断或发送异常）记日志不中断 handler——
// handler 的结果块发送会以同因失败并正常返回错误
func (r *heartbeatReporter) Heartbeat() {
	if hostContractVersion.Load() < HeartbeatMinHostVersion {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	now := time.Now()
	if !r.lastSend.IsZero() && now.Sub(r.lastSend) < r.interval {
		return
	}
	if err := r.send(); err != nil {
		log.Printf("[heartbeat] 心跳块发送失败(流已断或发送异常), handler 继续执行: %v", err)
	}
	r.lastSend = now
}

// Close 停用上报器：此后任何在途/后续 Heartbeat 不再发送。关闭与在途发送经
// 同一互斥锁串行，关闭返回后流上不再出现心跳块
func (r *heartbeatReporter) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
}
