package window

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/input"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/platform"
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
	nativeWin   platform.NativeWindow
	needsRedraw bool
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
		needsRedraw: true,
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

// Invalidate requests a redraw on the next event loop tick.
func (w *Window) Invalidate() {
	w.needsRedraw = true
}

// NeedsRedraw reports whether the window needs to be re-rendered.
func (w *Window) NeedsRedraw() bool {
	return w.needsRedraw
}

// Content sets the root component content for this window.
func (w *Window) Content(comp ui.Component) {
	if w.activeTheme != nil {
		theme.SetCurrent(w.activeTheme)
	}

	var prevFocusedComp ui.Component
	if w.focusedNode != nil {
		prevFocusedComp = w.focusedNode.Component
	}

	w.rootComp = comp
	w.rootNode = ui.NewNode(comp)
	w.rootNode.Mount(ui.BuildContext{
		Theme: w.activeTheme,
		Scale: w.scale,
	})

	if prevFocusedComp != nil {
		if newFocusedNode := findNodeByComponent(w.rootNode, prevFocusedComp); newFocusedNode != nil {
			w.focusedNode = newFocusedNode
			newFocusedNode.IsFocused = true
		} else {
			w.focusedNode = nil
		}
	} else {
		w.focusedNode = nil
	}

	w.needsRedraw = true
}

func findNodeByComponent(n *ui.Node, comp ui.Component) *ui.Node {
	if n == nil || comp == nil {
		return nil
	}
	if n.Component == comp {
		return n
	}
	for _, ch := range n.Children {
		if found := findNodeByComponent(ch, comp); found != nil {
			return found
		}
	}
	return nil
}

// AttachNative attaches an OS native window handle and registers callbacks.
func (w *Window) AttachNative(nw platform.NativeWindow) {
	w.nativeWin = nw
	if nw != nil {
		nw.SetCallbacks(
			func() {
				w.needsRedraw = true
			},
			func(newW, newH int) {
				w.size = geom.Sz(float64(newW), float64(newH))
				w.needsRedraw = true
			},
			func(p geom.Point, btn input.MouseButton) {
				w.DispatchPointerDown(p, int(btn))
				w.needsRedraw = true
			},
			func(p geom.Point, btn input.MouseButton) {
				w.DispatchPointerUp(p, int(btn))
				w.needsRedraw = true
			},
			func(p geom.Point) {
				if w.DispatchPointerMove(p) {
					w.needsRedraw = true
				}
			},
			func(e *event.KeyEvent) {
				w.DispatchKeyDown(e)
				w.needsRedraw = true
			},
			func() {
				w.nativeWin = nil
			},
		)
	}
}

// NativeWindow returns active OS window handle.
func (w *Window) NativeWindow() platform.NativeWindow {
	return w.nativeWin
}

// RenderFrame renders a complete frame of the UI tree and blits to native window if present.
func (w *Window) RenderFrame() {
	w.needsRedraw = false
	if w.rootNode == nil {
		return
	}
	if w.activeTheme != nil {
		theme.SetCurrent(w.activeTheme)
	}

	// Layout pass
	w.rootNode.Layout(layout.Tight(w.size))

	// Paint pass
	w.cmdBuffer.Clear()
	canvas := render.NewCanvas(w.cmdBuffer)

	// Draw background
	canvas.FillRect(geom.NewRect(0, 0, w.size.Width, w.size.Height), w.activeTheme.Palette.Background)

	// Paint base UI tree
	w.rootNode.Paint(canvas)

	// Paint floating overlays on top of all base UI content
	w.rootNode.PaintOverlays(canvas)

	// Rasterize
	w.rasterizer.BeginFrame(w.size, w.scale)
	w.rasterizer.Render(w.cmdBuffer.Commands())
	w.rasterizer.EndFrame()

	// Blit to native display surface
	if w.nativeWin != nil {
		w.nativeWin.BlitRGBA(w.rasterizer.Buffer())
	}
}

// DispatchPointerDown routes a mouse press event into the UI tree and switches focus cleanly.
func (w *Window) DispatchPointerDown(p geom.Point, btn int) {
	if w.rootNode == nil {
		return
	}

	// 1. Check if an active overlay (dropdown menu, dialog, popup) intercepts the click
	if w.rootNode.DispatchOverlayPointerDown(p) {
		w.needsRedraw = true
		return
	}

	hit, localP := w.rootNode.HitTestLocal(p)
	if w.focusedNode != hit {
		if w.focusedNode != nil {
			w.focusedNode.IsFocused = false
		}
		w.focusedNode = hit
		if hit != nil {
			hit.IsFocused = true
		}
	}
	if hit != nil {
		if hit.OnClick != nil {
			hit.OnClick()
		}
		if hit.OnPointerDown != nil {
			hit.OnPointerDown(&event.PointerEvent{Position: localP})
		}
	}
}

// DispatchPointerUp routes mouse release event into the UI tree.
func (w *Window) DispatchPointerUp(p geom.Point, btn int) {
	if w.rootNode == nil {
		return
	}
	hit, localP := w.rootNode.HitTestLocal(p)
	if hit != nil && hit.OnPointerUp != nil {
		hit.OnPointerUp(&event.PointerEvent{Position: localP})
	}
}

// DispatchPointerMove routes pointer movement and handles hover states without redundant frames.
func (w *Window) DispatchPointerMove(p geom.Point) bool {
	if w.rootNode == nil {
		return false
	}

	// Check if active overlay handled mouse movement (e.g. dropdown item hover)
	if w.rootNode.DispatchOverlayPointerMove(p) {
		return true
	}

	changed := false
	hit, localP := w.rootNode.HitTestLocal(p)
	if hit != w.hoveredNode {
		changed = true
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
		hit.OnPointerMove(&event.PointerEvent{Position: localP})
		changed = true
	}
	return changed
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
