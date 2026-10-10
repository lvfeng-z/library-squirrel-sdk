package transport

import (
	"context"
	"io"
	"time"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"google.golang.org/grpc"
)

// PluginContextClient 插件侧的 PluginContext 实现，通过 gRPC 调用主程序的 HostService 与 LibraryQuery
type PluginContextClient struct {
	hostClient     gen.HostServiceClient
	queryClient    gen.LibraryQueryClient
	logger         dto.Logger
	mainWindowHWND uintptr
	subCancelFuncs map[string]context.CancelFunc
}

// NewPluginContextClient 创建基于 gRPC 的 PluginContext 客户端
func NewPluginContextClient(conn *grpc.ClientConn) *PluginContextClient {
	return &PluginContextClient{
		hostClient:     gen.NewHostServiceClient(conn),
		queryClient:    gen.NewLibraryQueryClient(conn),
		logger:         NewGRPCLogger(gen.NewHostServiceClient(conn)),
		subCancelFuncs: make(map[string]context.CancelFunc),
	}
}

// SetMainWindowHandle 设置主窗口句柄
func (c *PluginContextClient) SetMainWindowHandle(hwnd uintptr) {
	c.mainWindowHWND = hwnd
}

func (c *PluginContextClient) GetValue(key string) (*dto.StorageValue, error) {
	resp, err := c.hostClient.GetValue(context.Background(), &gen.StorageKeyRequest{Key: key})
	if err != nil {
		return nil, err
	}
	return &dto.StorageValue{Value: resp.Value, SchemaVersion: resp.SchemaVersion}, nil
}

func (c *PluginContextClient) SetValue(key string, value string) error {
	_, err := c.hostClient.SetValue(context.Background(), &gen.StorageEntryRequest{
		Key:   key,
		Value: value,
	})
	return err
}

func (c *PluginContextClient) SetValueEncrypted(key string, value string) error {
	_, err := c.hostClient.SetValueEncrypted(context.Background(), &gen.StorageEntryRequest{
		Key:   key,
		Value: value,
	})
	return err
}

func (c *PluginContextClient) DeleteValue(key string) error {
	_, err := c.hostClient.DeleteValue(context.Background(), &gen.StorageKeyRequest{Key: key})
	return err
}

func (c *PluginContextClient) GetAllValues() (map[string]*dto.StorageValue, error) {
	resp, err := c.hostClient.GetAllValues(context.Background(), &gen.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

func (c *PluginContextClient) GetPreference(key string) (*dto.PreferenceValue, bool, error) {
	resp, err := c.hostClient.GetPreference(context.Background(), &gen.PreferenceKeyRequest{Key: key})
	if err != nil {
		return nil, false, err
	}
	if !resp.Ok {
		// 无记录是合法状态（用户已删除或从未写入），不是错误——调用方据此重新发起问答
		return nil, false, nil
	}
	return resp.Value, true, nil
}

func (c *PluginContextClient) SetPreference(key string, value *dto.PreferenceValue) error {
	_, err := c.hostClient.SetPreference(context.Background(), &gen.PreferenceEntryRequest{
		Key:   key,
		Value: value,
	})
	return err
}

func (c *PluginContextClient) ListMyPreferences() ([]string, error) {
	resp, err := c.hostClient.ListMyPreferences(context.Background(), &gen.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.Keys, nil
}

// resolveProxyTimeout 单次代理解析 RPC 的等待上限：决议是每个出网请求前的同步
// 步骤，宿主繁忙或卡顿时慢挂起会拖住插件全部出网请求，超时即返回错误由调用方
// 走降级链
const resolveProxyTimeout = time.Second

func (c *PluginContextClient) ResolveProxy(explicitURL, requestURL string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), resolveProxyTimeout)
	defer cancel()
	resp, err := c.hostClient.ResolveProxy(ctx, &gen.ResolveProxyRequest{
		ExplicitUrl: explicitURL,
		RequestUrl:  requestURL,
	})
	if err != nil {
		return "", "", err
	}
	return resp.ProxyUrl, resp.Source, nil
}

func (c *PluginContextClient) CreateTask(url string) (*dto.CreateTaskResult, error) {
	resp, err := c.hostClient.CreateTask(context.Background(), &gen.CreateTaskRequest{Url: url})
	if err != nil {
		return nil, err
	}
	return &dto.CreateTaskResult{
		Succeed:       resp.Succeed,
		AddedQuantity: int(resp.AddedQuantity),
		Msg:           resp.Msg,
	}, nil
}

func (c *PluginContextClient) GetPluginRoot(isRelative bool) string {
	resp, err := c.hostClient.GetPluginRoot(context.Background(), &gen.GetPluginRootRequest{
		IsRelative: isRelative,
	})
	if err != nil {
		return ""
	}
	return resp.Path
}

func (c *PluginContextClient) GetMainWindowHandle() uintptr {
	return c.mainWindowHWND
}

func (c *PluginContextClient) Infof(template string, args ...any) { c.logger.Infof(template, args...) }
func (c *PluginContextClient) Debugf(template string, args ...any) {
	c.logger.Debugf(template, args...)
}
func (c *PluginContextClient) Warnf(template string, args ...any) { c.logger.Warnf(template, args...) }
func (c *PluginContextClient) Errorf(template string, args ...any) {
	c.logger.Errorf(template, args...)
}
func (c *PluginContextClient) GetLogger() dto.Logger { return c.logger }

func (c *PluginContextClient) PublishToFrontend(topic string, data []byte) error {
	_, err := c.hostClient.PublishToFrontend(context.Background(), &gen.PublishToFrontendRequest{
		Topic: topic,
		Data:  data,
	})
	return err
}

func (c *PluginContextClient) SubscribeFrontend(topic string) (<-chan []byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := c.hostClient.SubscribeFrontend(ctx, &gen.SubscribeFrontendRequest{Topic: topic})
	if err != nil {
		cancel()
		return nil, err
	}
	c.subCancelFuncs[topic] = cancel

	ch := make(chan []byte, 16)
	go func() {
		defer close(ch)
		for {
			msg, err := stream.Recv()
			if err == io.EOF || err != nil {
				return
			}
			select {
			case ch <- msg.Data:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func (c *PluginContextClient) UnsubscribeFrontend(topic string) error {
	if cancel, ok := c.subCancelFuncs[topic]; ok {
		cancel()
		delete(c.subCancelFuncs, topic)
	}
	_, err := c.hostClient.UnsubscribeFrontend(context.Background(), &gen.UnsubscribeFrontendRequest{Topic: topic})
	return err
}

// 能力族满足性编译期断言：PluginContextClient 实现全部能力族与复合接口
// （族定义见 dto/context.go——任一族增改方法而客户端未跟进，此处编译期即失败，
// 与 dto 包内 TestContextFamiliesPartition 的划分不变量互为表里）
var (
	_ dto.KVStore        = (*PluginContextClient)(nil)
	_ dto.Preference     = (*PluginContextClient)(nil)
	_ dto.Proxy          = (*PluginContextClient)(nil)
	_ dto.TaskTrigger    = (*PluginContextClient)(nil)
	_ dto.FrontendEvents = (*PluginContextClient)(nil)
	_ dto.Environment    = (*PluginContextClient)(nil)
	_ dto.LibraryQuery   = (*PluginContextClient)(nil)
	_ dto.LogSink        = (*PluginContextClient)(nil)
	_ dto.PluginContext  = (*PluginContextClient)(nil)
)
