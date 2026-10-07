package dto

import "context"

// HeartbeatReporter 流等待期心跳上报器：插件在流式 RPC（Create/Start/Resume）
// handler 执行期的长等待点（等用户输入、串行外部请求、重试退避等自身带界的
// 等待）周期调用 Heartbeat，向当前流发送心跳块维持宿主空闲检测窗。
//
// 心跳语义=「本等待点有界且仍在推进」：每个上报心跳的等待点必须保有自身超时
//（问询弹窗有 dismiss 上限、扫码窗有二维码有效期、网络调用有 HTTP 超时），
// 永不超时的等待上报心跳会使宿主 hang 检测失效（「创建中」通知长期滞留）。
//
// 方法并发安全，可跨协程调用：消费方可为单个长阻塞调用挂后台 ticker，
// 等待终结时停止 ticker。SDK 保证 handler 返回后上报器关闭、随后的结果块
// 发送独占流，插件不管理生命周期。
type HeartbeatReporter interface {
	// Heartbeat 上报一次心跳。限频窗内、宿主契约版本低于心跳能力版本、已关闭时
	// 均为静默 no-op；发送失败仅记日志不报错（流已断时 handler 的结果块发送
	// 会以同因失败并正常返回错误）
	Heartbeat()
}

// heartbeatCtxKey Start/Resume ctx 注入心跳上报器的私有键
type heartbeatCtxKey struct{}

// WithHeartbeat 把心跳上报器放入 ctx。SDK 服务端在调用 Start/Resume handler 前
// 注入；插件代码不调用本函数，经 HeartbeatFromContext 取用
func WithHeartbeat(ctx context.Context, reporter HeartbeatReporter) context.Context {
	return context.WithValue(ctx, heartbeatCtxKey{}, reporter)
}

// HeartbeatFromContext 从 ctx 取流等待期心跳上报器；未注入（Start/Resume handler
// 之外的调用路径）时返回 no-op 实现，调用方无需判空
func HeartbeatFromContext(ctx context.Context) HeartbeatReporter {
	if r, ok := ctx.Value(heartbeatCtxKey{}).(HeartbeatReporter); ok {
		return r
	}
	return NoopHeartbeat{}
}

// NoopHeartbeat 心跳上报器空实现：任何调用无效果。供未注入上报器的路径占位
//（如插件 Create 委托 CreateWithHeartbeat 时的无心跳路径）
type NoopHeartbeat struct{}

// Heartbeat 实现 HeartbeatReporter，无效果
func (NoopHeartbeat) Heartbeat() {}
