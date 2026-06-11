//go:build windows

package plugin

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func init() {
	appID := os.Getenv("LIBRARY_SQUIRREL_APP_USER_MODEL_ID")
	if appID == "" {
		return
	}
	setCurrentProcessAppUserModelID(appID)
}

// setCurrentProcessAppUserModelID 调用 Shell32 API 设置当前进程的 AppUserModelID
func setCurrentProcessAppUserModelID(appID string) {
	shell32 := windows.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("SetCurrentProcessExplicitAppUserModelID")

	appIDPtr, err := windows.UTF16PtrFromString(appID)
	if err != nil {
		return
	}

	proc.Call(uintptr(unsafe.Pointer(appIDPtr)))
}
