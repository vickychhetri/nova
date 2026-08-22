package platform

import (
	"github.com/vickychhetri/nova/core/geom"
)

// WindowConfig configures window creation parameters.
type WindowConfig struct {
	Title     string
	Width     float64
	Height    float64
	Resizable bool
}

// WindowHandle represents the OS-level window instance.
type WindowHandle interface {
	SetTitle(title string)
	SetSize(size geom.Size)
	Close()
}

// Platform is the OS integration interface.
type Platform interface {
	CreateWindow(cfg WindowConfig) (WindowHandle, error)
	GetClipboard() string
	SetClipboard(text string)
}
