package transport

import (
	"context"
	"io"
	"testing"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
	"github.com/lvfeng-z/library-squirrel-sdk/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fetcherStub 记录到达的请求并按脚本回放响应块（err 非空时直接返回，不发块）
type fetcherStub struct {
	requests []*dto.FetchSiteAuthorInfoRequest
	chunks   []*dto.AuthorInfoChunk
	err      error
}

func (f *fetcherStub) FetchSiteAuthorInfo(ctx context.Context, req *dto.FetchSiteAuthorInfoRequest, send func(chunk *dto.AuthorInfoChunk) error) error {
	f.requests = append(f.requests, req)
	if f.err != nil {
		return f.err
	}
	for _, chunk := range f.chunks {
		if err := send(chunk); err != nil {
			return err
		}
	}
	return nil
}

// TestSiteAuthorFetchWithoutFetcherUnimplemented 未提供 SiteAuthorFetcher 形态：
// SiteAuthorFetchService 未注册，主程序侧拉取 RPC 得 codes.Unimplemented
// （对照 TestGRPCServerWithoutWorkFetcherUnimplemented 同款契约）
func TestSiteAuthorFetchWithoutFetcherUnimplemented(t *testing.T) {
	conn := serveGRPC(t, func(s *grpc.Server) {
		if err := (&LSPlugin{}).GRPCServer(nil, s); err != nil {
			t.Errorf("GRPCServer 注册失败: %v", err)
		}
	})
	client := gen.NewSiteAuthorFetchServiceClient(conn)
	stream, err := client.FetchSiteAuthorInfo(context.Background(), &gen.FetchSiteAuthorInfoRequest{SiteKey: "bilibili", SiteAuthorId: "42"})
	if err == nil {
		_, err = stream.Recv()
	}
	if status.Code(err) != codes.Unimplemented {
		t.Fatalf("错误码 = %v, 期望 codes.Unimplemented (err=%v)", status.Code(err), err)
	}
}

// TestSiteAuthorFetchWithFetcherStreams 对照：提供 fetcher 后拉取流全链到达插件处理器——
// 请求字段直通、首块 meta、后续块头像字节、EOF 收尾
func TestSiteAuthorFetchWithFetcherStreams(t *testing.T) {
	fetcher := &fetcherStub{chunks: []*dto.AuthorInfoChunk{
		{Payload: &dto.AuthorInfoChunk_Meta{Meta: &dto.AuthorInfoMeta{
			AuthorName:   "某UP主",
			Introduce:    "签名",
			Homepage:     "https://space.example.test/42",
			AvatarUrl:    "https://example.test/avatar.jpg",
			AvatarFormat: "jpg",
		}}},
		{Payload: &dto.AuthorInfoChunk_Resource{Resource: &dto.AuthorResourceData{Data: []byte("avatar-bytes")}}},
	}}
	conn := serveGRPC(t, func(s *grpc.Server) {
		if err := (&LSPlugin{SiteAuthorFetchers: map[string]dto.SiteAuthorFetcher{"main": fetcher}}).GRPCServer(nil, s); err != nil {
			t.Errorf("GRPCServer 注册失败: %v", err)
		}
	})
	client := gen.NewSiteAuthorFetchServiceClient(conn)
	stream, err := client.FetchSiteAuthorInfo(context.Background(), &gen.FetchSiteAuthorInfoRequest{SiteKey: "bilibili", SiteAuthorId: "42", ExtensionId: "main"})
	if err != nil {
		t.Fatalf("FetchSiteAuthorInfo 开流失败: %v", err)
	}
	meta, err := stream.Recv()
	if err != nil {
		t.Fatalf("首块(meta)接收失败: %v", err)
	}
	if meta.GetMeta().GetAuthorName() != "某UP主" || meta.GetMeta().GetAvatarUrl() != "https://example.test/avatar.jpg" || meta.GetMeta().GetAvatarFormat() != "jpg" {
		t.Fatalf("meta 内容不符: %+v", meta.GetMeta())
	}
	data, err := stream.Recv()
	if err != nil {
		t.Fatalf("第二块(resource)接收失败: %v", err)
	}
	if string(data.GetResource().GetData()) != "avatar-bytes" {
		t.Fatalf("resource 内容不符: %q", data.GetResource().GetData())
	}
	if _, err := stream.Recv(); err != io.EOF {
		t.Fatalf("流应结束于 EOF, 实得 %v", err)
	}
	if len(fetcher.requests) != 1 || fetcher.requests[0].SiteKey != "bilibili" || fetcher.requests[0].SiteAuthorId != "42" || fetcher.requests[0].ExtensionId != "main" {
		t.Fatalf("请求应直通插件处理器: %+v", fetcher.requests)
	}
}

// TestSiteAuthorFetchDispatchesByExtensionId 多实例分派：同插件注册两个条目实现，请求按
// extensionId 路由到对应条目，另一条目不被触达
func TestSiteAuthorFetchDispatchesByExtensionId(t *testing.T) {
	mainFetcher := &fetcherStub{}
	altFetcher := &fetcherStub{chunks: []*dto.AuthorInfoChunk{
		{Payload: &dto.AuthorInfoChunk_Meta{Meta: &dto.AuthorInfoMeta{AuthorName: "alt 实例作者"}}},
	}}
	conn := serveGRPC(t, func(s *grpc.Server) {
		if err := (&LSPlugin{SiteAuthorFetchers: map[string]dto.SiteAuthorFetcher{"main": mainFetcher, "alt": altFetcher}}).GRPCServer(nil, s); err != nil {
			t.Errorf("GRPCServer 注册失败: %v", err)
		}
	})
	client := gen.NewSiteAuthorFetchServiceClient(conn)
	stream, err := client.FetchSiteAuthorInfo(context.Background(), &gen.FetchSiteAuthorInfoRequest{SiteKey: "bilibili", SiteAuthorId: "42", ExtensionId: "alt"})
	if err != nil {
		t.Fatalf("FetchSiteAuthorInfo 开流失败: %v", err)
	}
	meta, err := stream.Recv()
	if err != nil {
		t.Fatalf("首块(meta)接收失败: %v", err)
	}
	if meta.GetMeta().GetAuthorName() != "alt 实例作者" {
		t.Fatalf("应路由到 alt 条目实现: %+v", meta.GetMeta())
	}
	if len(altFetcher.requests) != 1 || altFetcher.requests[0].ExtensionId != "alt" {
		t.Fatalf("alt 条目应收到带本条目 id 的请求: %+v", altFetcher.requests)
	}
	if len(mainFetcher.requests) != 0 {
		t.Fatalf("main 条目不应被触达: %+v", mainFetcher.requests)
	}
}

// TestSiteAuthorFetchUnknownExtensionIdInvalidArgument 条目 id 分派未命中：请求 extensionId
// 不在插件注册表内时返回 codes.InvalidArgument（插件侧分派拦截）
func TestSiteAuthorFetchUnknownExtensionIdInvalidArgument(t *testing.T) {
	conn := serveGRPC(t, func(s *grpc.Server) {
		if err := (&LSPlugin{SiteAuthorFetchers: map[string]dto.SiteAuthorFetcher{"main": &fetcherStub{}}}).GRPCServer(nil, s); err != nil {
			t.Errorf("GRPCServer 注册失败: %v", err)
		}
	})
	client := gen.NewSiteAuthorFetchServiceClient(conn)
	stream, err := client.FetchSiteAuthorInfo(context.Background(), &gen.FetchSiteAuthorInfoRequest{SiteKey: "bilibili", ExtensionId: "ghost"})
	if err == nil {
		_, err = stream.Recv()
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("错误码 = %v, 期望 codes.InvalidArgument (err=%v)", status.Code(err), err)
	}
}

// TestSiteAuthorFetchInternalStatus 处理器返回的错误按 Internal 转译（客户端侧无
// 归属信号分支，一律按拉取失败处理）
func TestSiteAuthorFetchInternalStatus(t *testing.T) {
	fetcher := &fetcherStub{err: io.ErrUnexpectedEOF}
	conn := serveGRPC(t, func(s *grpc.Server) {
		if err := (&LSPlugin{SiteAuthorFetchers: map[string]dto.SiteAuthorFetcher{"main": fetcher}}).GRPCServer(nil, s); err != nil {
			t.Errorf("GRPCServer 注册失败: %v", err)
		}
	})
	client := gen.NewSiteAuthorFetchServiceClient(conn)
	stream, err := client.FetchSiteAuthorInfo(context.Background(), &gen.FetchSiteAuthorInfoRequest{SiteKey: "bilibili", ExtensionId: "main"})
	if err == nil {
		_, err = stream.Recv()
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("错误码 = %v, 期望 codes.Internal (err=%v)", status.Code(err), err)
	}
}
