//go:build windows

package window

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/lvfeng-z/library-squirrel-plugin-sdk/dto"
	"github.com/wailsapp/go-webview2/pkg/edge"
)

var (
	user32   = windows.NewLazyDLL("user32.dll")
	kernel32 = windows.NewLazyDLL("kernel32.dll")
	ole32    = windows.NewLazyDLL("ole32.dll")

	procRegisterClassExW  = user32.NewProc("RegisterClassExW")
	procCreateWindowExW   = user32.NewProc("CreateWindowExW")
	procDefWindowProcW    = user32.NewProc("DefWindowProcW")
	procDestroyWindow     = user32.NewProc("DestroyWindow")
	procShowWindow        = user32.NewProc("ShowWindow")
	procGetMessageW       = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessageW  = user32.NewProc("DispatchMessageW")
	procPostQuitMessage   = user32.NewProc("PostQuitMessage")
	procPostMessageW      = user32.NewProc("PostMessageW")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procGetModuleHandleW  = kernel32.NewProc("GetModuleHandleW")
	procLoadCursorW       = user32.NewProc("LoadCursorW")
	procSetWindowTextW    = user32.NewProc("SetWindowTextW")
	procCoInitializeEx    = ole32.NewProc("CoInitializeEx")
	procGetClientRect     = user32.NewProc("GetClientRect")
)

const (
	SW_SHOW             = 5
	WM_CLOSE            = 0x0010
	WM_DESTROY          = 0x0002
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WM_EXEC_SCRIPT      = 0x0400 + 100 // WM_USER + 100
)

var gwlpUserdata = uintptr(0xFFFFFFEB) // GWLP_USERDATA = -21

type wndClassExW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CnClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type winMsg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

var (
	wndProcPtr uintptr
	classOnce  sync.Once
	className  = syscall.StringToUTF16Ptr("LibrarySquirrelPluginWindow")
)

// ========== popupWindow ==========

type popupWindow struct {
	options   dto.WindowOptions
	ownerHWND uintptr
	hwnd      uintptr
	chromium  *edge.Chromium
	webview   *edge.ICoreWebView2

	navCh       chan string // buffered(1): 拦截到的 URL
	done        chan struct{}
	closeOnce   sync.Once
	scriptQueue chan *scriptRequest
}

type scriptRequest struct {
	js       string
	resultCh chan<- scriptResult
}

type scriptResult struct {
	json string
	err  error
}

// WindowHandle 接口实现

func (pw *popupWindow) Close() {
	if pw.hwnd != 0 {
		procPostMessageW.Call(pw.hwnd, WM_CLOSE, 0, 0)
	}
}

func (pw *popupWindow) SetTitle(title string) {
	if pw.hwnd != 0 {
		titlePtr, _ := windows.UTF16PtrFromString(title)
		procSetWindowTextW.Call(pw.hwnd, uintptr(unsafe.Pointer(titlePtr)))
	}
}

func (pw *popupWindow) WaitForNavigation(urlPrefix string, timeoutMs int64) (string, error) {
	var waitCtx context.Context
	var cancel context.CancelFunc

	if timeoutMs > 0 {
		waitCtx, cancel = context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	} else {
		waitCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	select {
	case url := <-pw.navCh:
		if urlPrefix != "" && !strings.HasPrefix(url, urlPrefix) {
			// URL 不匹配前缀，放回 channel 并继续等待
			select {
			case pw.navCh <- url:
			default:
			}
			return pw.WaitForNavigation(urlPrefix, timeoutMs)
		}
		return url, nil
	case <-pw.done:
		return "", fmt.Errorf("window closed")
	case <-waitCtx.Done():
		return "", waitCtx.Err()
	}
}

func (pw *popupWindow) ExecuteScript(js string) (string, error) {
	resultCh := make(chan scriptResult, 1)
	pw.scriptQueue <- &scriptRequest{js: js, resultCh: resultCh}
	// 唤醒 UI 线程处理脚本
	procPostMessageW.Call(pw.hwnd, WM_EXEC_SCRIPT, 0, 0)

	select {
	case r := <-resultCh:
		return r.json, r.err
	case <-pw.done:
		return "", fmt.Errorf("window closed")
	}
}

func (pw *popupWindow) Done() <-chan struct{} {
	return pw.done
}

// ========== NavigationStarting COM 类型 ==========

type iCoreWebView2NavigationStartingEventArgsVtbl struct {
	QueryInterface     edge.ComProc
	AddRef             edge.ComProc
	Release            edge.ComProc
	GetUri             edge.ComProc
	GetIsUserInitiated edge.ComProc
	GetIsRedirected    edge.ComProc
	GetNavigationId    edge.ComProc
	GetRequestHeaders  edge.ComProc
	PutCancel          edge.ComProc
	GetCancel          edge.ComProc
	GetDeferral        edge.ComProc
}

type coreWebView2NavigationStartingEventArgs struct {
	vtbl *iCoreWebView2NavigationStartingEventArgsVtbl
}

func (a *coreWebView2NavigationStartingEventArgs) GetUri() (string, error) {
	var uriPtr *uint16
	hr, _, _ := a.vtbl.GetUri.Call(
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(&uriPtr)),
	)
	if hr != 0 {
		return "", syscall.Errno(hr)
	}
	uri := windows.UTF16PtrToString(uriPtr)
	windows.CoTaskMemFree(unsafe.Pointer(uriPtr))
	return uri, nil
}

