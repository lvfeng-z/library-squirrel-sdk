package proxy

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"github.com/lvfeng-z/library-squirrel-sdk/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// envProxyURL 环境变量代理桩地址。全测试统一值：标准库 ProxyFromEnvironment 的
// 环境快照进程内只做一次（gRPC 目标解析内部也会触发同一快照），首个触发快照的
// 时刻可能早于任何用例的桩设置，进程启动即钉住桩值，保证全量跑与单用例跑的
// env 断言同期望
const envProxyURL = "http://127.0.0.1:7777"

// TestMain 预置进程级环境代理桩：见 envProxyURL 注释的快照时序说明
func TestMain(m *testing.M) {
	_ = os.Setenv("HTTP_PROXY", envProxyURL)
	_ = os.Setenv("HTTPS_PROXY", envProxyURL)
	_ = os.Setenv("NO_PROXY", "")
	os.Exit(m.Run())
}

// stubEnv 注入环境变量代理桩（请求目标用非环回地址：环境代理解析对环回地址恒直连）
func stubEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HTTP_PROXY", envProxyURL)
	t.Setenv("HTTPS_PROXY", envProxyURL)
	t.Setenv("NO_PROXY", "")
}

// resetProxyState 清空包级注入状态并注册用例后清理
func resetProxyState(t *testing.T) {
	t.Helper()
	state.Store(nil)
	t.Cleanup(func() { state.Store(nil) })
}

// mustRequest 构造 GET 请求（代理函数只读 URL，不发真实网络请求）
func mustRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("构造请求失败: %v", err)
	}
	return req
}

