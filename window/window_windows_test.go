//go:build windows

package window

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lvfeng-z/library-squirrel-sdk/dto"
)

// gcSink 承接堆搅动分配,防止编译器将分配省略。
var gcSink []byte

// TestComHandlerKeepAliveRegistry 锚定保活注册表登记/释放语义（纯内存,不开窗）。
func TestComHandlerKeepAliveRegistry(t *testing.T) {
	pw := &popupWindow{comHandlers: make(map[any]struct{})}
	h := &executeScriptCompletedHandler{}

	pw.retainComHandler(h)
	if _, ok := pw.comHandlers[h]; !ok {
		t.Fatal("retain 后注册表应含 handler")
	}

	pw.releaseComHandler(h)
	if _, ok := pw.comHandlers[h]; ok {
		t.Fatal("release 后注册表不应再含 handler")
	}

	pw.releaseComHandler(h) // 重复释放须幂等
}

// TestExecuteScriptKeepAliveLive 实机回归：弹窗内单发 + 高频 ExecuteScript 叠加强制 GC 与堆搅动。
// 锚定两条修复：①完成回调 handler 经 unsafe.Pointer 交给原生侧后由注册表保活（否则对 GC 不可见,
// 回收后原生回调读悬空内存）;②resultObjectAsJson 为 WebView2 回调入参,仅复制不释放（越权
// CoTaskMemFree 非 CoTaskMem 堆内存即堆损坏 0xc0000374,bilibili 阶段D 首用即崩的实因）。
// 需桌面会话与 WebView2 运行时,默认跳过,置 LS_WINDOW_UAF_CHECK=1 启用（会短暂弹出小窗）。
func TestExecuteScriptKeepAliveLive(t *testing.T) {
	if os.Getenv("LS_WINDOW_UAF_CHECK") == "" {
		t.Skip("置 LS_WINDOW_UAF_CHECK=1 启用实机回归（需桌面会话+WebView2,会短暂弹出小窗）")
	}

	handle, err := OpenWindow(dto.WindowOptions{
		Title:    "ExecuteScript UAF 回归",
		Width:    320,
		Height:   240,
		URL:      "about:blank",
		DataPath: filepath.Join(t.TempDir(), "webview2"),
	}, 0)
	if err != nil {
		t.Fatalf("OpenWindow: %v", err)
	}
	defer handle.Close()

	// 单发验证完成回调链路（errorCode=S_OK、结果字符串解码）
	if _, err := handle.ExecuteScript("1+1"); err != nil {
		t.Fatalf("单发 ExecuteScript: %v", err)
	}

	for i := 0; i < 300; i++ {
		runtime.GC()
		gcSink = make([]byte, 1<<20) // 搅动堆,促使已回收内存被新分配覆写
		if _, err := handle.ExecuteScript("1+1"); err != nil {
			t.Fatalf("第 %d 次 ExecuteScript: %v", i, err)
		}
	}
}
