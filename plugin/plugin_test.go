package plugin

import (
	"testing"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
)

// workFetcherStub 占位作品拉取扩展（嵌入 nil dto.WorkFetcher 满足接口，方法不被触达）
type workFetcherStub struct {
	dto.WorkFetcher
}

// browserStub 占位站点浏览器（同上）
type browserStub struct {
	dto.SiteBrowser
}

// siteAuthorFetcherStub 占位站点作者信息拉取器（同上）
type siteAuthorFetcherStub struct {
	dto.SiteAuthorFetcher
}

// TestNewLSPluginWithoutWorkFetcher 仅可选项（Activate/Shutdown）构造：Handler 为 nil，
// 无作品拉取扩展的工具型插件运行时形态可构造
func TestNewLSPluginWithoutWorkFetcher(t *testing.T) {
	activated := false
	p := newLSPlugin(
		WithActivate(func(dto.PluginContext) { activated = true }),
		WithShutdown(func() {}),
	)
	if p.Handler != nil {
		t.Fatal("未设 WithWorkFetcher 时 Handler 应为 nil")
	}
	if p.OnActivate == nil || p.OnShutdown == nil {
		t.Fatal("可选项应装配到位")
	}
	p.OnActivate(nil)
	if !activated {
		t.Fatal("Activate 回调应为所设函数")
	}
}

// TestNewLSPluginWithWorkFetcher 作品拉取型插件全选项形态：四个选项各自落位
func TestNewLSPluginWithWorkFetcher(t *testing.T) {
	handler := &workFetcherStub{}
	browser := &browserStub{}
	p := newLSPlugin(
		WithWorkFetcher(handler),
		WithBrowser(browser),
		WithActivate(func(dto.PluginContext) {}),
		WithShutdown(func() {}),
	)
	if p.Handler != handler || p.Browser != browser || p.OnActivate == nil || p.OnShutdown == nil {
		t.Fatal("全部选项应各自落位")
	}
}

// TestNewLSPluginWithSiteAuthorFetcher 拉取扩展点选项落位：未设时为 nil（service 不注册）；
// 多次调用按条目 id 累积成 id→实现 表，同 id 后调覆盖
func TestNewLSPluginWithSiteAuthorFetcher(t *testing.T) {
	if p := newLSPlugin(WithActivate(func(dto.PluginContext) {})); p.SiteAuthorFetchers != nil {
		t.Fatal("未设 WithSiteAuthorFetcher 时 SiteAuthorFetchers 应为 nil")
	}
	first, second, alt := &siteAuthorFetcherStub{}, &siteAuthorFetcherStub{}, &siteAuthorFetcherStub{}
	p := newLSPlugin(
		WithSiteAuthorFetcher("main", first),
		WithSiteAuthorFetcher("alt", alt),
		WithSiteAuthorFetcher("main", second),
	)
	if len(p.SiteAuthorFetchers) != 2 || p.SiteAuthorFetchers["main"] != second || p.SiteAuthorFetchers["alt"] != alt {
		t.Fatal("WithSiteAuthorFetcher 多次调用应按条目 id 累积落位（同 id 后调覆盖）")
	}
}