// serveHost 在环回监听上启动 gRPC server 并返回接通的客户端连接（用例结束自动清理）
func serveHost(t *testing.T, register func(*grpc.Server)) *grpc.ClientConn {
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

// countLogger 告警计数日志桩：验证降级一次性告警
type countLogger struct {
	warns int
}

func (l *countLogger) Debugf(string, ...any)   {}
func (l *countLogger) Infof(string, ...any)    {}
func (l *countLogger) Warnf(string, ...any)    { l.warns++ }
func (l *countLogger) Errorf(string, ...any)   {}
func (l *countLogger) Named(string) dto.Logger { return l }

// resolveStub 代理解析桩：回放固定决议并记录收到的参数
type resolveStub struct {
	proxyURL    string
	source      string
	gotExplicit string
	gotRequest  string
}

func (f *resolveStub) ResolveProxy(_ context.Context, explicitURL, requestURL string) (string, string, error) {
	f.gotExplicit = explicitURL
	f.gotRequest = requestURL
	return f.proxyURL, f.source, nil
}

// hangingResolveStub 挂起桩：不回包直至调用侧取消（连接关闭或超时取消传导）
type hangingResolveStub struct{}

func (hangingResolveStub) ResolveProxy(ctx context.Context, explicitURL, requestURL string) (string, string, error) {
	<-ctx.Done()
	return "", "", ctx.Err()
}

// legacyHostServiceDesc 裁掉 ResolveProxy 方法的服务描述：模拟旧宿主（无代理解析
// RPC）的 HostService 注册形态——旧 gen 代码无此方法，gRPC 对未注册方法回 Unimplemented
func legacyHostServiceDesc() *grpc.ServiceDesc {
	desc := gen.HostService_ServiceDesc // 值拷贝，改切片头不触原变量
	methods := make([]grpc.MethodDesc, 0, len(desc.Methods))
	for _, m := range desc.Methods {
		if m.MethodName == "ResolveProxy" {
			continue
		}
		methods = append(methods, m)
	}
	desc.Methods = methods
	return &desc
}

// TestTransportUsesHostResolution 决议透传：stub 宿主回 (url, source)，Transport 的
// 按请求代理函数生效——请求目标透传宿主、决议地址落到 Transport、空决议 = 直连
func TestTransportUsesHostResolution(t *testing.T) {
	resetProxyState(t)
	stub := &resolveStub{proxyURL: "http://127.0.0.1:8888", source: "system"}
	conn := serveHost(t, func(s *grpc.Server) {
		transport.RegisterHostService(s, transport.HostDeps{ProxyResolveProvider: stub})
	})
	Configure(transport.NewPluginContextClient(conn), "", &countLogger{})
	tr := NewTransport(false)

	req := mustRequest(t, "https://www.example.com/a")
	u, err := tr.Proxy(req)
	if err != nil {
		t.Fatalf("Transport 代理函数报错: %v", err)
	}
	if u == nil || u.String() != "http://127.0.0.1:8888" {
		t.Fatalf("宿主决议地址未生效: %v", u)
	}
	if stub.gotExplicit != "" {
		t.Fatalf("显式代理应透传空值(纯自动检测), 实得 %q", stub.gotExplicit)
	}
	if stub.gotRequest != "https://www.example.com/a" {
		t.Fatalf("请求目标未透传宿主: %q", stub.gotRequest)
	}

	// 空决议 = 直连 (nil, nil)
	stub.proxyURL = ""
	u, err = tr.Proxy(req)
	if err != nil {
		t.Fatalf("直连决议报错: %v", err)
	}
	if u != nil {
		t.Fatalf("空决议应直连(nil), 实得 %v", u)
	}

	// 显式代理透传：Configure 注入值原样到达宿主
	Configure(transport.NewPluginContextClient(conn), "http://127.0.0.1:6666", &countLogger{})
	stub.proxyURL = "http://127.0.0.1:6666"
	if _, err := tr.Proxy(req); err != nil {
		t.Fatalf("显式代理请求报错: %v", err)
	}
	if stub.gotExplicit != "http://127.0.0.1:6666" {
		t.Fatalf("显式代理未透传宿主: %q", stub.gotExplicit)
	}
}

// TestLegacyHostFallbackExplicitThenEnv 旧宿主（ResolveProxy 未注册，gRPC 回
// Unimplemented）降级：显式代理仍生效（降级链第一级）、无显式值时 env-only、
// 降级告警一次性
func TestLegacyHostFallbackExplicitThenEnv(t *testing.T) {
	resetProxyState(t)
	stubEnv(t)
	conn := serveHost(t, func(s *grpc.Server) {
		s.RegisterService(legacyHostServiceDesc(), transport.NewHostServiceServer(transport.HostDeps{}))
	})
	logger := &countLogger{}
	Configure(transport.NewPluginContextClient(conn), "http://127.0.0.1:6666", logger)
	tr := NewTransport(true)

	req := mustRequest(t, "https://www.example.com/a")
	for i := 0; i < 2; i++ {
		u, err := tr.Proxy(req)
		if err != nil {
			t.Fatalf("降级代理函数报错: %v", err)
		}
		if u == nil || u.String() != "http://127.0.0.1:6666" {
			t.Fatalf("显式代理应仍生效(降级链第一级): %v", u)
		}
	}
	if logger.warns != 1 {
		t.Fatalf("降级告警应一次性(共 1 次), 实得 %d 次", logger.warns)
	}

	// 无显式值：env-only（系统代理层缺失，env 层按请求目标匹配）
	loggerEnv := &countLogger{}
	Configure(transport.NewPluginContextClient(conn), "", loggerEnv)
	u, err := tr.Proxy(req)
	if err != nil {
		t.Fatalf("env 降级报错: %v", err)
	}
	if u == nil || u.String() != envProxyURL {
		t.Fatalf("env-only 降级应回落环境变量 %s, 实得 %v", envProxyURL, u)
	}
	if loggerEnv.warns != 1 {
		t.Fatalf("env 降级也应告警一次, 实得 %d 次", loggerEnv.warns)
	}
}

// TestHangingHostTimeoutFallback 宿主挂起不回包：按超时预算降级（预算注入缩短为
// 50ms，不真等默认 1 秒），env-only 回落
func TestHangingHostTimeoutFallback(t *testing.T) {
	resetProxyState(t)
	stubEnv(t)
	resolveTimeoutNanos.Store(int64(50 * time.Millisecond))
	t.Cleanup(func() { resolveTimeoutNanos.Store(0) })
	conn := serveHost(t, func(s *grpc.Server) {
		transport.RegisterHostService(s, transport.HostDeps{ProxyResolveProvider: hangingResolveStub{}})
	})
	logger := &countLogger{}
	Configure(transport.NewPluginContextClient(conn), "", logger)
	tr := NewTransport(true)

	req := mustRequest(t, "https://www.example.com/a")
	start := time.Now()
	u, err := tr.Proxy(req)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("挂起降级报错: %v", err)
	}
	if u == nil || u.String() != envProxyURL {
		t.Fatalf("超时降级应回落环境变量 %s, 实得 %v", envProxyURL, u)
	}
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("应按注入预算(50ms)降级, 实耗时 %v", elapsed)
	}
	if logger.warns != 1 {
		t.Fatalf("超时降级应告警一次, 实得 %d 次", logger.warns)
	}
}

// TestNewTransportBeforeConfigureEnvOnly Configure 前调用 NewTransport（测试路径
// 形态）：不 panic，代理决策回落 env-only
func TestNewTransportBeforeConfigureEnvOnly(t *testing.T) {
	resetProxyState(t)
	stubEnv(t)
	tr := NewTransport(true)

	u, err := tr.Proxy(mustRequest(t, "https://www.example.com/zero")) // 不 panic
	if err != nil {
		t.Fatalf("零值状态代理函数报错: %v", err)
	}
	if u == nil || u.String() != envProxyURL {
		t.Fatalf("零值状态应 env-only 回落 %s, 实得 %v", envProxyURL, u)
	}
}
