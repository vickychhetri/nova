package window

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/renderer/software"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
)

// Window represents an application window.
type Window struct {
	title       string
	size        geom.Size
	scale       float64
	rootComp    ui.Component
	rootNode    *ui.Node
	rasterizer  *software.Rasterizer
	cmdBuffer   *render.CommandBuffer
	activeTheme *theme.Theme
	focusedNode *ui.Node
	hoveredNode *ui.Node
}

// WindowOption configures Window attributes.
type WindowOption func(*Window)

// NewWindow creates a new application window.
func NewWindow(options ...WindowOption) *Window {
	w := &Window{
		title:       "Nova Application",
		size:        geom.Sz(1024, 768),
		scale:       1.0,
		rasterizer:  software.NewRasterizer(),
		cmdBuffer:   render.NewCommandBuffer(),
		activeTheme: theme.Current(),
	}

	for _, opt := range options {
		opt(w)
	}

	return w
}

// Option helpers
func WithTitle(title string) WindowOption {
	return func(w *Window) {
		w.title = title
	}
}

func WithSize(width, height float64) WindowOption {
	return func(w *Window) {
		w.size = geom.Sz(width, height)
	}
}

func WithTheme(t *theme.Theme) WindowOption {
	return func(w *Window) {
		w.activeTheme = t
	}
}

// Content sets the root component content for this window.
func (w *Window) Content(comp ui.Component) {
	w.rootComp = comp
	w.rootNode = ui.NewNode(comp)
	w.rootNode.Mount(ui.BuildContext{
		Theme: w.activeTheme,
		Scale: w.scale,
	})
}

// RenderFrame renders a complete frame of the UI tree.
func (w *Window) RenderFrame() {
	if w.rootNode == nil {
		return
	}

	// Layout pass
	w.rootNode.Layout(layout.Tight(w.size))

	// Paint pass
	w.cmdBuffer.Clear()
	canvas := render.NewCanvas(w.cmdBuffer)

	// Draw background
	canvas.FillRect(geom.NewRect(0, 0, w.size.Width, w.size.Height), w.activeTheme.Palette.Background)

	w.rootNode.Paint(canvas)

	// Rasterize
	w.rasterizer.BeginFrame(w.size, w.scale)
	w.rasterizer.Render(w.cmdBuffer.Commands())
	w.rasterizer.EndFrame()
}

// DispatchPointerDown routes a mouse press event into the UI tree.
func (w *Window) DispatchPointerDown(p geom.Point, btn int) {
	if w.rootNode == nil {
		return
	}
	hit := w.rootNode.HitTest(p)
	if hit != nil {
		w.focusedNode = hit
		hit.IsFocused = true
		if hit.OnClick != nil {
			hit.OnClick()
		}
		if hit.OnPointerDown != nil {
			hit.OnPointerDown(&event.PointerEvent{Position: p})
		}
	}
}

// DispatchPointerMove routes pointer movement and handles hover states.
func (w *Window) DispatchPointerMove(p geom.Point) {
	if w.rootNode == nil {
		return
	}
	hit := w.rootNode.HitTest(p)
	if hit != w.hoveredNode {
		if w.hoveredNode != nil {
			w.hoveredNode.IsHovered = false
			if w.hoveredNode.OnPointerLeave != nil {
				w.hoveredNode.OnPointerLeave()
			}
		}
		w.hoveredNode = hit
		if w.hoveredNode != nil {
			w.hoveredNode.IsHovered = true
			if w.hoveredNode.OnPointerEnter != nil {
				w.hoveredNode.OnPointerEnter()
			}
		}
	}

	if hit != nil && hit.OnPointerMove != nil {
		hit.OnPointerMove(&event.PointerEvent{Position: p})
	}
}

// DispatchKeyDown routes keyboard events to focused node.
func (w *Window) DispatchKeyDown(e *event.KeyEvent) {
	if w.focusedNode != nil && w.focusedNode.OnKeyDown != nil {
		w.focusedNode.OnKeyDown(e)
	}
}

// SaveScreenshot exports current window frame to PNG.
func (w *Window) SaveScreenshot(path string) error {
	w.RenderFrame()
	return w.rasterizer.SavePNG(path)
}

// Title returns current window title.
func (w *Window) Title() string {
	return w.title
}

// Size returns current window size.
func (w *Window) Size() geom.Size {
	return w.size
}
