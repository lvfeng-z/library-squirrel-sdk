package dto

// PluginContext 插件上下文，主程序提供给插件的完整 API
type PluginContext interface {
	// 插件自存信息（统一 KV 存储，取代临时 plugin_data 与加密存储）
	GetValue(key string) (*StorageValue, error)
	SetValue(key string, value string) error
	SetValueEncrypted(key string, value string) error
	DeleteValue(key string) error
	GetAllValues() (map[string]*StorageValue, error)

	// 用户决策偏好（偏好域）：插件经用户问答沉淀的决策记忆——与上面统一 KV 的配置面
	// 正交（边界判据：删掉它之后用户会被重新问吗？会 → 偏好域；不会 → settings/
	// plugin_storage）。写入纪律：仅写经用户问答确认的决策；无删除方法——「忘掉」是
	// 用户权利，删除仅经宿主记忆管理面，插件只能覆写不能销毁记忆。
	GetPreference(key string) (*PreferenceValue, bool, error) // 无记录返回 (nil,false,nil) 不报错，调用方据此重新发起问答
	SetPreference(key string, value *PreferenceValue) error   // 整值覆写：同键已存在则整体重写
	ListMyPreferences() ([]string, error)                     // 本插件全部偏好键（跨插件键互不可见）

	// 代理解析：宿主三级检测（显式 > 系统代理 > 环境变量）逐请求现查，代理开关对
	// 下一请求即时生效。proxyURL 空 = 直连；source 为来源标签（explicit/system/env/none，
	// 仅日志排障用，不驱动分支）；旧宿主（无此 RPC）返回 gRPC Unimplemented
	ResolveProxy(explicitURL, requestURL string) (proxyURL, source string, err error)

	// 任务
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
	// 作者（本地轨 / 站点轨 / 作品关联，作品关联含关联级 role 维度）
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
