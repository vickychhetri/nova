package layout

import (
	"math"

	"github.com/vickychhetri/nova/core/geom"
)

// BoxDecoration defines styling parameters that affect visual layout (border, padding, margin).
type BoxDecoration struct {
	Padding geom.Insets
	Margin  geom.Insets
	Width   *float64
	Height  *float64
	MinWidth  *float64
	MaxWidth  *float64
	MinHeight *float64
	MaxHeight *float64
}

// ComputeBoxLayout measures and positions a single child inside a decorated box container.
func ComputeBoxLayout(
	constraints BoxConstraints,
	deco BoxDecoration,
	measureChild func(childConstraints BoxConstraints) geom.Size,
) (containerSize geom.Size, childBounds geom.Rect) {
	effectiveConstraints := constraints.Deflate(deco.Margin)

	// Apply explicit width/height overrides
	if deco.Width != nil {
		effectiveConstraints.MinWidth = *deco.Width
		effectiveConstraints.MaxWidth = *deco.Width
	}
	if deco.Height != nil {
		effectiveConstraints.MinHeight = *deco.Height
		effectiveConstraints.MaxHeight = *deco.Height
	}
	if deco.MinWidth != nil {
		effectiveConstraints.MinWidth = math.Max(effectiveConstraints.MinWidth, *deco.MinWidth)
	}
	if deco.MaxWidth != nil {
		effectiveConstraints.MaxWidth = math.Min(effectiveConstraints.MaxWidth, *deco.MaxWidth)
	}
	if deco.MinHeight != nil {
		effectiveConstraints.MinHeight = math.Max(effectiveConstraints.MinHeight, *deco.MinHeight)
	}
	if deco.MaxHeight != nil {
		effectiveConstraints.MaxHeight = math.Min(effectiveConstraints.MaxHeight, *deco.MaxHeight)
	}

	innerConstraints := effectiveConstraints.Deflate(deco.Padding)

	var childSize geom.Size
	if measureChild != nil {
		childSize = measureChild(innerConstraints)
	}

	totalW := childSize.Width + deco.Padding.Horizontal()
	totalH := childSize.Height + deco.Padding.Vertical()

	boxSize := effectiveConstraints.Constrain(geom.Sz(totalW, totalH))

	childX := deco.Margin.Left + deco.Padding.Left
	childY := deco.Margin.Top + deco.Padding.Top

	childBounds = geom.NewRect(childX, childY, childSize.Width, childSize.Height)
	containerSize = geom.Sz(boxSize.Width+deco.Margin.Horizontal(), boxSize.Height+deco.Margin.Vertical())

	return containerSize, childBounds
}