func (a *coreWebView2NavigationStartingEventArgs) PutCancel(cancel bool) error {
	var v int32
	if cancel {
		v = 1
	}
	hr, _, _ := a.vtbl.PutCancel.Call(
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(&v)),
	)
	if hr != 0 {
		return syscall.Errno(hr)
	}
	return nil
}

// ========== NavigationStarting 事件处理器 ==========

type iCoreWebView2NavigationStartingEventHandlerVtbl struct {
	QueryInterface edge.ComProc
	AddRef         edge.ComProc
	Release        edge.ComProc
	Invoke         edge.ComProc
}

type coreWebView2NavigationStartingEventHandler struct {
	vtbl *iCoreWebView2NavigationStartingEventHandlerVtbl
	impl navigationStartingCallback
}

type navigationStartingCallback func(sender *edge.ICoreWebView2, args *coreWebView2NavigationStartingEventArgs) uintptr

func navStartingQI(*coreWebView2NavigationStartingEventHandler, uintptr, uintptr) uintptr {
	return 0x80004002
}
func navStartingAddRef(*coreWebView2NavigationStartingEventHandler) uintptr  { return 1 }
func navStartingRelease(*coreWebView2NavigationStartingEventHandler) uintptr { return 1 }
func navStartingInvoke(this *coreWebView2NavigationStartingEventHandler, sender *edge.ICoreWebView2, args *coreWebView2NavigationStartingEventArgs) uintptr {
	return this.impl(sender, args)
}

var navStartingVtbl = iCoreWebView2NavigationStartingEventHandlerVtbl{
	QueryInterface: edge.NewComProc(navStartingQI),
	AddRef:         edge.NewComProc(navStartingAddRef),
	Release:        edge.NewComProc(navStartingRelease),
	Invoke:         edge.NewComProc(navStartingInvoke),
}

// ========== ExecuteScript 完成处理器 ==========

type iCoreWebView2ExecuteScriptCompletedHandlerVtbl struct {
	QueryInterface edge.ComProc
	AddRef         edge.ComProc
	Release        edge.ComProc
	Invoke         edge.ComProc
}

type executeScriptCompletedHandler struct {
	vtbl *iCoreWebView2ExecuteScriptCompletedHandlerVtbl
	impl func(errorCode uintptr, resultObjectAsJson *uint16) uintptr
}

func execScriptQI(*executeScriptCompletedHandler, uintptr, uintptr) uintptr {
	return 0x80004002
}
func execScriptAddRef(*executeScriptCompletedHandler) uintptr  { return 1 }
func execScriptRelease(*executeScriptCompletedHandler) uintptr { return 1 }
func execScriptInvoke(this *executeScriptCompletedHandler, errorCode uintptr, resultObjectAsJson *uint16) uintptr {
	return this.impl(errorCode, resultObjectAsJson)
}

var execScriptVtbl = iCoreWebView2ExecuteScriptCompletedHandlerVtbl{
	QueryInterface: edge.NewComProc(execScriptQI),
	AddRef:         edge.NewComProc(execScriptAddRef),
	Release:        edge.NewComProc(execScriptRelease),
	Invoke:         edge.NewComProc(execScriptInvoke),
}

