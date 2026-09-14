package dto

// PluginContext 插件上下文，主程序提供给插件的完整 API
type PluginContext interface {
	// 扩展点注册
	RegisterTaskHandler(id string, name string, description string, handler TaskHandler) error
	RegisterSiteBrowser(id string, name string, description string, browser SiteBrowser) error

	// 扩展点注销
	UnregisterSiteBrowser(id string) error

	// 插件自存信息（统一 KV 存储，取代临时 plugin_data 与加密存储）
	GetValue(key string) (*StorageValue, error)
	SetValue(key string, value string) error
	SetValueEncrypted(key string, value string) error
	DeleteValue(key string) error
	GetAllValues() (map[string]*StorageValue, error)

	// 任务
	RegisterUrlListener(extensionId string, patterns []string) error
	UnregisterUrlListener(extensionId string) error
	CreateTask(url string) (*CreateTaskResult, error)

	// 前后端通信
	PublishToFrontend(topic string, data []byte) error
	SubscribeFrontend(topic string) (<-chan []byte, error)
	UnsubscribeFrontend(topic string) error

	// 路径
	GetPluginRoot(isRelative bool) string

	// 库查询（Tier 1 只读）：查询主程序库内已在的作品及周边数据，调用打到宿主 LibraryQuery service。
	// 默认只返回活数据；Get* 未命中返回 gRPC NotFound，GetWorkDir 未配置返回 FailedPrecondition；
	// StoreInfo.file_path 为库内相对路径（relPath 域，分隔符恒正斜杠，拼 OS 路径由插件自理）。
	// Get*/List* 以身份锚点为显式参数；Query*（无界集合，强制分页）以过滤契约消息为参数。
	// 作品
	GetWorkById(workId int64) (*WorkWithSite, error)
	GetWorkBySiteKey(siteKey string, siteWorkId string) (*WorkWithSite, error)
	QueryWorks(req *QueryWorksRequest) (*QueryWorksResponse, error)
	// 资源与 store
	ListResourcesByWorkId(workId int64) (*ListResourcesByWorkIdResponse, error)
	// 作者（本地轨 / 站点轨 / 作品关联）
	GetLocalAuthorById(localAuthorId int64) (*LocalAuthorDTO, error)
	QueryLocalAuthors(req *QueryLocalAuthorsRequest) (*QueryLocalAuthorsResponse, error)
	GetSiteAuthorBySiteKey(siteKey string, siteAuthorId string) (*SiteAuthorInfo, error)
	QuerySiteAuthors(req *QuerySiteAuthorsRequest) (*QuerySiteAuthorsResponse, error)
	ListAuthorsByWorkId(workId int64) (*ListAuthorsByWorkIdResponse, error)
	// 标签（与作者对称；作品关联含关联级 namespace 维度）
	GetLocalTagById(localTagId int64) (*LocalTagDTO, error)
	QueryLocalTags(req *QueryLocalTagsRequest) (*QueryLocalTagsResponse, error)
	GetSiteTagBySiteKey(siteKey string, siteTagId string) (*SiteTagInfo, error)
	QuerySiteTags(req *QuerySiteTagsRequest) (*QuerySiteTagsResponse, error)
	ListTagsByWorkId(workId int64) (*ListTagsByWorkIdResponse, error)
	// 作品集
	GetWorkSetById(workSetId int64) (*WorkSetDTO, error)
	GetWorkSetBySiteKey(siteKey string, siteWorkSetId string) (*WorkSetDTO, error)
	ListWorkSetsByWorkId(workId int64) (*ListWorkSetsByWorkIdResponse, error)
	ListParentWorkSets(workSetId int64) (*ListParentWorkSetsResponse, error)
	ListChildWorkSets(workSetId int64) (*ListChildWorkSetsResponse, error)
	// 站点（注册表投影，只读）
	ListSites() (*ListSitesResponse, error)
	// 工作目录（绝对路径；未配置返回 FailedPrecondition）
	GetWorkDir() (*GetWorkDirResponse, error)

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
