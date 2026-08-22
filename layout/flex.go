package layout

import (
	"math"

	"github.com/vickychhetri/nova/core/geom"
)

// FlexChildInput represents layout parameters for a child in a flex container.
type FlexChildInput struct {
	// Measure measures the child with given constraints and returns its intrinsic size.
	Measure func(constraints BoxConstraints) geom.Size
	// Flex factor (0 = fixed size, >0 = flexible/expanded)
	Flex float64
}

// FlexConfig configures a flex layout pass.
type FlexConfig struct {
	Direction        Axis
	MainAxis         MainAxisAlignment
	CrossAxis        CrossAxisAlignment
	Gap              float64
	AllowCrossShrink bool
}

// FlexResult contains the computed layout geometry for the flex container and its children.
type FlexResult struct {
	Size         geom.Size
	ChildBounds  []geom.Rect
}

// ComputeFlex performs flexbox layout on a set of children under the given constraints.
func ComputeFlex(constraints BoxConstraints, config FlexConfig, children []FlexChildInput) FlexResult {
	n := len(children)
	if n == 0 {
		return FlexResult{
			Size:        constraints.Constrain(geom.Sz(0, 0)),
			ChildBounds: nil,
		}
	}

	childSizes := make([]geom.Size, n)
	totalFlex := 0.0
	allocatedMain := 0.0
	maxCross := 0.0

	var maxMain float64
	var maxCrossConstraint float64
	if config.Direction == AxisHorizontal {
		maxMain = constraints.MaxWidth
		maxCrossConstraint = constraints.MaxHeight
	} else {
		maxMain = constraints.MaxHeight
		maxCrossConstraint = constraints.MaxWidth
	}

	// 1. Measure fixed children (Flex == 0)
	for i, ch := range children {
		if ch.Flex > 0 {
			totalFlex += ch.Flex
		} else {
			var childConstraint BoxConstraints
			if config.Direction == AxisHorizontal {
				childConstraint = BoxConstraints{
					MinWidth:  0,
					MaxWidth:  math.Max(0, maxMain-allocatedMain),
					MinHeight: 0,
					MaxHeight: maxCrossConstraint,
				}
				if config.CrossAxis == CrossStretch && constraints.HasBoundedHeight() {
					childConstraint.MinHeight = constraints.MaxHeight
				}
			} else {
				childConstraint = BoxConstraints{
					MinWidth:  0,
					MaxWidth:  maxCrossConstraint,
					MinHeight: 0,
					MaxHeight: math.Max(0, maxMain-allocatedMain),
				}
				if config.CrossAxis == CrossStretch && constraints.HasBoundedWidth() {
					childConstraint.MinWidth = constraints.MaxWidth
				}
			}

			sz := ch.Measure(childConstraint)
			childSizes[i] = sz

			if config.Direction == AxisHorizontal {
				allocatedMain += sz.Width
				maxCross = math.Max(maxCross, sz.Height)
			} else {
				allocatedMain += sz.Height
				maxCross = math.Max(maxCross, sz.Width)
			}
		}
	}

	// Total gaps
	totalGap := float64(n-1) * config.Gap
	allocatedMain += totalGap

	// 2. Measure flexible children
	remainingMain := math.Max(0, maxMain-allocatedMain)
	if totalFlex > 0 && maxMain < Infinity {
		spacePerFlex := remainingMain / totalFlex
		for i, ch := range children {
			if ch.Flex > 0 {
				childMain := spacePerFlex * ch.Flex
				var childConstraint BoxConstraints
				if config.Direction == AxisHorizontal {
					childConstraint = BoxConstraints{
						MinWidth:  childMain,
						MaxWidth:  childMain,
						MinHeight: 0,
						MaxHeight: maxCrossConstraint,
					}
					if config.CrossAxis == CrossStretch && constraints.HasBoundedHeight() {
						childConstraint.MinHeight = constraints.MaxHeight
					}
				} else {
					childConstraint = BoxConstraints{
						MinWidth:  0,
						MaxWidth:  maxCrossConstraint,
						MinHeight: childMain,
						MaxHeight: childMain,
					}
					if config.CrossAxis == CrossStretch && constraints.HasBoundedWidth() {
						childConstraint.MinWidth = constraints.MaxWidth
					}
				}

				sz := ch.Measure(childConstraint)
				childSizes[i] = sz

				if config.Direction == AxisHorizontal {
					allocatedMain += sz.Width
					maxCross = math.Max(maxCross, sz.Height)
				} else {
					allocatedMain += sz.Height
					maxCross = math.Max(maxCross, sz.Width)
				}
			}
		}
	}

	// Determine container size
	var containerSize geom.Size
	if config.Direction == AxisHorizontal {
		width := allocatedMain
		if constraints.HasBoundedWidth() && (totalFlex > 0 || constraints.IsTight()) {
			width = constraints.MaxWidth
		}
		height := maxCross
		if constraints.HasBoundedHeight() && constraints.IsTight() {
			height = constraints.MaxHeight
		}
		containerSize = constraints.Constrain(geom.Sz(width, height))
	} else {
		height := allocatedMain
		if constraints.HasBoundedHeight() && (totalFlex > 0 || constraints.IsTight()) {
			height = constraints.MaxHeight
		}
		width := maxCross
		if constraints.HasBoundedWidth() && constraints.IsTight() {
			width = constraints.MaxWidth
		}
		containerSize = constraints.Constrain(geom.Sz(width, height))
	}

	// 3. Compute child positions along main & cross axes
	childBounds := make([]geom.Rect, n)
	var mainContainerSize float64
	var crossContainerSize float64
	if config.Direction == AxisHorizontal {
		mainContainerSize = containerSize.Width
		crossContainerSize = containerSize.Height
	} else {
		mainContainerSize = containerSize.Height
		crossContainerSize = containerSize.Width
	}

	// Total content main size excluding flex filling
	totalUsedMain := 0.0
	for _, sz := range childSizes {
		if config.Direction == AxisHorizontal {
			totalUsedMain += sz.Width
		} else {
			totalUsedMain += sz.Height
		}
	}
	totalUsedMain += totalGap

	freeSpace := math.Max(0, mainContainerSize-totalUsedMain)

	mainOffset := 0.0
	itemSpacing := config.Gap

	switch config.MainAxis {
	case MainStart:
		mainOffset = 0
	case MainCenter:
		mainOffset = freeSpace / 2.0
	case MainEnd:
		mainOffset = freeSpace
	case MainSpaceBetween:
		if n > 1 {
			itemSpacing = config.Gap + freeSpace/float64(n-1)
		}
	case MainSpaceAround:
		if n > 0 {
			spacing := freeSpace / float64(n)
			mainOffset = spacing / 2.0
			itemSpacing = config.Gap + spacing
		}
	case MainSpaceEvenly:
		if n > 0 {
			spacing := freeSpace / float64(n+1)
			mainOffset = spacing
			itemSpacing = config.Gap + spacing
		}
	}

	for i, sz := range childSizes {
		var childMain, childCross float64
		if config.Direction == AxisHorizontal {
			childMain = sz.Width
			childCross = sz.Height
		} else {
			childMain = sz.Height
			childCross = sz.Width
		}

		var crossOffset float64
		switch config.CrossAxis {
		case CrossStart:
			crossOffset = 0
		case CrossCenter:
			crossOffset = (crossContainerSize - childCross) / 2.0
		case CrossEnd:
			crossOffset = crossContainerSize - childCross
		case CrossStretch:
			crossOffset = 0
			childCross = crossContainerSize
		}

		if config.Direction == AxisHorizontal {
			childBounds[i] = geom.NewRect(mainOffset, crossOffset, childMain, childCross)
		} else {
			childBounds[i] = geom.NewRect(crossOffset, mainOffset, childCross, childMain)
		}

		mainOffset += childMain + itemSpacing
	}

	return FlexResult{
		Size:        containerSize,
		ChildBounds: childBounds,
	}
}
