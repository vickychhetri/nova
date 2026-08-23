//go:build !linux || headless

package platform

// CreatePlatformWindow returns nil on headless or unsupported platforms.
func CreatePlatformWindow(title string, width, height int) (NativeWindow, error) {
	return nil, nil
}
