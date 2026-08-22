package render

import (
	"image"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
)

// Canvas provides an ergonomic 2D drawing API recording commands into a CommandBuffer.
type Canvas struct {
	Buffer *CommandBuffer
	offset geom.Point
	stack  []geom.Point
}

// NewCanvas creates a Canvas recording to the given CommandBuffer.
func NewCanvas(buf *CommandBuffer) *Canvas {
	return &Canvas{
		Buffer: buf,
		stack:  make([]geom.Point, 0, 16),
	}
}

// Save pushes current coordinate transform state onto stack.
func (c *Canvas) Save() {
	c.stack = append(c.stack, c.offset)
}

// Restore pops coordinate transform state from stack.
func (c *Canvas) Restore() {
	if len(c.stack) > 0 {
		c.offset = c.stack[len(c.stack)-1]
		c.stack = c.stack[:len(c.stack)-1]
	}
}

// Translate shifts all subsequent coordinate operations by (dx, dy).
func (c *Canvas) Translate(dx, dy float64) {
	c.offset.X += dx
	c.offset.Y += dy
}

// CurrentOffset returns current translation offset.
func (c *Canvas) CurrentOffset() geom.Point {
	return c.offset
}

// FillRect draws a filled rectangle.
func (c *Canvas) FillRect(rect geom.Rect, col color.Color) {
	c.Buffer.Push(Command{
		Type:   CmdFillRect,
		Bounds: rect.Offset(c.offset.X, c.offset.Y),
		Color:  col,
	})
}

// StrokeRect draws a stroked rectangle border.
func (c *Canvas) StrokeRect(rect geom.Rect, col color.Color, strokeWidth float64) {
	c.Buffer.Push(Command{
		Type:        CmdStrokeRect,
		Bounds:      rect.Offset(c.offset.X, c.offset.Y),
		StrokeColor: col,
		StrokeWidth: strokeWidth,
	})
}

// FillRoundedRect draws a filled rounded rectangle.
func (c *Canvas) FillRoundedRect(rect geom.Rect, radius geom.CornerRadius, col color.Color) {
	c.Buffer.Push(Command{
		Type:   CmdFillRoundedRect,
		Bounds: rect.Offset(c.offset.X, c.offset.Y),
		Radius: radius,
		Color:  col,
	})
}

// StrokeRoundedRect draws a stroked rounded rectangle border.
func (c *Canvas) StrokeRoundedRect(rect geom.Rect, radius geom.CornerRadius, col color.Color, strokeWidth float64) {
	c.Buffer.Push(Command{
		Type:        CmdStrokeRoundedRect,
		Bounds:      rect.Offset(c.offset.X, c.offset.Y),
		Radius:      radius,
		StrokeColor: col,
		StrokeWidth: strokeWidth,
	})
}

// FillCircle draws a filled circle.
func (c *Canvas) FillCircle(center geom.Point, radius float64, col color.Color) {
	c.Buffer.Push(Command{
		Type:   CmdFillCircle,
		P1:     center.Add(c.offset),
		Radius: geom.RadiusUniform(radius),
		Color:  col,
	})
}

// StrokeCircle draws a stroked circle outline.
func (c *Canvas) StrokeCircle(center geom.Point, radius float64, col color.Color, strokeWidth float64) {
	c.Buffer.Push(Command{
		Type:        CmdStrokeCircle,
		P1:          center.Add(c.offset),
		Radius:      geom.RadiusUniform(radius),
		StrokeColor: col,
		StrokeWidth: strokeWidth,
	})
}

// DrawLine draws a straight line between two points.
func (c *Canvas) DrawLine(p1, p2 geom.Point, col color.Color, strokeWidth float64) {
	c.Buffer.Push(Command{
		Type:        CmdLine,
		P1:          p1.Add(c.offset),
		P2:          p2.Add(c.offset),
		StrokeColor: col,
		StrokeWidth: strokeWidth,
	})
}

// DrawText draws a string of text at origin point.
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

// DrawImage draws an image within specified destination rect.
func (c *Canvas) DrawImage(img image.Image, dest geom.Rect) {
	c.Buffer.Push(Command{
		Type:   CmdImage,
		Bounds: dest.Offset(c.offset.X, c.offset.Y),
		Image:  img,
	})
}

// DrawShadow draws a drop shadow for a rectangle with corner radius.
func (c *Canvas) DrawShadow(bounds geom.Rect, radius geom.CornerRadius, shadow ShadowParams) {
	c.Buffer.Push(Command{
		Type:   CmdShadow,
		Bounds: bounds.Offset(c.offset.X, c.offset.Y),
		Radius: radius,
		Shadow: shadow,
	})
}

// PushClip activates a rectangular clipping region.
func (c *Canvas) PushClip(rect geom.Rect) {
	c.Buffer.Push(Command{
		Type:   CmdPushClip,
		Bounds: rect.Offset(c.offset.X, c.offset.Y),
	})
}

// PopClip restores previous clipping state.
func (c *Canvas) PopClip() {
	c.Buffer.Push(Command{
		Type: CmdPopClip,
	})
}