// ========== ICoreWebView2 vtable 访问 ==========

type iCoreWebView2VtblPartial struct {
	QueryInterface                          edge.ComProc
	AddRef                                  edge.ComProc
	Release                                 edge.ComProc
	GetSettings                             edge.ComProc
	GetSource                               edge.ComProc
	Navigate                                edge.ComProc
	NavigateToString                        edge.ComProc
	AddNavigationStarting                   edge.ComProc
	RemoveNavigationStarting                edge.ComProc
	AddContentLoading                       edge.ComProc
	RemoveContentLoading                    edge.ComProc
	AddSourceChanged                        edge.ComProc
	RemoveSourceChanged                     edge.ComProc
	AddHistoryChanged                       edge.ComProc
	RemoveHistoryChanged                    edge.ComProc
	AddNavigationCompleted                  edge.ComProc
	RemoveNavigationCompleted               edge.ComProc
	AddFrameNavigationStarting              edge.ComProc
	RemoveFrameNavigationStarting           edge.ComProc
	AddFrameNavigationCompleted             edge.ComProc
	RemoveFrameNavigationCompleted          edge.ComProc
	AddScriptDialogOpening                  edge.ComProc
	RemoveScriptDialogOpening               edge.ComProc
	AddPermissionRequested                  edge.ComProc
	RemovePermissionRequested               edge.ComProc
	AddProcessFailed                        edge.ComProc
	RemoveProcessFailed                     edge.ComProc
	AddScriptToExecuteOnDocumentCreated     edge.ComProc
	RemoveScriptToExecuteOnDocumentCreated  edge.ComProc
	ExecuteScript                           edge.ComProc
}

type eventRegistrationToken struct {
	value int64
}

func getICoreWebView2Vtbl(wv *edge.ICoreWebView2) *iCoreWebView2VtblPartial {
	return *(**iCoreWebView2VtblPartial)(unsafe.Pointer(wv))
}

// ========== 获取 Chromium 非导出 webview 字段 ==========

func getChromiumWebview(c *edge.Chromium) *edge.ICoreWebView2 {
	v := reflect.ValueOf(c).Elem()
	f := v.FieldByName("webview")
	if f.IsNil() {
		return nil
	}
	return *(**edge.ICoreWebView2)(unsafe.Pointer(f.UnsafeAddr()))
}

// ========== 公共入口 ==========

func openWindow(options dto.WindowOptions, ownerHWND uintptr) (dto.WindowHandle, error) {
	pw := &popupWindow{
		options:     options,
		ownerHWND:   ownerHWND,
		navCh:       make(chan string, 1),
		done:        make(chan struct{}),
		scriptQueue: make(chan *scriptRequest, 16),
	}

	errCh := make(chan error, 1)
	go pw.run(errCh)

	if err := <-errCh; err != nil {
		return nil, err
	}
	return pw, nil
}

// ========== UI goroutine ==========

