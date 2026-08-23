//go:build linux && !headless

package platform

import (
	"github.com/vickychhetri/nova/platform/linux"
)

// CreatePlatformWindow initializes a native X11 window on Linux.
func CreatePlatformWindow(title string, width, height int) (NativeWindow, error) {
	win, err := linux.NewX11Window(title, width, height)
	if err != nil || win == nil {
		return nil, err
	}
	return win, nil
}
