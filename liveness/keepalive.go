// Package liveness 汇聚插件 RPC 活动检测的策略:keepalive 连接级无响应检测、
// unary 调用超时、reader 流空闲超时。阈值与 gRPC options 集中在此定义,
// 主程序(client)与插件(server)两端引用同一组常量,保证双端 keepalive 参数一致。
package liveness

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	// KeepalivePingInterval 连接无活动 stream 时发送 keepalive ping 的间隔
	KeepalivePingInterval = 10 * time.Second
	// KeepalivePingTimeout 发出 ping 后等待 pong 的时长,超时则判定连接已死(对端进程崩溃或网络中断)
	KeepalivePingTimeout = 5 * time.Second
	// enforcementMinTime server 端 EnforcementPolicy.MinTime;grpc 中 server 仅接受发送间隔 ≥ MinTime 的 client ping,故取 KeepalivePingInterval 之半
	enforcementMinTime = 5 * time.Second

	// UnaryRPCTimeout CreateWorkInfo/Pause/Stop/Retry 等 unary 调用的超时;正常应秒级返回,
	// 达此阈值说明插件 handler 卡死——keepalive 探测不到此类故障(gRPC 连接层仍照常回 ping)
	UnaryRPCTimeout = 30 * time.Second

	// ReaderIdleTimeout Start/Resume 流式读取时两次收到数据之间允许的最大空闲,
	// 超过判应用级无响应(连接活但插件在等上游 HTTP);由 reader 侧实施
	ReaderIdleTimeout = 60 * time.Second
)

// ClientDialOptions 返回主程序连接插件子进程时注入的 gRPC dial options,启用 client 主动 keepalive 探测。
// PermitWithoutStream=true 使 client 在无活动 stream 时也发 ping:unary 调用间隙连接无活动 stream,需开启此项才持续探测进程存活。
func ClientDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                KeepalivePingInterval,
			Timeout:             KeepalivePingTimeout,
			PermitWithoutStream: true,
		}),
	}
}

// ServerOptions 返回插件创建 gRPC server 时追加的 server options,启用 keepalive 响应与频率约束。
// EnforcementPolicy.MinTime 取 enforcementMinTime,小于 client 的 ping 间隔(KeepalivePingInterval),server 据此接受 client 的 ping。
func ServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    KeepalivePingInterval,
			Timeout: KeepalivePingTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             enforcementMinTime,
			PermitWithoutStream: true,
		}),
	}
}

// GRPCServerFactory 供 go-plugin 的 ServeConfig.GRPCServer 字段使用:在 go-plugin 传入的 opts
// (含 TLS、内部 stream 拦截器等框架必需 option)基础上追加 keepalive server options,再创建 server。
// 保留传入的 opts——框架注入的 option 为 server 正常工作所必需。
func GRPCServerFactory(opts []grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(append(opts, ServerOptions()...)...)
}
