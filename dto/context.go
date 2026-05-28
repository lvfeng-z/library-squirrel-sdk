package dto

// PluginContext 插件上下文，主程序提供给插件的完整 API
type PluginContext interface {
	// 扩展点注册
	RegisterTaskHandler(id string, name string, description string, handler TaskHandler) error
	RegisterSiteBrowser(id string, name string, description string, browser SiteBrowser) error

	// 扩展点注销
	UnregisterSiteBrowser(id string) error

	// 插件数据持久化
	GetPluginData() (string, error)
	SetPluginData(data string) error

	// 加密存储
	StoreEncryptedValue(plainValue string, description string) (string, error)
	GetDecryptedValue(storageKey string) (string, error)
	RemoveEncryptedValue(storageKey string) error

	// 业务查询
	GetWorkSetBySiteWorkSetId(siteWorkSetId string, siteName string) (*WorkSetDTO, error)
	AddSite(sites []*SiteDTO) error

	// 任务
	RegisterUrlListener(contributionId string, patterns []string) error
	UnregisterUrlListener() error
	CreateTask(url string) (*CreateTaskResult, error)

	// 前后端通信
	PublishToFrontend(topic string, data []byte) error
	SubscribeFrontend(topic string) (<-chan []byte, error)
	UnsubscribeFrontend(topic string) error

	// 路径
	GetPluginRoot(isRelative bool) string

	// 窗口
	GetMainWindowHandle() uintptr

	// 日志
	Infof(template string, args ...any)
	Debugf(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)

	// 获取可传递给子组件的 Logger
	GetLogger() Logger
}
