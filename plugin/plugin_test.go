package plugin

import (
	"testing"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
)

// taskHandlerStub 占位任务处理器（嵌入 nil dto.TaskHandler 满足接口，方法不被触达）
type taskHandlerStub struct {
	dto.TaskHandler
}

// browserStub 占位站点浏览器（同上）
type browserStub struct {
	dto.SiteBrowser
}

// TestNewLSPluginWithoutTaskHandler 仅可选项（Activate/Shutdown）构造：Handler 为 nil，
// 无任务处理器的工具型插件运行时形态可构造
func TestNewLSPluginWithoutTaskHandler(t *testing.T) {
	activated := false
	p := newLSPlugin(
		WithActivate(func(dto.PluginContext) { activated = true }),
		WithShutdown(func() {}),
	)
	if p.Handler != nil {
		t.Fatal("未设 WithTaskHandler 时 Handler 应为 nil")
	}
	if p.OnActivate == nil || p.OnShutdown == nil {
		t.Fatal("可选项应装配到位")
	}
	p.OnActivate(nil)
	if !activated {
		t.Fatal("Activate 回调应为所设函数")
	}
}

// TestNewLSPluginWithTaskHandler 下载型插件全选项形态：四个选项各自落位
func TestNewLSPluginWithTaskHandler(t *testing.T) {
	handler := &taskHandlerStub{}
	browser := &browserStub{}
	p := newLSPlugin(
		WithTaskHandler(handler),
		WithBrowser(browser),
		WithActivate(func(dto.PluginContext) {}),
		WithShutdown(func() {}),
	)
	if p.Handler != handler || p.Browser != browser || p.OnActivate == nil || p.OnShutdown == nil {
		t.Fatal("全部选项应各自落位")
	}
}
