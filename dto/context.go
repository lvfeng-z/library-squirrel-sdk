package dto

// PluginContext 插件上下文，主程序提供给插件的完整 API。
//
// 按能力族组织：每个族子接口 = 一个能力族，族内方法共享同一契约、
// 降级纪律与文档类别（.dsh/rules/plugin.md「插件 SDK 能力边界」表的类别列
// 与族一一对应）；复合接口由全族嵌入组成，方法集与历史平铺形态完全一致——
// 既有插件与消费方零改动。新能力一律「建族或入族」，不再向复合接口直接添加
// 散方法（结构不变量由 context_families_test.go 锚定：族间两两不相交、复合
// 方法集恰为全族并集）。
//
// 消费方按需收窄：依赖声明为族类型（如 func NewCache(ctx KVStore)），
// 即在编译期排除对其他能力族的触碰。
type PluginContext interface {
	KVStore
	Preference
	Proxy
	TaskTrigger
	FrontendEvents
	Environment
	LibraryQuery
	LogSink
}

// KVStore 自存信息族（操作型）：统一 KV 持久化（plugin_storage 单表），
// 取代旧临时 plugin_data 与加密存储。明文项直接读写；加密项 SetValueEncrypted
// 存密文、读取自动解密；读取返回带 schemaVersion（宿主按清单 configSchemaVersion
// 盖戳，供插件配置迁移感知，见主仓 doc/plugin-dev-guide.md 8.3）
type KVStore interface {
	GetValue(key string) (*StorageValue, error)
	SetValue(key string, value string) error
	SetValueEncrypted(key string, value string) error
	DeleteValue(key string) error
	GetAllValues() (map[string]*StorageValue, error)
}

// Preference 偏好域族：插件经用户问答沉淀的决策记忆——与上面 KV 配置面
// 正交（边界判据：删掉它之后用户会被重新问吗？会 → 偏好域；不会 → settings/
// plugin_storage）。写入纪律：仅写经用户问答确认的决策；无删除方法——「忘掉」是
// 用户权利，删除仅经宿主记忆管理面，插件只能覆写不能销毁记忆
type Preference interface {
	GetPreference(key string) (*PreferenceValue, bool, error) // 无记录返回 (nil,false,nil) 不报错，调用方据此重新发起问答
	SetPreference(key string, value *PreferenceValue) error   // 整值覆写：同键已存在则整体重写
	ListMyPreferences() ([]string, error)                     // 本插件全部偏好键（跨插件键互不可见）
}

// Proxy 代理解析族（查询型，拉取式）：宿主三级检测（显式 > 系统代理 >
// 环境变量）逐请求现查，代理开关对下一请求即时生效。proxyURL 空 = 直连；
// source 为来源标签（explicit/system/env/none，仅日志排障用，不驱动分支）；
// 旧宿主（无此 RPC）返回 gRPC Unimplemented（SDK proxy 包自动降级）
type Proxy interface {
	ResolveProxy(explicitURL, requestURL string) (proxyURL, source string, err error)
}

// TaskTrigger 任务触发族（指令型）：向主程序提交 URL 创建任务（宿主路由到匹配插件）
type TaskTrigger interface {
	CreateTask(url string) (*CreateTaskResult, error)
}

// FrontendEvents 前端事件族（中继）：与前端双向 pub/sub（经宿主转发），
// topic 约定 plugin:{plugin-name}:{feature}:{action}
type FrontendEvents interface {
	PublishToFrontend(topic string, data []byte) error
	SubscribeFrontend(topic string) (<-chan []byte, error)
	UnsubscribeFrontend(topic string) error
}

// Environment 环境服务族（查询型）：宿主侧进程环境事实——插件根路径与主窗口句柄
type Environment interface {
	GetPluginRoot(isRelative bool) string
	GetMainWindowHandle() uintptr
}

// LibraryQuery 库查询族（查询型，Tier 1 只读）：查询主程序库内已在的作品及
// 周边数据，调用打到宿主 LibraryQuery service。默认只返回活数据；Get* 未命中返回
// gRPC NotFound，GetWorkDir 未配置返回 FailedPrecondition；StoreInfo.file_path 为库内
// 相对路径（relPath 域，分隔符恒正斜杠，拼 OS 路径由插件自理）。Get*/List* 以身份锚点
// 为显式参数；Query*（无界集合，强制分页）以过滤契约消息为参数
type LibraryQuery interface {
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
}

// LogSink 日志族（单向写）：写入主程序日志系统；GetLogger 返回可传递给子组件的 Logger
type LogSink interface {
	Infof(template string, args ...any)
	Debugf(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	GetLogger() Logger
}
