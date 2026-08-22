package layout

import (
	"math"

	"github.com/vickychhetri/nova/core/geom"
)

// GridConfig defines parameters for a uniform or responsive grid layout.
type GridConfig struct {
	Columns      int
	ColumnGap    float64
	RowGap       float64
	ItemHeight   float64 // If > 0, fixed row height. If 0, determined by child measurement.
	AspectRatio  float64 // If > 0 and ItemHeight == 0, height = width / AspectRatio.
}

// GridChildInput represents a child inside a Grid.
type GridChildInput struct {
	Measure func(constraints BoxConstraints) geom.Size
}

// GridResult contains computed grid container size and child rectangles.
type GridResult struct {
	Size        geom.Size
	ChildBounds []geom.Rect
}

// ComputeGrid lays out children in rows and columns.
func ComputeGrid(constraints BoxConstraints, config GridConfig, children []GridChildInput) GridResult {
	n := len(children)
	if n == 0 {
		return GridResult{
			Size:        constraints.Constrain(geom.Sz(0, 0)),
			ChildBounds: nil,
		}
	}

	cols := config.Columns
	if cols < 1 {
		cols = 1
	}

	totalColGaps := float64(cols-1) * config.ColumnGap
	var colWidth float64
	if constraints.HasBoundedWidth() {
		colWidth = math.Max(0, (constraints.MaxWidth-totalColGaps)/float64(cols))
	} else {
		colWidth = 100 // fallback default
	}

	numRows := int(math.Ceil(float64(n) / float64(cols)))
	childBounds := make([]geom.Rect, n)
	rowHeights := make([]float64, numRows)

	// Measure rows
	for r := 0; r < numRows; r++ {
		rowH := config.ItemHeight
		if rowH == 0 && config.AspectRatio > 0 {
			rowH = colWidth / config.AspectRatio
		}

		if rowH == 0 {
			// Measure each child in this row to find max height
			for c := 0; c < cols; c++ {
				idx := r*cols + c
				if idx >= n {
					break
				}
				sz := children[idx].Measure(TightWidth(colWidth))
				rowH = math.Max(rowH, sz.Height)
			}
		}
		rowHeights[r] = rowH
	}

	// Calculate bounds
	y := 0.0
	for r := 0; r < numRows; r++ {
		h := rowHeights[r]
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if idx >= n {
				break
			}
			x := float64(c) * (colWidth + config.ColumnGap)
			childBounds[idx] = geom.NewRect(x, y, colWidth, h)
		}
		y += h + config.RowGap
	}

	totalH := y
	if numRows > 0 {
		totalH -= config.RowGap
	}

	totalW := float64(cols)*colWidth + totalColGaps
	containerSize := constraints.Constrain(geom.Sz(totalW, totalH))

	return GridResult{
		Size:        containerSize,
		ChildBounds: childBounds,
	}
}
