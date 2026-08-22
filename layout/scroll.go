package layout

import (
	"math"

	"github.com/vickychhetri/nova/core/geom"
)

// ScrollDirection defines scrolling axes.
type ScrollDirection int

const (
	ScrollVertical ScrollDirection = iota
	ScrollHorizontal
	ScrollBoth
)

// ScrollState manages scroll offset and bounds.
type ScrollState struct {
	OffsetX      float64
	OffsetY      float64
	ContentWidth float64
	ContentHeight float64
	ViewportWidth float64
	ViewportHeight float64
}

// MaxOffsetX returns maximum allowable horizontal scroll offset.
func (s ScrollState) MaxOffsetX() float64 {
	return math.Max(0, s.ContentWidth-s.ViewportWidth)
}

// MaxOffsetY returns maximum allowable vertical scroll offset.
func (s ScrollState) MaxOffsetY() float64 {
	return math.Max(0, s.ContentHeight-s.ViewportHeight)
}

// ClampOffsets returns clamped (X, Y) scroll offsets within valid bounds.
func (s ScrollState) ClampOffsets(x, y float64) (float64, float64) {
	clampedX := math.Max(0, math.Min(s.MaxOffsetX(), x))
	clampedY := math.Max(0, math.Min(s.MaxOffsetY(), y))
	return clampedX, clampedY
}

// VisibleRect returns the currently visible sub-rectangle in content coordinates.
func (s ScrollState) VisibleRect() geom.Rect {
	return geom.NewRect(s.OffsetX, s.OffsetY, s.ViewportWidth, s.ViewportHeight)
}
