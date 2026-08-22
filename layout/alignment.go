package layout

import "github.com/vickychhetri/nova/core/geom"

// Axis defines layout direction.
type Axis int

const (
	AxisHorizontal Axis = iota
	AxisVertical
)

// MainAxisAlignment defines how children are placed along the primary axis.
type MainAxisAlignment int

const (
	MainStart MainAxisAlignment = iota
	MainCenter
	MainEnd
	MainSpaceBetween
	MainSpaceAround
	MainSpaceEvenly
)

// CrossAxisAlignment defines how children are placed along the cross axis.
type CrossAxisAlignment int

const (
	CrossStart CrossAxisAlignment = iota
	CrossCenter
	CrossEnd
	CrossStretch
)

// Alignment represents 2D fractional alignment [-1..1, -1..1].
type Alignment struct {
	X float64 // -1 (left), 0 (center), 1 (right)
	Y float64 // -1 (top), 0 (center), 1 (bottom)
}

// Common alignments
var (
	AlignTopLeft      = Alignment{X: -1, Y: -1}
	AlignTopCenter    = Alignment{X: 0, Y: -1}
	AlignTopRight     = Alignment{X: 1, Y: -1}
	AlignCenterLeft   = Alignment{X: -1, Y: 0}
	AlignCenter       = Alignment{X: 0, Y: 0}
	AlignCenterRight  = Alignment{X: 1, Y: 0}
	AlignBottomLeft   = Alignment{X: -1, Y: 1}
	AlignBottomCenter = Alignment{X: 0, Y: 1}
	AlignBottomRight  = Alignment{X: 1, Y: 1}
)

// Align calculates top-left offset to position a child inside parent size.
func (a Alignment) Align(parentSize, childSize geom.Size) geom.Point {
	freeWidth := parentSize.Width - childSize.Width
	freeHeight := parentSize.Height - childSize.Height

	x := (a.X + 1.0) / 2.0 * freeWidth
	y := (a.Y + 1.0) / 2.0 * freeHeight

	return geom.Pt(x, y)
}
