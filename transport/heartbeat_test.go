package transport

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// setHostContractVersionForTest 设置进程级协商版本，测试结束恢复 0（未 Activate 态）
func setHostContractVersionForTest(t *testing.T, v uint32) {
	t.Helper()
	hostContractVersion.Store(v)
	t.Cleanup(func() { hostContractVersion.Store(0) })
}

// countSendFn 构造计数发送函数；计数用原子变量以在 -race 下稳妥
//（上报器互斥锁保证实际串行调用）
func countSendFn(err error) (send func() error, calls *atomic.Int64) {
	var n atomic.Int64
	send = func() error {
		n.Add(1)
		return err
	}
	return send, &n
}

// TestHeartbeatThrottledWithinInterval 限频：限频窗内多次调用恰发一块；
// 窗过后再调用应发送（回拨上次发送时刻模拟窗流逝）
func TestHeartbeatThrottledWithinInterval(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	send, calls := countSendFn(nil)
	r := &heartbeatReporter{send: send, interval: time.Hour}

	r.Heartbeat()
	r.Heartbeat()
	r.Heartbeat()
	if got := calls.Load(); got != 1 {
		t.Fatalf("限频窗内 3 次调用应恰发 1 块, 实发 %d 块", got)
	}

	r.mu.Lock()
	r.lastSend = time.Now().Add(-2 * time.Hour)
	r.mu.Unlock()
	r.Heartbeat()
	if got := calls.Load(); got != 2 {
		t.Fatalf("限频窗过后应再次发送, 实发 %d 块", got)
	}
}

// TestHeartbeatReportersIndependentThrottle 流程级限频隔离：两个上报器实例
//（两条流各一）限频独立计——A 消耗掉自身限频窗后 B 首次发送不受影响，
// 并发流程不互相饿死
func TestHeartbeatReportersIndependentThrottle(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	sendA, callsA := countSendFn(nil)
	sendB, callsB := countSendFn(nil)
	a := &heartbeatReporter{send: sendA, interval: time.Hour}
	b := &heartbeatReporter{send: sendB, interval: time.Hour}

	a.Heartbeat()
	b.Heartbeat()
	if callsA.Load() != 1 || callsB.Load() != 1 {
		t.Fatalf("两条流应各自发送 1 块, 实得 A=%d B=%d", callsA.Load(), callsB.Load())
	}
	a.Heartbeat()
	if callsA.Load() != 1 {
		t.Fatalf("A 限频窗内第二次应被吞, 实发 %d", callsA.Load())
	}
}

// TestHeartbeatNoopBelowMinHostVersion 协商门控：宿主版本低于心跳能力版本
//（含未 Activate 的 0）时 no-op，不发送
func TestHeartbeatNoopBelowMinHostVersion(t *testing.T) {
	for _, v := range []uint32{0, 1, HeartbeatMinHostVersion - 1} {
		setHostContractVersionForTest(t, v)
		send, calls := countSendFn(nil)
		r := &heartbeatReporter{send: send, interval: 0}
		r.Heartbeat()
		if got := calls.Load(); got != 0 {
			t.Fatalf("宿主版本 %d 时应 no-op, 实发 %d 块", v, got)
		}
	}
}

// TestHeartbeatSendFailureLoggedNotFatal 发送失败不中断：send 返回错误时
// Heartbeat 正常返回，后续调用继续尝试发送
func TestHeartbeatSendFailureLoggedNotFatal(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	send, calls := countSendFn(errors.New("stream closed"))
	r := &heartbeatReporter{send: send, interval: 0}
	r.Heartbeat()
	r.Heartbeat()
	if got := calls.Load(); got != 2 {
		t.Fatalf("发送失败后应继续尝试发送, 实尝试 %d 次", got)
	}
}

// TestHeartbeatClosedNoSend 关闭后不再发送：Close 返回后任何 Heartbeat 均为 no-op
func TestHeartbeatClosedNoSend(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	send, calls := countSendFn(nil)
	r := &heartbeatReporter{send: send, interval: 0}
	r.Heartbeat()
	r.Close()
	r.Heartbeat()
	r.Heartbeat()
	if got := calls.Load(); got != 1 {
		t.Fatalf("Close 后不应再发送, 实发 %d 块", got)
	}
}

// TestHeartbeatConcurrentCallAndCloseRace 并发压测：多协程高频调用 Heartbeat
// 与主协程 Close 并发（关闭与在途发送经同一互斥锁串行），-race 锚定无数据
// 竞态；全部调用方结束后再调 Heartbeat，发送计数不再增长
func TestHeartbeatConcurrentCallAndCloseRace(t *testing.T) {
	setHostContractVersionForTest(t, HeartbeatMinHostVersion)
	send, calls := countSendFn(nil)
	r := &heartbeatReporter{send: send, interval: 0}

	const workers = 8
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 2000; j++ {
				r.Heartbeat()
			}
		}()
	}
	r.Close()
	wg.Wait()

	after := calls.Load()
	r.Heartbeat()
	if calls.Load() != after {
		t.Fatal("Close 后发送计数不应增长")
	}
}

// TestHostContractVersionAccessor 访问器返回进程级协商值，未 Activate 时为 0
func TestHostContractVersionAccessor(t *testing.T) {
	if got := HostContractVersion(); got != 0 {
		t.Fatalf("未 Activate 时应为 0, 实得 %d", got)
	}
	setHostContractVersionForTest(t, 14)
	if got := HostContractVersion(); got != 14 {
		t.Fatalf("应返回协商值 14, 实得 %d", got)
	}
}
