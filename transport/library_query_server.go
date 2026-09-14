package transport

import (
	"context"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
)

// libraryQueryServer 宿主侧 LibraryQuery 的 gRPC 服务端：逐方法委托 HostDeps 注入的
// dto.LibraryQueryProvider（主程序实现）。错误原样透传——provider 以 gRPC status
// 携带语义码（Get* 未命中 NotFound、GetWorkDir 未配置 FailedPrecondition）
type libraryQueryServer struct {
	gen.UnimplementedLibraryQueryServer
	provider dto.LibraryQueryProvider
}

// ===== 作品 =====

func (s *libraryQueryServer) GetWorkById(ctx context.Context, req *gen.GetWorkByIdRequest) (*gen.WorkWithSite, error) {
	return s.provider.GetWorkById(ctx, req)
}

func (s *libraryQueryServer) GetWorkBySiteKey(ctx context.Context, req *gen.GetWorkBySiteKeyRequest) (*gen.WorkWithSite, error) {
	return s.provider.GetWorkBySiteKey(ctx, req)
}

func (s *libraryQueryServer) QueryWorks(ctx context.Context, req *gen.QueryWorksRequest) (*gen.QueryWorksResponse, error) {
	return s.provider.QueryWorks(ctx, req)
}

// ===== 资源与 store =====

func (s *libraryQueryServer) ListResourcesByWorkId(ctx context.Context, req *gen.ListResourcesByWorkIdRequest) (*gen.ListResourcesByWorkIdResponse, error) {
	return s.provider.ListResourcesByWorkId(ctx, req)
}

// ===== 作者（本地轨 / 站点轨 / 作品关联）=====

func (s *libraryQueryServer) GetLocalAuthorById(ctx context.Context, req *gen.GetLocalAuthorByIdRequest) (*gen.LocalAuthorDTO, error) {
	return s.provider.GetLocalAuthorById(ctx, req)
}

func (s *libraryQueryServer) QueryLocalAuthors(ctx context.Context, req *gen.QueryLocalAuthorsRequest) (*gen.QueryLocalAuthorsResponse, error) {
	return s.provider.QueryLocalAuthors(ctx, req)
}

func (s *libraryQueryServer) GetSiteAuthorBySiteKey(ctx context.Context, req *gen.GetSiteAuthorBySiteKeyRequest) (*gen.SiteAuthorInfo, error) {
	return s.provider.GetSiteAuthorBySiteKey(ctx, req)
}

func (s *libraryQueryServer) QuerySiteAuthors(ctx context.Context, req *gen.QuerySiteAuthorsRequest) (*gen.QuerySiteAuthorsResponse, error) {
	return s.provider.QuerySiteAuthors(ctx, req)
}

func (s *libraryQueryServer) ListAuthorsByWorkId(ctx context.Context, req *gen.ListAuthorsByWorkIdRequest) (*gen.ListAuthorsByWorkIdResponse, error) {
	return s.provider.ListAuthorsByWorkId(ctx, req)
}

// ===== 标签（与作者对称；作品关联含关联级 namespace 维度）=====

func (s *libraryQueryServer) GetLocalTagById(ctx context.Context, req *gen.GetLocalTagByIdRequest) (*gen.LocalTagDTO, error) {
	return s.provider.GetLocalTagById(ctx, req)
}

func (s *libraryQueryServer) QueryLocalTags(ctx context.Context, req *gen.QueryLocalTagsRequest) (*gen.QueryLocalTagsResponse, error) {
	return s.provider.QueryLocalTags(ctx, req)
}

func (s *libraryQueryServer) GetSiteTagBySiteKey(ctx context.Context, req *gen.GetSiteTagBySiteKeyRequest) (*gen.SiteTagInfo, error) {
	return s.provider.GetSiteTagBySiteKey(ctx, req)
}

func (s *libraryQueryServer) QuerySiteTags(ctx context.Context, req *gen.QuerySiteTagsRequest) (*gen.QuerySiteTagsResponse, error) {
	return s.provider.QuerySiteTags(ctx, req)
}

func (s *libraryQueryServer) ListTagsByWorkId(ctx context.Context, req *gen.ListTagsByWorkIdRequest) (*gen.ListTagsByWorkIdResponse, error) {
	return s.provider.ListTagsByWorkId(ctx, req)
}

// ===== 作品集 =====

func (s *libraryQueryServer) GetWorkSetById(ctx context.Context, req *gen.GetWorkSetByIdRequest) (*gen.WorkSet, error) {
	return s.provider.GetWorkSetById(ctx, req)
}

func (s *libraryQueryServer) GetWorkSetBySiteKey(ctx context.Context, req *gen.GetWorkSetBySiteKeyRequest) (*gen.WorkSet, error) {
	return s.provider.GetWorkSetBySiteKey(ctx, req)
}

func (s *libraryQueryServer) ListWorkSetsByWorkId(ctx context.Context, req *gen.ListWorkSetsByWorkIdRequest) (*gen.ListWorkSetsByWorkIdResponse, error) {
	return s.provider.ListWorkSetsByWorkId(ctx, req)
}

func (s *libraryQueryServer) ListParentWorkSets(ctx context.Context, req *gen.ListParentWorkSetsRequest) (*gen.ListParentWorkSetsResponse, error) {
	return s.provider.ListParentWorkSets(ctx, req)
}

func (s *libraryQueryServer) ListChildWorkSets(ctx context.Context, req *gen.ListChildWorkSetsRequest) (*gen.ListChildWorkSetsResponse, error) {
	return s.provider.ListChildWorkSets(ctx, req)
}

// ===== 站点（注册表投影，只读）=====

func (s *libraryQueryServer) ListSites(ctx context.Context, req *gen.Empty) (*gen.ListSitesResponse, error) {
	return s.provider.ListSites(ctx, req)
}

// ===== 工作目录 =====

func (s *libraryQueryServer) GetWorkDir(ctx context.Context, req *gen.Empty) (*gen.GetWorkDirResponse, error) {
	return s.provider.GetWorkDir(ctx, req)
}
