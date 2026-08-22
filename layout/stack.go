package layout

import (
	"math"

	"github.com/vickychhetri/nova/core/geom"
)

// StackPosition defines absolute positioning parameters for a child in a Stack.
type StackPosition struct {
	IsPositioned bool
	Top          *float64
	Right        *float64
	Bottom       *float64
	Left         *float64
	Width        *float64
	Height       *float64
}

// StackChildInput represents a child in a Stack layout.
type StackChildInput struct {
	Measure  func(constraints BoxConstraints) geom.Size
	Position StackPosition
}

// StackResult represents computed layout for a Stack container.
type StackResult struct {
	Size        geom.Size
	ChildBounds []geom.Rect
}

// ComputeStack calculates Stack layout for children, sizing by non-positioned children and positioning overlays.
func ComputeStack(constraints BoxConstraints, defaultAlign Alignment, children []StackChildInput) StackResult {
	n := len(children)
	if n == 0 {
		return StackResult{
			Size:        constraints.Constrain(geom.Sz(0, 0)),
			ChildBounds: nil,
		}
	}

	childSizes := make([]geom.Size, n)
	maxNonPositionedWidth := 0.0
	maxNonPositionedHeight := 0.0

	// 1. Measure non-positioned children to determine container size
	for i, ch := range children {
		if !ch.Position.IsPositioned {
			sz := ch.Measure(constraints.Deflate(geom.Insets{}))
			childSizes[i] = sz
			maxNonPositionedWidth = math.Max(maxNonPositionedWidth, sz.Width)
			maxNonPositionedHeight = math.Max(maxNonPositionedHeight, sz.Height)
		}
	}

	containerWidth := maxNonPositionedWidth
	if constraints.HasBoundedWidth() && constraints.IsTight() {
		containerWidth = constraints.MaxWidth
	}
	containerHeight := maxNonPositionedHeight
	if constraints.HasBoundedHeight() && constraints.IsTight() {
		containerHeight = constraints.MaxHeight
	}

	containerSize := constraints.Constrain(geom.Sz(containerWidth, containerHeight))

	// 2. Position all children
	childBounds := make([]geom.Rect, n)
	for i, ch := range children {
		if !ch.Position.IsPositioned {
			offset := defaultAlign.Align(containerSize, childSizes[i])
			childBounds[i] = geom.NewRect(offset.X, offset.Y, childSizes[i].Width, childSizes[i].Height)
		} else {
			pos := ch.Position
			var w, h float64
			if pos.Width != nil {
				w = *pos.Width
			}
			if pos.Height != nil {
				h = *pos.Height
			}

			// Horizontal constraints
			var x float64
			if pos.Left != nil && pos.Right != nil {
				x = *pos.Left
				w = math.Max(0, containerSize.Width-*pos.Left-*pos.Right)
			} else if pos.Left != nil {
				x = *pos.Left
			} else if pos.Right != nil {
				x = containerSize.Width - *pos.Right - w
			}

			// Vertical constraints
			var y float64
			if pos.Top != nil && pos.Bottom != nil {
				y = *pos.Top
				h = math.Max(0, containerSize.Height-*pos.Top-*pos.Bottom)
			} else if pos.Top != nil {
				y = *pos.Top
			} else if pos.Bottom != nil {
				y = containerSize.Height - *pos.Bottom - h
			}

			// Measure positioned child with computed width/height if needed
			if w == 0 || h == 0 {
				childConstraint := BoxConstraints{
					MinWidth:  0,
					MaxWidth:  containerSize.Width,
					MinHeight: 0,
					MaxHeight: containerSize.Height,
				}
				if pos.Width != nil {
					childConstraint.MinWidth = *pos.Width
					childConstraint.MaxWidth = *pos.Width
				}
				if pos.Height != nil {
					childConstraint.MinHeight = *pos.Height
					childConstraint.MaxHeight = *pos.Height
				}
				sz := ch.Measure(childConstraint)
				if pos.Width == nil {
					w = sz.Width
					if pos.Right != nil && pos.Left == nil {
						x = containerSize.Width - *pos.Right - w
					}
				}
				if pos.Height == nil {
					h = sz.Height
					if pos.Bottom != nil && pos.Top == nil {
						y = containerSize.Height - *pos.Bottom - h
					}
				}
			}

			childBounds[i] = geom.NewRect(x, y, w, h)
		}
	}

	return StackResult{
		Size:        containerSize,
		ChildBounds: childBounds,
	}
}
