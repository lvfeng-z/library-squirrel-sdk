package transport

import (
	"context"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
)

// ===== Tier 1 库查询：PluginContextClient 的查询方法组 =====
// 经宿主侧 LibraryQuery service（GRPCBroker 反向连接）打到主程序。gRPC status 错误原样
// 上抛（Get* 未命中 NotFound、GetWorkDir 未配置 FailedPrecondition、宿主未实现该服务
// Unimplemented），插件以 google.golang.org/grpc/status 判码。
// 响应消息经 dto 别名直通（别名即 gen 类型），无二次映射。

// ===== 作品 =====

func (c *PluginContextClient) GetWorkById(workId int64) (*dto.WorkWithSite, error) {
	return c.queryClient.GetWorkById(context.Background(), &gen.GetWorkByIdRequest{WorkId: workId}, callOpts...)
}

func (c *PluginContextClient) GetWorkBySiteKey(siteKey string, siteWorkId string) (*dto.WorkWithSite, error) {
	return c.queryClient.GetWorkBySiteKey(context.Background(), &gen.GetWorkBySiteKeyRequest{
		SiteKey:    siteKey,
		SiteWorkId: siteWorkId,
	}, callOpts...)
}

func (c *PluginContextClient) QueryWorks(req *dto.QueryWorksRequest) (*dto.QueryWorksResponse, error) {
	return c.queryClient.QueryWorks(context.Background(), req, callOpts...)
}

// ===== 资源与 store =====

func (c *PluginContextClient) ListResourcesByWorkId(workId int64) (*dto.ListResourcesByWorkIdResponse, error) {
	return c.queryClient.ListResourcesByWorkId(context.Background(), &gen.ListResourcesByWorkIdRequest{WorkId: workId}, callOpts...)
}

// ===== 作者（本地轨 / 站点轨 / 作品关联，作品关联含关联级 role 维度）=====

func (c *PluginContextClient) GetLocalAuthorById(localAuthorId int64) (*dto.LocalAuthorDTO, error) {
	return c.queryClient.GetLocalAuthorById(context.Background(), &gen.GetLocalAuthorByIdRequest{LocalAuthorId: localAuthorId}, callOpts...)
}

func (c *PluginContextClient) QueryLocalAuthors(req *dto.QueryLocalAuthorsRequest) (*dto.QueryLocalAuthorsResponse, error) {
	return c.queryClient.QueryLocalAuthors(context.Background(), req, callOpts...)
}

func (c *PluginContextClient) GetSiteAuthorBySiteKey(siteKey string, siteAuthorId string) (*dto.SiteAuthorInfo, error) {
	return c.queryClient.GetSiteAuthorBySiteKey(context.Background(), &gen.GetSiteAuthorBySiteKeyRequest{
		SiteKey:      siteKey,
		SiteAuthorId: siteAuthorId,
	}, callOpts...)
}

func (c *PluginContextClient) QuerySiteAuthors(req *dto.QuerySiteAuthorsRequest) (*dto.QuerySiteAuthorsResponse, error) {
	return c.queryClient.QuerySiteAuthors(context.Background(), req, callOpts...)
}

func (c *PluginContextClient) ListAuthorsByWorkId(workId int64) (*dto.ListAuthorsByWorkIdResponse, error) {
	return c.queryClient.ListAuthorsByWorkId(context.Background(), &gen.ListAuthorsByWorkIdRequest{WorkId: workId}, callOpts...)
}

// ===== 标签（与作者对称；作品关联含关联级 namespace 维度）=====

func (c *PluginContextClient) GetLocalTagById(localTagId int64) (*dto.LocalTagDTO, error) {
	return c.queryClient.GetLocalTagById(context.Background(), &gen.GetLocalTagByIdRequest{LocalTagId: localTagId}, callOpts...)
}

func (c *PluginContextClient) QueryLocalTags(req *dto.QueryLocalTagsRequest) (*dto.QueryLocalTagsResponse, error) {
	return c.queryClient.QueryLocalTags(context.Background(), req, callOpts...)
}

func (c *PluginContextClient) GetSiteTagBySiteKey(siteKey string, siteTagId string) (*dto.SiteTagInfo, error) {
	return c.queryClient.GetSiteTagBySiteKey(context.Background(), &gen.GetSiteTagBySiteKeyRequest{
		SiteKey:   siteKey,
		SiteTagId: siteTagId,
	}, callOpts...)
}

func (c *PluginContextClient) QuerySiteTags(req *dto.QuerySiteTagsRequest) (*dto.QuerySiteTagsResponse, error) {
	return c.queryClient.QuerySiteTags(context.Background(), req, callOpts...)
}

func (c *PluginContextClient) ListTagsByWorkId(workId int64) (*dto.ListTagsByWorkIdResponse, error) {
	return c.queryClient.ListTagsByWorkId(context.Background(), &gen.ListTagsByWorkIdRequest{WorkId: workId}, callOpts...)
}

// ===== 作品集 =====

func (c *PluginContextClient) GetWorkSetById(workSetId int64) (*dto.WorkSetDTO, error) {
	return c.queryClient.GetWorkSetById(context.Background(), &gen.GetWorkSetByIdRequest{WorkSetId: workSetId}, callOpts...)
}

func (c *PluginContextClient) GetWorkSetBySiteKey(siteKey string, siteWorkSetId string) (*dto.WorkSetDTO, error) {
	return c.queryClient.GetWorkSetBySiteKey(context.Background(), &gen.GetWorkSetBySiteKeyRequest{
		SiteKey:       siteKey,
		SiteWorkSetId: siteWorkSetId,
	}, callOpts...)
}

func (c *PluginContextClient) ListWorkSetsByWorkId(workId int64) (*dto.ListWorkSetsByWorkIdResponse, error) {
	return c.queryClient.ListWorkSetsByWorkId(context.Background(), &gen.ListWorkSetsByWorkIdRequest{WorkId: workId}, callOpts...)
}

func (c *PluginContextClient) ListParentWorkSets(workSetId int64) (*dto.ListParentWorkSetsResponse, error) {
	return c.queryClient.ListParentWorkSets(context.Background(), &gen.ListParentWorkSetsRequest{WorkSetId: workSetId}, callOpts...)
}

func (c *PluginContextClient) ListChildWorkSets(workSetId int64) (*dto.ListChildWorkSetsResponse, error) {
	return c.queryClient.ListChildWorkSets(context.Background(), &gen.ListChildWorkSetsRequest{WorkSetId: workSetId}, callOpts...)
}

// QueryLocalWorkSets 库内作品集名全集分页（互证索引用）
func (c *PluginContextClient) QueryLocalWorkSets(req *dto.QueryLocalWorkSetsRequest) (*dto.QueryLocalWorkSetsResponse, error) {
	return c.queryClient.QueryLocalWorkSets(context.Background(), req, callOpts...)
}

// ===== 站点（注册表投影，只读）=====

func (c *PluginContextClient) ListSites() (*dto.ListSitesResponse, error) {
	return c.queryClient.ListSites(context.Background(), &gen.Empty{}, callOpts...)
}

// ===== 工作目录 =====

func (c *PluginContextClient) GetWorkDir() (*dto.GetWorkDirResponse, error) {
	return c.queryClient.GetWorkDir(context.Background(), &gen.Empty{}, callOpts...)
}
