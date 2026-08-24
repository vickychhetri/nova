package render

import (
	"image"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
)

// Canvas provides an ergonomic 2D drawing API that records commands into a
// CommandBuffer instead of drawing immediately.
//
// Canvas keeps a lightweight local coordinate state. Drawing methods apply the
// current translation to their geometry before the resulting Command is added
// to Buffer. The renderer later consumes those commands in recording order.
type Canvas struct {
	// Buffer receives every command recorded by this canvas. It must be
	// non-nil before calling any drawing method.
	Buffer *CommandBuffer
	// offset is the translation applied to subsequently recorded geometry.
	offset geom.Point
	// stack stores offsets saved by Save, in last-in-first-out order.
	stack []geom.Point
}

// NewCanvas creates a Canvas that records to buf.
//
// The buffer is supplied by the caller so multiple drawing layers can share
// one command stream when needed. The canvas starts with a zero translation
// and reserves stack capacity for common nesting depths.
func NewCanvas(buf *CommandBuffer) *Canvas {
	return &Canvas{
		Buffer: buf,
		stack:  make([]geom.Point, 0, 16),
	}
}

// Save pushes the current translation onto the transform stack.
//
// Save does not add a render command. It only stores Canvas-side state so a
// later Restore can return to the same local coordinate system.
func (c *Canvas) Save() {
	c.stack = append(c.stack, c.offset)
}

// Restore restores the most recently saved translation.
//
// An unmatched Restore is intentionally a no-op. The method does not affect
// commands already recorded and does not modify clipping state.
func (c *Canvas) Restore() {
	if len(c.stack) > 0 {
		c.offset = c.stack[len(c.stack)-1]
		c.stack = c.stack[:len(c.stack)-1]
	}
}

// Translate shifts all subsequent coordinate operations by (dx, dy).
//
// Translation is cumulative and affects only commands recorded after this
// call. It moves positions and rectangles, but not size values such as circle
// radii or stroke widths.
func (c *Canvas) Translate(dx, dy float64) {
	c.offset.X += dx
	c.offset.Y += dy
}

// CurrentOffset returns the current translation offset as a value.
//
// Modifying the returned point does not modify the Canvas state.
func (c *Canvas) CurrentOffset() geom.Point {
	return c.offset
}

// FillRect records a filled rectangle with translated bounds and the supplied
// fill color.
func (c *Canvas) FillRect(rect geom.Rect, col color.Color) {
	c.Buffer.Push(Command{
		Type:   CmdFillRect,
		Bounds: rect.Offset(c.offset.X, c.offset.Y),
		Color:  col,
	})
}

// StrokeRect records a rectangle outline. The renderer interprets StrokeWidth
// and determines how the border is rasterized around the bounds.
func (c *Canvas) StrokeRect(rect geom.Rect, col color.Color, strokeWidth float64) {
	c.Buffer.Push(Command{
		Type:        CmdStrokeRect,
		Bounds:      rect.Offset(c.offset.X, c.offset.Y),
		StrokeColor: col,
		StrokeWidth: strokeWidth,
	})
}

// FillRoundedRect records a filled rectangle with rounded corners. The corner
// radius is preserved as supplied; validation or clamping belongs downstream.
func (c *Canvas) FillRoundedRect(rect geom.Rect, radius geom.CornerRadius, col color.Color) {
	c.Buffer.Push(Command{
		Type:   CmdFillRoundedRect,
		Bounds: rect.Offset(c.offset.X, c.offset.Y),
		Radius: radius,
		Color:  col,
	})
}

// StrokeRoundedRect records the outline of a rounded rectangle, including its
// translated bounds, corner radius, outline color, and stroke width.
func (c *Canvas) StrokeRoundedRect(rect geom.Rect, radius geom.CornerRadius, col color.Color, strokeWidth float64) {
	c.Buffer.Push(Command{
		Type:        CmdStrokeRoundedRect,
		Bounds:      rect.Offset(c.offset.X, c.offset.Y),
		Radius:      radius,
		StrokeColor: col,
		StrokeWidth: strokeWidth,
	})
}

// FillCircle records a filled circle. The center is translated, while radius
// remains unchanged because it describes a size rather than a position.
func (c *Canvas) FillCircle(center geom.Point, radius float64, col color.Color) {
	c.Buffer.Push(Command{
		Type:   CmdFillCircle,
		P1:     center.Add(c.offset),
		Radius: geom.RadiusUniform(radius),
		Color:  col,
	})
}

// StrokeCircle records a circle outline with translated center, radius, color,
// and stroke width.
func (c *Canvas) StrokeCircle(center geom.Point, radius float64, col color.Color, strokeWidth float64) {
	c.Buffer.Push(Command{
		Type:        CmdStrokeCircle,
		P1:          center.Add(c.offset),
		Radius:      geom.RadiusUniform(radius),
		StrokeColor: col,
		StrokeWidth: strokeWidth,
	})
}

// DrawLine records a line between two translated points. Both endpoints receive
// the same offset, preserving the line's shape while moving it as a unit.
func (c *Canvas) DrawLine(p1, p2 geom.Point, col color.Color, strokeWidth float64) {
	c.Buffer.Push(Command{
		Type:        CmdLine,
		P1:          p1.Add(c.offset),
		P2:          p2.Add(c.offset),
		StrokeColor: col,
		StrokeWidth: strokeWidth,
	})
}

// DrawText records text at a translated origin.
//
// Text measurement, font selection, baseline interpretation, and glyph
// rasterization are responsibilities of the renderer and text subsystem.
func (c *Canvas) DrawText(text string, origin geom.Point, fontSize float64, fontWeight int, col color.Color) {
	c.Buffer.Push(Command{
		Type:       CmdText,
		P1:         origin.Add(c.offset),
		Text:       text,
		FontSize:   fontSize,
		FontWeight: fontWeight,
		Color:      col,
	})
}

// DrawImage records an image and its translated destination rectangle.
//
// The image is stored as an image.Image reference; Canvas does not copy or
// decode its pixels. Scaling and filtering are decided by the renderer.
func (c *Canvas) DrawImage(img image.Image, dest geom.Rect) {
	c.Buffer.Push(Command{
		Type:   CmdImage,
		Bounds: dest.Offset(c.offset.X, c.offset.Y),
		Image:  img,
	})
}

// DrawShadow records shadow parameters for a rounded rectangle.
//
// Canvas records the description only. Blur, spread, compositing, and the
// final shadow shape are calculated by the renderer.
func (c *Canvas) DrawShadow(bounds geom.Rect, radius geom.CornerRadius, shadow ShadowParams) {
	c.Buffer.Push(Command{
		Type:   CmdShadow,
		Bounds: bounds.Offset(c.offset.X, c.offset.Y),
		Radius: radius,
		Shadow: shadow,
	})
}

// PushClip records the beginning of a rectangular clipping scope.
//
// The clip is represented as a command so the renderer can apply it to later
// commands in stream order. It is separate from Canvas's translation stack.
func (c *Canvas) PushClip(rect geom.Rect) {
	c.Buffer.Push(Command{
		Type:   CmdPushClip,
		Bounds: rect.Offset(c.offset.X, c.offset.Y),
	})
}

// PopClip records the end of the current clipping scope.
//
// The renderer is responsible for maintaining its clip stack. This method does
// not alter Canvas's coordinate offset or Save/Restore stack.
func (c *Canvas) PopClip() {
	c.Buffer.Push(Command{
		Type: CmdPopClip,
	})
}
