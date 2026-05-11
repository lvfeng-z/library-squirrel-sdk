package pluginsdk

import (
	"fmt"
	"sync/atomic"
	"time"
)

// PluginContextClient 插件侧的 PluginContext 实现，通过 RPC 与主进程通信
// 实现 PluginContext 接口，每个方法映射为一次 JSON-RPC 调用
type PluginContextClient struct {
	client   *RPCClient
	init     *PluginContextInit
	streamID atomic.Int64
	logger   Logger
}

// NewPluginContextClient 创建 PluginContext 客户端
// init 包含插件的初始化参数
func NewPluginContextClient(client *RPCClient, init *PluginContextInit) *PluginContextClient {
	return &PluginContextClient{
		client: client,
		init:   init,
		logger: NewPluginLogger(client, ""),
	}
}

func (c *PluginContextClient) RegisterTaskHandler(id, name, description string, handler TaskHandler) error {
	type params struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	return c.client.Call("ctx/registerTaskHandler", params{ID: id, Name: name, Description: description}, nil)
}

func (c *PluginContextClient) RegisterSiteBrowser(id, name, description string, browser SiteBrowser) error {
	type params struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	return c.client.Call("ctx/registerSiteBrowser", params{ID: id, Name: name, Description: description}, nil)
}

func (c *PluginContextClient) UnregisterSiteBrowser(id string) error {
	type params struct {
		ID string `json:"id"`
	}
	return c.client.Call("ctx/unregisterSiteBrowser", params{ID: id}, nil)
}

func (c *PluginContextClient) GetPluginData() (string, error) {
	type result struct {
		Data string `json:"data"`
	}
	var r result
	if err := c.client.Call("ctx/getPluginData", nil, &r); err != nil {
		return "", err
	}
	return r.Data, nil
}

func (c *PluginContextClient) SetPluginData(data string) error {
	type params struct {
		Data string `json:"data"`
	}
	return c.client.Call("ctx/setPluginData", params{Data: data}, nil)
}

func (c *PluginContextClient) StoreEncryptedValue(plainValue, description string) (string, error) {
	type params struct {
		PlainValue   string `json:"plainValue"`
		Description string `json:"description"`
	}
	type result struct {
		Key string `json:"key"`
	}
	var r result
	if err := c.client.Call("ctx/storeEncryptedValue", params{PlainValue: plainValue, Description: description}, &r); err != nil {
		return "", err
	}
	return r.Key, nil
}

func (c *PluginContextClient) GetDecryptedValue(storageKey string) (string, error) {
	type params struct {
		StorageKey string `json:"storageKey"`
	}
	type result struct {
		Value string `json:"value"`
	}
	var r result
	if err := c.client.Call("ctx/getDecryptedValue", params{StorageKey: storageKey}, &r); err != nil {
		return "", err
	}
	return r.Value, nil
}

func (c *PluginContextClient) RemoveEncryptedValue(storageKey string) error {
	type params struct {
		StorageKey string `json:"storageKey"`
	}
	return c.client.Call("ctx/removeEncryptedValue", params{StorageKey: storageKey}, nil)
}

func (c *PluginContextClient) GetWorkSetBySiteWorkSetId(siteWorkSetId, siteName string) (*WorkSet, error) {
	type params struct {
		SiteWorkSetID string `json:"siteWorkSetId"`
		SiteName      string `json:"siteName"`
	}
	var result *WorkSet
	if err := c.client.Call("ctx/getWorkSetBySiteWorkSetId", params{SiteWorkSetID: siteWorkSetId, SiteName: siteName}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *PluginContextClient) AddSite(sites []*Site) error {
	type params struct {
		Sites []*Site `json:"sites"`
	}
	return c.client.Call("ctx/addSite", params{Sites: sites}, nil)
}

func (c *PluginContextClient) RegisterUrlListener(contributionId string, patterns []string) error {
	type params struct {
		ContributionID string   `json:"contributionId"`
		Patterns       []string `json:"patterns"`
	}
	return c.client.Call("ctx/registerUrlListener", params{ContributionID: contributionId, Patterns: patterns}, nil)
}

func (c *PluginContextClient) UnregisterUrlListener() error {
	return c.client.Call("ctx/unregisterUrlListener", nil, nil)
}

func (c *PluginContextClient) CreateTask(url string) (*TaskCreateResult, error) {
	type params struct {
		URL string `json:"url"`
	}
	var result *TaskCreateResult
	if err := c.client.Call("ctx/createTask", params{URL: url}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *PluginContextClient) GetPluginRoot(isRelative bool) string {
	type params struct {
		IsRelative bool `json:"isRelative"`
	}
	type result struct {
		Path string `json:"path"`
	}
	var r result
	if err := c.client.Call("ctx/getPluginRoot", params{IsRelative: isRelative}, &r); err != nil {
		return ""
	}
	return r.Path
}

func (c *PluginContextClient) Infof(template string, args ...any)   { c.logger.Infof(template, args...) }
func (c *PluginContextClient) Debugf(template string, args ...any)  { c.logger.Debugf(template, args...) }
func (c *PluginContextClient) Warnf(template string, args ...any)   { c.logger.Warnf(template, args...) }
func (c *PluginContextClient) Errorf(template string, args ...any)  { c.logger.Errorf(template, args...) }

func (c *PluginContextClient) GetLogger() Logger { return c.logger }

func (c *PluginContextClient) GetMainWindow() WindowHandle {
	// 子进程模式不支持窗口管理
	return nil
}

func (c *PluginContextClient) CreateWindow(options WindowOptions) (WindowHandle, error) {
	return nil, fmt.Errorf("window management not supported in subprocess mode")
}

// NextStreamID 生成下一个流 ID
func (c *PluginContextClient) NextStreamID() string {
	id := c.streamID.Add(1)
	return fmt.Sprintf("stream-%d-%d", time.Now().UnixNano(), id)
}
