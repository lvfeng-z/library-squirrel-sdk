package transport

import (
	"context"
	"io"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"google.golang.org/grpc"
)

// PluginContextClient 插件侧的 PluginContext 实现，通过 gRPC 调用主程序的 HostService
type PluginContextClient struct {
	hostClient     gen.HostServiceClient
	logger         dto.Logger
	mainWindowHWND uintptr
	subCancelFuncs map[string]context.CancelFunc
}

// NewPluginContextClient 创建基于 gRPC 的 PluginContext 客户端
func NewPluginContextClient(conn *grpc.ClientConn) *PluginContextClient {
	return &PluginContextClient{
		hostClient:     gen.NewHostServiceClient(conn),
		logger:         NewGRPCLogger(gen.NewHostServiceClient(conn)),
		subCancelFuncs: make(map[string]context.CancelFunc),
	}
}

// SetMainWindowHandle 设置主窗口句柄
func (c *PluginContextClient) SetMainWindowHandle(hwnd uintptr) {
	c.mainWindowHWND = hwnd
}

func (c *PluginContextClient) RegisterTaskHandler(id, name, description string, handler dto.TaskHandler) error {
	_, err := c.hostClient.RegisterTaskHandler(context.Background(), &gen.RegisterExtensionRequest{
		ExtensionId: id,
		Name:           name,
		Description:    description,
	})
	return err
}

func (c *PluginContextClient) RegisterSiteBrowser(id, name, description string, browser dto.SiteBrowser) error {
	_, err := c.hostClient.RegisterSiteBrowser(context.Background(), &gen.RegisterExtensionRequest{
		ExtensionId: id,
		Name:           name,
		Description:    description,
	})
	return err
}

func (c *PluginContextClient) UnregisterSiteBrowser(id string) error {
	_, err := c.hostClient.UnregisterSiteBrowser(context.Background(), &gen.UnregisterRequest{
		ExtensionId: id,
	})
	return err
}

func (c *PluginContextClient) GetValue(key string) (string, error) {
	resp, err := c.hostClient.GetValue(context.Background(), &gen.StorageKeyRequest{Key: key})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
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

func (c *PluginContextClient) GetAllValues() (map[string]string, error) {
	resp, err := c.hostClient.GetAllValues(context.Background(), &gen.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

func (c *PluginContextClient) GetWorkSetBySiteWorkSetId(siteWorkSetId, siteName string) (*dto.WorkSetDTO, error) {
	resp, err := c.hostClient.GetWorkSetBySiteWorkSetId(context.Background(), &gen.WorkSetQueryRequest{
		SiteWorkSetId: siteWorkSetId,
		SiteName:      siteName,
	})
	if err != nil {
		return nil, err
	}
	return protoToWorkSet(resp.WorkSet), nil
}

func (c *PluginContextClient) AddSite(sites []*dto.SiteDTO) error {
	pbSites := make([]*gen.Site, len(sites))
	for i, s := range sites {
		pbSites[i] = &gen.Site{
			Id:              s.ID,
			CreateTime:      s.CreateTime,
			UpdateTime:      s.UpdateTime,
			SiteName:        s.SiteName,
			SiteDescription: s.SiteDescription,
			Homepage:        s.Homepage,
		}
	}
	_, err := c.hostClient.AddSite(context.Background(), &gen.AddSiteRequest{Sites: pbSites})
	return err
}

func (c *PluginContextClient) RegisterUrlListener(extensionId string, patterns []string) error {
	_, err := c.hostClient.RegisterUrlListener(context.Background(), &gen.UrlListenerRequest{
		ExtensionId: extensionId,
		Patterns:       patterns,
	})
	return err
}

func (c *PluginContextClient) UnregisterUrlListener(extensionId string) error {
	_, err := c.hostClient.UnregisterUrlListener(context.Background(), &gen.UnregisterRequest{
		ExtensionId: extensionId,
	})
	return err
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

// protoToWorkSet 将 proto WorkSet 转换为 WorkSetDTO
func protoToWorkSet(pb *gen.WorkSet) *dto.WorkSetDTO {
	if pb == nil {
		return nil
	}
	return &dto.WorkSetDTO{
		ID:                     pb.Id,
		CreateTime:             pb.CreateTime,
		UpdateTime:             pb.UpdateTime,
		SiteID:                 pb.SiteId,
		SiteWorkSetID:          pb.SiteWorkSetId,
		SiteWorkSetName:        pb.SiteWorkSetName,
		SiteAuthorID:           pb.SiteAuthorId,
		SiteWorkSetDescription: pb.SiteWorkSetDescription,
		SiteUploadTime:         pb.SiteUploadTime,
		SiteUpdateTime:         pb.SiteUpdateTime,
		NickName:               pb.NickName,
		LastView:               pb.LastView,
	}
}
