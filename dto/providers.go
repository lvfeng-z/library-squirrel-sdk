package dto

import "context"

// StorageProvider 插件自存信息（统一 KV 存储，取代临时 plugin_data 与加密存储）
type StorageProvider interface {
	GetValue(ctx context.Context, key string) (*StorageValue, error)
	SetValue(ctx context.Context, key, value string) error
	SetValueEncrypted(ctx context.Context, key, value string) error
	DeleteValue(ctx context.Context, key string) error
	GetAllValues(ctx context.Context) (map[string]*StorageValue, error)
}

// PluginRootProvider 插件根路径
type PluginRootProvider interface {
	GetPluginRoot(ctx context.Context, isRelative bool) string
}

// StorePathQueryProvider 资源 store 路径查询(主程序据 task+role+store_seq 查真实落盘路径)
type StorePathQueryProvider interface {
	GetStoreRelPath(ctx context.Context, taskId int64, role string, storeSeq int) (string, error)
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
	RegisterUrlListener(ctx context.Context, extensionId string, patterns []string) error
	UnregisterUrlListener(ctx context.Context, extensionId string) error
}

// FrontendEventProvider 前后端事件桥接
type FrontendEventProvider interface {
	PublishToFrontend(topic string, data []byte) error
	SubscribeFrontend(topic string, pushCh func([]byte)) (cancel func(), err error)
	UnsubscribeFrontend(topic string) error
}
