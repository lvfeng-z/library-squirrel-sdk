package dto

import (
	"context"

	"github.com/lvfeng-z/library-squirrel-sdk/gen"
)

// ===== Tier 1 库查询契约（LibraryQuery 服务）消息别名 =====

// PageRequest 无界集合查询（Query* 族）的分页参数（别名 gen.PageRequest，proto 单源）
type PageRequest = gen.PageRequest

// PageInfo 分页回执：过滤后总数 + 实际生效的页码/页大小（别名 gen.PageInfo，proto 单源）
type PageInfo = gen.PageInfo

// WorkWithSite 作品 + 归属站点并排组合：Work 只带库内 site_id，复合身份键 (site_key, site_work_id) 经 SiteDTO 补全（别名 gen.WorkWithSite，proto 单源）
type WorkWithSite = gen.WorkWithSite

// ResourceInfo 资源摘要（含活行 store 摘要列表）（别名 gen.ResourceInfo，proto 单源）
type ResourceInfo = gen.ResourceInfo

// StoreInfo 单个活行 store 的文件摘要（file_path 为 relPath 域，分隔符恒正斜杠）（别名 gen.StoreInfo，proto 单源）
type StoreInfo = gen.StoreInfo

// SiteAuthorInfo 站点作者查询结果：字段面对齐任务声明期 TaskSiteAuthorDTO，另携库内行 id 与身份键锚点（别名 gen.SiteAuthorInfo，proto 单源）
type SiteAuthorInfo = gen.SiteAuthorInfo

// SiteTagInfo 站点标签查询结果：字段面对齐任务声明期 TaskSiteTagDTO，另携库内行 id 与身份键锚点（别名 gen.SiteTagInfo，proto 单源）
type SiteTagInfo = gen.SiteTagInfo

// WorkLocalTagEntry 作品的本地标签关联条目（标签 + 关联级 namespace）（别名 gen.WorkLocalTagEntry，proto 单源）
type WorkLocalTagEntry = gen.WorkLocalTagEntry

// WorkSiteTagEntry 作品的站点标签关联条目（标签 + 关联级 namespace 镜像）（别名 gen.WorkSiteTagEntry，proto 单源）
type WorkSiteTagEntry = gen.WorkSiteTagEntry

// QueryWorksResponse 作品分页查询结果（别名 gen.QueryWorksResponse，proto 单源）
type QueryWorksResponse = gen.QueryWorksResponse

// ListResourcesByWorkIdResponse 作品的资源全集（按归属锚定，不分页）（别名 gen.ListResourcesByWorkIdResponse，proto 单源）
type ListResourcesByWorkIdResponse = gen.ListResourcesByWorkIdResponse

// QueryLocalAuthorsResponse 本地作者分页查询结果（别名 gen.QueryLocalAuthorsResponse，proto 单源）
type QueryLocalAuthorsResponse = gen.QueryLocalAuthorsResponse

// QuerySiteAuthorsResponse 站点作者分页查询结果（别名 gen.QuerySiteAuthorsResponse，proto 单源）
type QuerySiteAuthorsResponse = gen.QuerySiteAuthorsResponse

// ListAuthorsByWorkIdResponse 作品的作者关联全集（local/site 两轨）（别名 gen.ListAuthorsByWorkIdResponse，proto 单源）
type ListAuthorsByWorkIdResponse = gen.ListAuthorsByWorkIdResponse

// QueryLocalTagsResponse 本地标签分页查询结果（别名 gen.QueryLocalTagsResponse，proto 单源）
type QueryLocalTagsResponse = gen.QueryLocalTagsResponse

// QuerySiteTagsResponse 站点标签分页查询结果（别名 gen.QuerySiteTagsResponse，proto 单源）
type QuerySiteTagsResponse = gen.QuerySiteTagsResponse

// ListTagsByWorkIdResponse 作品的标签关联全集（local/site 两轨，含关联级 namespace）（别名 gen.ListTagsByWorkIdResponse，proto 单源）
type ListTagsByWorkIdResponse = gen.ListTagsByWorkIdResponse

// ListWorkSetsByWorkIdResponse 作品归属的作品集全集（别名 gen.ListWorkSetsByWorkIdResponse，proto 单源）
type ListWorkSetsByWorkIdResponse = gen.ListWorkSetsByWorkIdResponse

// ListParentWorkSetsResponse 作品集的父集全集（别名 gen.ListParentWorkSetsResponse，proto 单源）
type ListParentWorkSetsResponse = gen.ListParentWorkSetsResponse

// ListChildWorkSetsResponse 作品集的子集全集（别名 gen.ListChildWorkSetsResponse，proto 单源）
type ListChildWorkSetsResponse = gen.ListChildWorkSetsResponse

// ListSitesResponse 站点注册表投影全集（别名 gen.ListSitesResponse，proto 单源）
type ListSitesResponse = gen.ListSitesResponse

// GetWorkDirResponse 资源库根目录绝对路径；未配置时 RPC 返回 FailedPrecondition（别名 gen.GetWorkDirResponse，proto 单源）
type GetWorkDirResponse = gen.GetWorkDirResponse

