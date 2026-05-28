//go:build !windows

package window

import (
	"fmt"

	"github.com/lvfeng-z/library-squirrel-plugin-sdk/dto"
)

func openWindow(options dto.WindowOptions, ownerHWND uintptr) (dto.WindowHandle, error) {
	return nil, fmt.Errorf("popup windows not supported on this platform")
}
