package layout

import (
	"fmt"
	"math"

	"github.com/vickychhetri/nova/core/geom"
)

const Infinity = math.MaxFloat64

// BoxConstraints defines the minimum and maximum boundaries for width and height.
type BoxConstraints struct {
	MinWidth  float64
	MaxWidth  float64
	MinHeight float64
	MaxHeight float64
}

// Loose creates unconstrained loose bounds up to max size.
func Loose(size geom.Size) BoxConstraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  size.Width,
		MinHeight: 0,
		MaxHeight: size.Height,
	}
}

// Tight creates tight constraints where min and max are equal.
func Tight(size geom.Size) BoxConstraints {
	return BoxConstraints{
		MinWidth:  size.Width,
		MaxWidth:  size.Width,
		MinHeight: size.Height,
		MaxHeight: size.Height,
	}
}

// TightWidth creates tight width with loose height.
func TightWidth(width float64) BoxConstraints {
	return BoxConstraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: Infinity,
	}
}

// TightHeight creates tight height with loose width.
func TightHeight(height float64) BoxConstraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  Infinity,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Expand creates infinite maximum constraints.
func Expand() BoxConstraints {
	return BoxConstraints{
		MinWidth:  0,
		MaxWidth:  Infinity,
		MinHeight: 0,
		MaxHeight: Infinity,
	}
}

// Constrain restricts a size within these constraints.
func (c BoxConstraints) Constrain(size geom.Size) geom.Size {
	return geom.Size{
		Width:  math.Max(c.MinWidth, math.Min(c.MaxWidth, size.Width)),
		Height: math.Max(c.MinHeight, math.Min(c.MaxHeight, size.Height)),
	}
}

// ConstrainWidth restricts width within min and max.
func (c BoxConstraints) ConstrainWidth(width float64) float64 {
	return math.Max(c.MinWidth, math.Min(c.MaxWidth, width))
}

// ConstrainHeight restricts height within min and max.
func (c BoxConstraints) ConstrainHeight(height float64) float64 {
	return math.Max(c.MinHeight, math.Min(c.MaxHeight, height))
}

// Deflate shrinks constraints by insets (e.g. for padding).
func (c BoxConstraints) Deflate(insets geom.Insets) BoxConstraints {
	deflatedMinWidth := math.Max(0, c.MinWidth-insets.Horizontal())
	deflatedMaxWidth := math.Max(deflatedMinWidth, c.MaxWidth-insets.Horizontal())
	deflatedMinHeight := math.Max(0, c.MinHeight-insets.Vertical())
	deflatedMaxHeight := math.Max(deflatedMinHeight, c.MaxHeight-insets.Vertical())

	return BoxConstraints{
		MinWidth:  deflatedMinWidth,
		MaxWidth:  deflatedMaxWidth,
		MinHeight: deflatedMinHeight,
		MaxHeight: deflatedMaxHeight,
	}
}

// IsTight returns true if min and max equal for both dimensions.
func (c BoxConstraints) IsTight() bool {
	return c.MinWidth >= c.MaxWidth && c.MinHeight >= c.MaxHeight
}

// HasBoundedWidth returns true if MaxWidth is not infinite.
func (c BoxConstraints) HasBoundedWidth() bool {
	return c.MaxWidth < Infinity
}

// HasBoundedHeight returns true if MaxHeight is not infinite.
func (c BoxConstraints) HasBoundedHeight() bool {
	return c.MaxHeight < Infinity
}

// MaxSize returns maximum size allowed.
func (c BoxConstraints) MaxSize() geom.Size {
	return geom.Sz(c.MaxWidth, c.MaxHeight)
}

// MinSize returns minimum size allowed.
func (c BoxConstraints) MinSize() geom.Size {
	return geom.Sz(c.MinWidth, c.MinHeight)
}

func (c BoxConstraints) String() string {
	return fmt.Sprintf("BoxConstraints(w: %.1f..%.1f, h: %.1f..%.1f)", c.MinWidth, c.MaxWidth, c.MinHeight, c.MaxHeight)
}