// ===== Query* 族请求别名（PluginContext 查询方法组的过滤契约载体）=====

// QueryWorksRequest 作品过滤查询请求（站点键/名称/作者/标签/入库时间过滤 + 分页）（别名 gen.QueryWorksRequest，proto 单源）
type QueryWorksRequest = gen.QueryWorksRequest

// QueryLocalAuthorsRequest 本地作者过滤查询请求（名称模糊 + 分页）（别名 gen.QueryLocalAuthorsRequest，proto 单源）
type QueryLocalAuthorsRequest = gen.QueryLocalAuthorsRequest

// QueryLocalTagsRequest 本地标签过滤查询请求（名称模糊 + 分页）（别名 gen.QueryLocalTagsRequest，proto 单源）
type QueryLocalTagsRequest = gen.QueryLocalTagsRequest

// QuerySiteAuthorsRequest 站点作者过滤查询请求（站点键精确 + 名称模糊 + 分页）（别名 gen.QuerySiteAuthorsRequest，proto 单源）
type QuerySiteAuthorsRequest = gen.QuerySiteAuthorsRequest

// QuerySiteTagsRequest 站点标签过滤查询请求（站点键精确 + 名称模糊 + 分页）（别名 gen.QuerySiteTagsRequest，proto 单源）
type QuerySiteTagsRequest = gen.QuerySiteTagsRequest

// LibraryQueryProvider 宿主侧 Tier 1 库查询实现契约：主程序实现本接口并注入 HostDeps，
// 插件经 PluginContext 查询方法组发起的 gRPC 调用最终到达此处。方法面与
// gen.LibraryQueryServer 逐方法一致（请求/响应消息即 proto 单源类型，无二次映射）；
// 错误以 gRPC status 携带语义码：Get* 未命中 NotFound、GetWorkDir 未配置 FailedPrecondition
//（见 proto LibraryQuery 服务注释）。
type LibraryQueryProvider interface {
	// 作品
	GetWorkById(ctx context.Context, req *gen.GetWorkByIdRequest) (*gen.WorkWithSite, error)
	GetWorkBySiteKey(ctx context.Context, req *gen.GetWorkBySiteKeyRequest) (*gen.WorkWithSite, error)
	QueryWorks(ctx context.Context, req *QueryWorksRequest) (*QueryWorksResponse, error)

	// 资源与 store
	ListResourcesByWorkId(ctx context.Context, req *gen.ListResourcesByWorkIdRequest) (*ListResourcesByWorkIdResponse, error)

	// 作者（本地轨 / 站点轨 / 作品关联）
	GetLocalAuthorById(ctx context.Context, req *gen.GetLocalAuthorByIdRequest) (*gen.LocalAuthorDTO, error)
	QueryLocalAuthors(ctx context.Context, req *QueryLocalAuthorsRequest) (*QueryLocalAuthorsResponse, error)
	GetSiteAuthorBySiteKey(ctx context.Context, req *gen.GetSiteAuthorBySiteKeyRequest) (*gen.SiteAuthorInfo, error)
	QuerySiteAuthors(ctx context.Context, req *QuerySiteAuthorsRequest) (*QuerySiteAuthorsResponse, error)
	ListAuthorsByWorkId(ctx context.Context, req *gen.ListAuthorsByWorkIdRequest) (*gen.ListAuthorsByWorkIdResponse, error)

	// 标签（与作者对称；作品关联含关联级 namespace 维度）
	GetLocalTagById(ctx context.Context, req *gen.GetLocalTagByIdRequest) (*gen.LocalTagDTO, error)
	QueryLocalTags(ctx context.Context, req *gen.QueryLocalTagsRequest) (*gen.QueryLocalTagsResponse, error)
	GetSiteTagBySiteKey(ctx context.Context, req *gen.GetSiteTagBySiteKeyRequest) (*gen.SiteTagInfo, error)
	QuerySiteTags(ctx context.Context, req *QuerySiteTagsRequest) (*QuerySiteTagsResponse, error)
	ListTagsByWorkId(ctx context.Context, req *gen.ListTagsByWorkIdRequest) (*gen.ListTagsByWorkIdResponse, error)

	// 作品集
	GetWorkSetById(ctx context.Context, req *gen.GetWorkSetByIdRequest) (*gen.WorkSet, error)
	GetWorkSetBySiteKey(ctx context.Context, req *gen.GetWorkSetBySiteKeyRequest) (*gen.WorkSet, error)
	ListWorkSetsByWorkId(ctx context.Context, req *gen.ListWorkSetsByWorkIdRequest) (*gen.ListWorkSetsByWorkIdResponse, error)
	ListParentWorkSets(ctx context.Context, req *gen.ListParentWorkSetsRequest) (*gen.ListParentWorkSetsResponse, error)
	ListChildWorkSets(ctx context.Context, req *gen.ListChildWorkSetsRequest) (*gen.ListChildWorkSetsResponse, error)

	// 站点（注册表投影，只读）
	ListSites(ctx context.Context, req *gen.Empty) (*gen.ListSitesResponse, error)

	// 工作目录
	GetWorkDir(ctx context.Context, req *gen.Empty) (*gen.GetWorkDirResponse, error)
}
