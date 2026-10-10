// Package proxy 提供插件出网默认 Transport：代理决策经宿主 ResolveProxy RPC 三级
// 检测（显式代理 > 系统代理 > 环境变量）逐请求决议，代理开关对下一请求即时生效；
// 宿主解析不可用（失败/超时/旧宿主 Unimplemented）时降级为显式代理 > 环境变量，
// 请求不因解析失败或超时而失败
package proxy

import (
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
)

// defaultResolveTimeout 每次代理解析的默认超时预算
const defaultResolveTimeout = time.Second

// resolveTimeoutNanos 超时预算（纳秒），0 = 默认值。预算到点即降级不再等待宿主；
// 测试注入缩短值避免真等满预算
var resolveTimeoutNanos atomic.Int64

// proxyState Configure 注入的运行态。Configure（激活期一次）与 NewTransport /
// 按请求求值（后续并发）以原子指针承载并发安全
type proxyState struct {
	ctx              dto.Proxy // 收窄：本包只消费代理解析族
	explicitProxyURL string
	logger           dto.Logger
	warned           atomic.Bool // 降级告警一次性标记
}

var state atomic.Pointer[proxyState]

func resolveBudget() time.Duration {
	if v := resolveTimeoutNanos.Load(); v > 0 {
		return time.Duration(v)
	}
	return defaultResolveTimeout
}

// Configure 注入宿主上下文与显式代理（激活期一次，早于任何网络请求）。
// explicitProxyURL 为插件设置 proxyUrl 透传（空 = 纯自动检测）；打一条来源日志。
func Configure(ctx dto.Proxy, explicitProxyURL string, logger dto.Logger) {
	state.Store(&proxyState{ctx: ctx, explicitProxyURL: explicitProxyURL, logger: logger})
	if logger == nil {
		return
	}
	if explicitProxyURL != "" {
		logger.Infof("代理来源：显式代理")
	} else {
		logger.Infof("代理来源：自动检测（系统代理 > 环境变量）")
	}
}

// NewTransport 构造插件出网 Transport：代理决策 = 宿主 ResolveProxy RPC（每请求求值）；
// disableKeepAlives=true 禁用连接复用（API 路径），false 启用连接池（IdleConnTimeout 30s、
// MaxIdleConnsPerHost 4、ResponseHeaderTimeout 30s、ForceAttemptHTTP2）。
func NewTransport(disableKeepAlives bool) *http.Transport {
	t := &http.Transport{
		DisableKeepAlives: disableKeepAlives,
		Proxy:             resolveProxy,
	}
	if !disableKeepAlives {
		t.IdleConnTimeout = 30 * time.Second
		t.MaxIdleConnsPerHost = 4
		t.ResponseHeaderTimeout = 30 * time.Second
		t.ForceAttemptHTTP2 = true
	}
	return t
}

// resolveProxy Transport 的按请求代理决策
func resolveProxy(req *http.Request) (*url.URL, error) {
	st := state.Load()
	if st == nil || st.ctx == nil {
		// 未 Configure（或未注入上下文）：env-only 回落
		return http.ProxyFromEnvironment(req)
	}
	requestURL := ""
	if req.URL != nil {
		requestURL = req.URL.String()
	}
	proxyURL, err := resolveWithBudget(st, requestURL)
	if err != nil {
		warnFallbackOnce(st, err)
		return fallbackProxy(req, st.explicitProxyURL)
	}
	if proxyURL == "" {
		return nil, nil // 宿主决议：直连
	}
	u, err := parseProxyURL(proxyURL)
	if err != nil {
		// 决议地址不可解析：与解析失败同走降级链
		warnFallbackOnce(st, err)
		return fallbackProxy(req, st.explicitProxyURL)
	}
	return u, nil
}

// resolveWithBudget 在超时预算内等待宿主决议。Proxy 的解析方法无 ctx 参数
// （客户端实现内部另有 1 秒硬顶取消），此处竞速兜底超出预算的挂起：到点即放弃等待，
// 迟到的决议写入带缓冲通道后被丢弃，决议 goroutine 自行退出不阻塞
func resolveWithBudget(st *proxyState, requestURL string) (string, error) {
	type outcome struct {
		proxyURL string
		err      error
	}
	done := make(chan outcome, 1)
	go func() {
		proxyURL, _, err := st.ctx.ResolveProxy(st.explicitProxyURL, requestURL)
		done <- outcome{proxyURL: proxyURL, err: err}
	}()
	timer := time.NewTimer(resolveBudget())
	defer timer.Stop()
	select {
	case out := <-done:
		return out.proxyURL, out.err
	case <-timer.C:
		return "", errors.New("代理解析超出超时预算")
	}
}

// warnFallbackOnce 降级一次性告警：首次降级打告警日志，后续降级静默——宿主不可用
// 是持续状态，逐请求重复告警只会淹没日志
func warnFallbackOnce(st *proxyState, err error) {
	if st.logger != nil && st.warned.CompareAndSwap(false, true) {
		st.logger.Warnf("宿主代理解析不可用，出网代理降级为显式设置/环境变量：%v", err)
	}
}

// fallbackProxy 降级链：显式代理 > 环境变量。显式值 Configure 时本地已知，不受
// 宿主解析失败影响；env 层按请求目标做 NO_PROXY 例外匹配（标准库语义）
func fallbackProxy(req *http.Request, explicitProxyURL string) (*url.URL, error) {
	if explicitProxyURL != "" {
		return parseProxyURL(explicitProxyURL)
	}
	return http.ProxyFromEnvironment(req)
}

// parseProxyURL 解析代理地址；缺 scheme 或缺 host 时按缺省 http 代理补全重试
// （与标准库环境变量代理解析同规则）
func parseProxyURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		if u, err = url.Parse("http://" + raw); err != nil {
			return nil, err
		}
	}
	return u, nil
}
