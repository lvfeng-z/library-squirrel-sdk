package dto

import (
	"context"

	"github.com/lvfeng-z/library-squirrel-sdk/gen"
)

// FetchSiteAuthorInfoRequest 站点作者信息拉取请求（别名 gen.FetchSiteAuthorInfoRequest，proto 单源）
type FetchSiteAuthorInfoRequest = gen.FetchSiteAuthorInfoRequest

// AuthorInfoChunk 拉取响应块（别名 gen.AuthorInfoChunk，proto 单源；oneof payload：
// meta 恒为首块且仅一块，resource 头像字节块可多块）
type AuthorInfoChunk = gen.AuthorInfoChunk

// AuthorInfoChunk_Meta meta 载荷的 oneof 包装类型（别名 gen.AuthorInfoChunk_Meta，
// 插件/主程序构造块时经本别名引用，不直接 import gen）
type AuthorInfoChunk_Meta = gen.AuthorInfoChunk_Meta

// AuthorInfoChunk_Resource resource 载荷的 oneof 包装类型（别名 gen.AuthorInfoChunk_Resource）
type AuthorInfoChunk_Resource = gen.AuthorInfoChunk_Resource

// AuthorInfoMeta 作者元数据（别名 gen.AuthorInfoMeta，proto 单源；字段语义见 gen 包）
type AuthorInfoMeta = gen.AuthorInfoMeta

// AuthorResourceData 头像字节块（别名 gen.AuthorResourceData，proto 单源）
type AuthorResourceData = gen.AuthorResourceData

// SiteAuthorFetcher 站点作者信息拉取处理器（可选扩展点，「站点实体元数据+资源拉取」
// 契约家族首成员）。经 WithSiteAuthorFetcher 注册；主程序按请求 siteKey 广播路由。
type SiteAuthorFetcher interface {
	// FetchSiteAuthorInfo 按站点身份键拉取作者最新元数据与头像资源。
	// 首块恒为 meta（含头像 URL/format 前置声明），后续块为头像字节；
	// send 向主程序发送一个响应块，返回错误时中止流。ctx 为 gRPC stream context，
	// 主程序取消拉取时经 ctx 传播，插件应在阻塞读取中响应取消。
	FetchSiteAuthorInfo(ctx context.Context, req *FetchSiteAuthorInfoRequest, send func(chunk *AuthorInfoChunk) error) error
}
