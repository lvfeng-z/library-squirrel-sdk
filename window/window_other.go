//go:build !windows

package window

import (
	"fmt"

	pluginsdk "github.com/lvfeng-z/library-squirrel-plugin-sdk"
)

func openWindow(options pluginsdk.WindowOptions, ownerHWND uintptr) (pluginsdk.WindowHandle, error) {
	return nil, fmt.Errorf("popup windows not supported on this platform")
}
