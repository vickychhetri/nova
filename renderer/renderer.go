package renderer

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/render"
)

// RenderTarget represents a window or offscreen surface that can be rendered to.
type RenderTarget interface {
	Size() geom.Size
	Scale() float64
}

// Renderer is the pluggable graphics backend interface.
type Renderer interface {
	Init(target RenderTarget) error
	BeginFrame(size geom.Size, scale float64)
	Render(commands []render.Command)
	EndFrame()
	Destroy()
}
