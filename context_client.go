package pluginsdk

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/lvfeng-z/library-squirrel-plugin-sdk/gen"
)

// PluginContextClient 插件侧的 PluginContext 实现，通过 gRPC 调用主程序的 HostService
type PluginContextClient struct {
	hostClient gen.HostServiceClient
	logger     Logger
}

// NewPluginContextClient 创建基于 gRPC 的 PluginContext 客户端
func NewPluginContextClient(conn *grpc.ClientConn) *PluginContextClient {
	return &PluginContextClient{
		hostClient: gen.NewHostServiceClient(conn),
		logger:     NewGRPCLogger(gen.NewHostServiceClient(conn)),
	}
}

func (c *PluginContextClient) RegisterTaskHandler(id, name, description string, handler TaskHandler) error {
	_, err := c.hostClient.RegisterTaskHandler(context.Background(), &gen.RegisterExtensionRequest{
		ContributionId: id,
		Name:           name,
		Description:    description,
	})
	return err
}

func (c *PluginContextClient) RegisterSiteBrowser(id, name, description string, browser SiteBrowser) error {
	_, err := c.hostClient.RegisterSiteBrowser(context.Background(), &gen.RegisterExtensionRequest{
		ContributionId: id,
		Name:           name,
		Description:    description,
	})
	return err
}

func (c *PluginContextClient) UnregisterSiteBrowser(id string) error {
	_, err := c.hostClient.UnregisterSiteBrowser(context.Background(), &gen.UnregisterRequest{
		ContributionId: id,
	})
	return err
}

func (c *PluginContextClient) GetPluginData() (string, error) {
	resp, err := c.hostClient.GetPluginData(context.Background(), &gen.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Data, nil
}

func (c *PluginContextClient) SetPluginData(data string) error {
	_, err := c.hostClient.SetPluginData(context.Background(), &gen.PluginDataRequest{Data: data})
	return err
}

func (c *PluginContextClient) StoreEncryptedValue(plainValue, description string) (string, error) {
	resp, err := c.hostClient.StoreEncryptedValue(context.Background(), &gen.EncryptRequest{
		PlainValue:  plainValue,
		Description: description,
	})
	if err != nil {
		return "", err
	}
	return resp.Key, nil
}

func (c *PluginContextClient) GetDecryptedValue(storageKey string) (string, error) {
	resp, err := c.hostClient.GetDecryptedValue(context.Background(), &gen.DecryptRequest{
		StorageKey: storageKey,
	})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func (c *PluginContextClient) RemoveEncryptedValue(storageKey string) error {
	_, err := c.hostClient.RemoveEncryptedValue(context.Background(), &gen.DecryptRequest{
		StorageKey: storageKey,
	})
	return err
}

func (c *PluginContextClient) GetWorkSetBySiteWorkSetId(siteWorkSetId, siteName string) (*WorkSet, error) {
	resp, err := c.hostClient.GetWorkSetBySiteWorkSetId(context.Background(), &gen.WorkSetQueryRequest{
		SiteWorkSetId: siteWorkSetId,
		SiteName:      siteName,
	})
	if err != nil {
		return nil, err
	}
	return protoToWorkSet(resp.WorkSet), nil
}

func (c *PluginContextClient) AddSite(sites []*Site) error {
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

func (c *PluginContextClient) RegisterUrlListener(contributionId string, patterns []string) error {
	_, err := c.hostClient.RegisterUrlListener(context.Background(), &gen.UrlListenerRequest{
		ContributionId: contributionId,
		Patterns:       patterns,
	})
	return err
}

func (c *PluginContextClient) UnregisterUrlListener() error {
	_, err := c.hostClient.UnregisterUrlListener(context.Background(), &gen.Empty{})
	return err
}

func (c *PluginContextClient) CreateTask(url string) (*TaskCreateResult, error) {
	resp, err := c.hostClient.CreateTask(context.Background(), &gen.CreateTaskRequest{Url: url})
	if err != nil {
		return nil, err
	}
	return &TaskCreateResult{
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

func (c *PluginContextClient) GetMainWindow() WindowHandle {
	return nil
}

func (c *PluginContextClient) CreateWindow(options WindowOptions) (WindowHandle, error) {
	return nil, fmt.Errorf("window management not supported in subprocess mode")
}

func (c *PluginContextClient) Infof(template string, args ...any)   { c.logger.Infof(template, args...) }
func (c *PluginContextClient) Debugf(template string, args ...any)  { c.logger.Debugf(template, args...) }
func (c *PluginContextClient) Warnf(template string, args ...any)   { c.logger.Warnf(template, args...) }
func (c *PluginContextClient) Errorf(template string, args ...any)  { c.logger.Errorf(template, args...) }
func (c *PluginContextClient) GetLogger() Logger { return c.logger }

// protoToWorkSet 将 proto WorkSet 转换为 SDK WorkSet
func protoToWorkSet(pb *gen.WorkSet) *WorkSet {
	if pb == nil {
		return nil
	}
	return &WorkSet{
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
