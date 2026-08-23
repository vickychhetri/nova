package platform

import (
	"image"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/input"
)

// NativeWindow defines methods for interacting with an OS window.
type NativeWindow interface {
	BlitRGBA(img *image.RGBA)
	PollEvents() bool
	Close()
	SetCallbacks(
		onExpose func(),
		onResize func(w, h int),
		onPointerDown func(p geom.Point, btn input.MouseButton),
		onPointerUp func(p geom.Point, btn input.MouseButton),
		onPointerMove func(p geom.Point),
		onKeyDown func(e *event.KeyEvent),
		onClose func(),
	)
}
