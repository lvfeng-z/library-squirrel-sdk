package pluginsdk

import "context"

// PluginHost 主程序侧实现的接口，插件通过 RPC 调用此接口方法
// 此接口在 plugin-sdk 中定义，供插件侧构造 RPC 调用使用
type PluginHost interface {
	// 扩展点注册（对应 ctx/registerTaskHandler）
	RegisterTaskHandler(ctx context.Context, id, name, description string, handler TaskHandler) error
	// 扩展点注册（对应 ctx/registerSiteBrowser）
	RegisterSiteBrowser(ctx context.Context, id, name, description string, browser SiteBrowser) error
	// 扩展点注销（对应 ctx/unregisterSiteBrowser）
	UnregisterSiteBrowser(ctx context.Context, id string) error
	// 插件数据（对应 ctx/getPluginData）
	GetPluginData(ctx context.Context) (string, error)
	// 插件数据（对应 ctx/setPluginData）
	SetPluginData(ctx context.Context, data string) error
	// 加密存储（对应 ctx/storeEncryptedValue）
	StoreEncryptedValue(ctx context.Context, plainValue, description string) (string, error)
	// 加密存储（对应 ctx/getDecryptedValue）
	GetDecryptedValue(ctx context.Context, storageKey string) (string, error)
	// 加密存储（对应 ctx/removeEncryptedValue）
	RemoveEncryptedValue(ctx context.Context, storageKey string) error
	// 业务查询（对应 ctx/getWorkSetBySiteWorkSetId）
	GetWorkSetBySiteWorkSetId(ctx context.Context, siteWorkSetId, siteName string) (*WorkSet, error)
	// 业务查询（对应 ctx/addSite）
	AddSite(ctx context.Context, sites []*Site) error
	// 任务（对应 ctx/registerUrlListener）
	RegisterUrlListener(ctx context.Context, contributionId string, patterns []string) error
	// 任务（对应 ctx/unregisterUrlListener）
	UnregisterUrlListener(ctx context.Context) error
	// 任务（对应 ctx/createTask）
	CreateTask(ctx context.Context, url string) (*TaskCreateResult, error)
	// 路径（对应 ctx/getPluginRoot）
	GetPluginRoot(ctx context.Context, isRelative bool) string
	// 日志（对应 ctx/infof）
	Infof(ctx context.Context, template string, args []any)
	// 日志（对应 ctx/debugf）
	Debugf(ctx context.Context, template string, args []any)
	// 日志（对应 ctx/warnf）
	Warnf(ctx context.Context, template string, args []any)
	// 日志（对应 ctx/errorf）
	Errorf(ctx context.Context, template string, args []any)
}

// NoOpPluginHost 空实现，用于单元测试
type NoOpPluginHost struct{}

func (NoOpPluginHost) RegisterTaskHandler(context.Context, string, string, string, TaskHandler) error { return nil }
func (NoOpPluginHost) RegisterSiteBrowser(context.Context, string, string, string, SiteBrowser) error    { return nil }
func (NoOpPluginHost) UnregisterSiteBrowser(context.Context, string) error                              { return nil }
func (NoOpPluginHost) GetPluginData(context.Context) (string, error)                                   { return "", nil }
func (NoOpPluginHost) SetPluginData(context.Context, string) error                                      { return nil }
func (NoOpPluginHost) StoreEncryptedValue(context.Context, string, string) (string, error)             { return "", nil }
func (NoOpPluginHost) GetDecryptedValue(context.Context, string) (string, error)                       { return "", nil }
func (NoOpPluginHost) RemoveEncryptedValue(context.Context, string) error                               { return nil }
func (NoOpPluginHost) GetWorkSetBySiteWorkSetId(context.Context, string, string) (*WorkSet, error)      { return nil, nil }
func (NoOpPluginHost) AddSite(context.Context, []*Site) error                                          { return nil }
func (NoOpPluginHost) RegisterUrlListener(context.Context, string, []string) error                     { return nil }
func (NoOpPluginHost) UnregisterUrlListener(context.Context) error                                       { return nil }
func (NoOpPluginHost) CreateTask(context.Context, string) (*TaskCreateResult, error)                   { return nil, nil }
func (NoOpPluginHost) GetPluginRoot(context.Context, bool) string                                      { return "" }
func (NoOpPluginHost) Infof(context.Context, string, []any)                                           {}
func (NoOpPluginHost) Debugf(context.Context, string, []any)                                          {}
func (NoOpPluginHost) Warnf(context.Context, string, []any)                                            {}
func (NoOpPluginHost) Errorf(context.Context, string, []any)                                           {}
