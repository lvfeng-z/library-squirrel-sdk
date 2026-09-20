package identity

import (
	"errors"
	"fmt"
)

// ErrSiteNotOwned 请求的站点键不属于本插件（拉取契约能力广播路由的消费信号）。
// 插件实现 dto.SiteAuthorFetcher.FetchSiteAuthorInfo 时经 CheckSiteOwnership 归属自判
// 返回本错误（可被 errors.Is 识别），SDK 服务端适配层将其转译为 gRPC PermissionDenied
// 跨进程传递，主程序广播路由据此静默跳过本插件继续遍历（区别于拉取失败的记日志处理）。
var ErrSiteNotOwned = errors.New("站点实体拉取请求未归属本插件")

// CheckSiteOwnership 拉取请求归属自判：比对请求 siteKey 与插件自身归属的站点身份键。
// 插件侧用法（以 bilibili 插件为例）：
//
//	if err := identity.CheckSiteOwnership(identity.Bilibili.Key, req.SiteKey); err != nil {
//		return err
//	}
//
// 相符返回 nil；不符返回包装 ErrSiteNotOwned 的错误。
func CheckSiteOwnership(pluginSiteKey string, requestSiteKey string) error {
	if pluginSiteKey != requestSiteKey {
		return fmt.Errorf("%w: 请求站点键 %q, 本插件归属 %q", ErrSiteNotOwned, requestSiteKey, pluginSiteKey)
	}
	return nil
}
