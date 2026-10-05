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

// TaskCreateProvider 任务创建
type TaskCreateProvider interface {
	CreateTask(ctx context.Context, url string) (*CreateTaskResult, error)
}

// FrontendEventProvider 前后端事件桥接
type FrontendEventProvider interface {
	PublishToFrontend(topic string, data []byte) error
	SubscribeFrontend(topic string, pushCh func([]byte)) (cancel func(), err error)
	UnsubscribeFrontend(topic string) error
}

// PreferenceProvider 用户决策偏好域（插件经用户问答沉淀的用户决策记忆，与统一 KV
// 配置面正交）。读（无记录 ok=false 不报错）/ 整值覆写 / 列自身键三能力；无删除——
//「忘掉」是用户权利，删除仅经宿主记忆管理面
type PreferenceProvider interface {
	GetPreference(ctx context.Context, key string) (*PreferenceValue, bool, error)
	SetPreference(ctx context.Context, key string, value *PreferenceValue) error
	ListMyPreferences(ctx context.Context) ([]string, error)
}
