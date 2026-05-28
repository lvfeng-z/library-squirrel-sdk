package dto

import "context"

// PluginDataProvider 插件数据持久化
type PluginDataProvider interface {
	GetPluginData(ctx context.Context) (string, error)
	SetPluginData(ctx context.Context, data string) error
	GetPluginRoot(ctx context.Context, isRelative bool) string
}

// SecureStorageProvider 加密存储
type SecureStorageProvider interface {
	StoreEncryptedValue(ctx context.Context, plainValue, description string) (string, error)
	GetDecryptedValue(ctx context.Context, storageKey string) (string, error)
	RemoveEncryptedValue(ctx context.Context, storageKey string) error
}

// WorkSetQueryProvider 作品集查询
type WorkSetQueryProvider interface {
	GetWorkSetBySiteWorkSetId(ctx context.Context, siteWorkSetId, siteName string) (*WorkSetDTO, error)
}

// SiteSaveProvider 站点保存
type SiteSaveProvider interface {
	AddSite(ctx context.Context, sites []*SiteDTO) error
}

// TaskCreateProvider 任务创建
type TaskCreateProvider interface {
	CreateTask(ctx context.Context, url string) (*CreateTaskResult, error)
}

// UrlListenerRegistry URL 监听器注册
type UrlListenerRegistry interface {
	RegisterUrlListener(ctx context.Context, contributionId string, patterns []string) error
	UnregisterUrlListener(ctx context.Context) error
}

// FrontendEventProvider 前后端事件桥接
type FrontendEventProvider interface {
	PublishToFrontend(topic string, data []byte) error
	SubscribeFrontend(topic string, pushCh func([]byte)) (cancel func(), err error)
	UnsubscribeFrontend(topic string) error
}