func (pw *popupWindow) run(errCh chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 初始化 COM（STA）
	procCoInitializeEx.Call(0, 0x2)

	// 注册窗口类
	classOnce.Do(func() {
		wndProcPtr = windows.NewCallback(globalWndProc)
		hInst, _, _ := procGetModuleHandleW.Call(0)
		hCursor, _, _ := procLoadCursorW.Call(0, 32512)
		wc := wndClassExW{
			CbSize:        uint32(unsafe.Sizeof(wndClassExW{})),
			LpfnWndProc:   wndProcPtr,
			HInstance:     hInst,
			HCursor:       hCursor,
			HbrBackground: 6,
			LpszClassName: className,
		}
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	})

	// 创建 Win32 窗口
	hInst, _, _ := procGetModuleHandleW.Call(0)
	titlePtr, _ := windows.UTF16PtrFromString(pw.options.Title)
	width := pw.options.Width
	height := pw.options.Height
	if width == 0 {
		width = 800
	}
	if height == 0 {
		height = 600
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(WS_OVERLAPPEDWINDOW),
		0x80000000, 0x80000000,
		uintptr(width), uintptr(height),
		pw.ownerHWND, 0, hInst, 0,
	)
	if hwnd == 0 {
		errCh <- fmt.Errorf("CreateWindowExW failed")
		return
	}
	pw.hwnd = hwnd

	// 存储 popupWindow 指针到窗口用户数据
	procSetWindowLongPtrW.Call(hwnd, gwlpUserdata, uintptr(unsafe.Pointer(pw)))

	// 初始化 WebView2
	chromium := edge.NewChromium()
	chromium.DataPath = pw.options.DataPath
	chromium.SetErrorCallback(func(err error) {
		log.Printf("[Window] WebView2 ERROR: %v", err)
	})

	if !chromium.Embed(hwnd) {
		errCh <- fmt.Errorf("WebView2 Embed failed")
		return
	}

	webview := getChromiumWebview(chromium)
	if webview == nil {
		errCh <- fmt.Errorf("WebView2 webview is nil after Embed")
		return
	}
	pw.chromium = chromium
	pw.webview = webview

	// 注册 NavigationStarting 事件
	if pw.options.OnNavigationStarting != nil {
		callback := pw.options.OnNavigationStarting
		handler := &coreWebView2NavigationStartingEventHandler{
			vtbl: &navStartingVtbl,
			impl: func(_ *edge.ICoreWebView2, args *coreWebView2NavigationStartingEventArgs) uintptr {
				uri, err := args.GetUri()
				if err != nil {
					return 0
				}
				if !callback(uri) {
					_ = args.PutCancel(true)
					select {
					case pw.navCh <- uri:
					default:
					}
				}
				return 0
			},
		}

		var token eventRegistrationToken
		vtbl := getICoreWebView2Vtbl(webview)
		vtbl.AddNavigationStarting.Call(
			uintptr(unsafe.Pointer(webview)),
			uintptr(unsafe.Pointer(handler)),
			uintptr(unsafe.Pointer(&token)),
		)
	}

	// 设置 WebView2 渲染区域为窗口客户区
	chromium.Resize()

	// 显示 WebView2 控制器（默认 IsVisible=false）
	if err := chromium.Show(); err != nil {
		log.Printf("[Window] chromium.Show() error: %v", err)
	}

	// 导航到目标 URL
	if pw.options.URL != "" {
		chromium.Navigate(pw.options.URL)
	}

	// 显示窗口
	procShowWindow.Call(hwnd, SW_SHOW)

	// 通知创建成功
	errCh <- nil

	// 消息循环
	var message winMsg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if ret == 0 {
			break
		}
		if message.Message == WM_EXEC_SCRIPT {
			pw.handleScriptQueue()
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
	}

	// 清理
	chromium.ShuttingDown()
	procDestroyWindow.Call(hwnd)
	pw.closeOnce.Do(func() { close(pw.done) })
}

func (pw *popupWindow) handleScriptQueue() {
	for {
		select {
		case req := <-pw.scriptQueue:
			resultCh := req.resultCh
			handler := &executeScriptCompletedHandler{
				vtbl: &execScriptVtbl,
				impl: func(_ uintptr, jsonPtr *uint16) uintptr {
					var json string
					if jsonPtr != nil {
						json = windows.UTF16PtrToString(jsonPtr)
						windows.CoTaskMemFree(unsafe.Pointer(jsonPtr))
					}
					resultCh <- scriptResult{json: json}
					return 0
				},
			}
			scriptPtr, _ := windows.UTF16PtrFromString(req.js)
			vtbl := getICoreWebView2Vtbl(pw.webview)
			vtbl.ExecuteScript.Call(
				uintptr(unsafe.Pointer(pw.webview)),
				uintptr(unsafe.Pointer(scriptPtr)),
				uintptr(unsafe.Pointer(handler)),
			)
		default:
			return
		}
	}
}

// ========== 全局窗口过程 ==========

func globalWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CLOSE:
		procDestroyWindow.Call(hwnd)
		return 0
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	case WM_EXEC_SCRIPT:
		// 从 GWLP_USERDATA 取回 popupWindow 指针
		ptr, _, _ := procSetWindowLongPtrW.Call(hwnd, gwlpUserdata, 0)
		procSetWindowLongPtrW.Call(hwnd, gwlpUserdata, ptr) // 写回
		if ptr != 0 {
			pw := (*popupWindow)(unsafe.Pointer(ptr))
			pw.handleScriptQueue()
		}
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}
