package virtualization

import (
	"math"
)

// Virtualizer calculates the visible window of items in large or infinite lists/tables.
type Virtualizer struct {
	TotalCount   int
	ItemHeight   float64
	ItemHeightFn func(index int) float64
	Overscan     int
}

// NewVirtualizer creates a fixed-height virtualizer.
func NewVirtualizer(totalCount int, itemHeight float64) *Virtualizer {
	return &Virtualizer{
		TotalCount: totalCount,
		ItemHeight: itemHeight,
		Overscan:   3,
	}
}

// TotalContentHeight returns the total height of all items combined.
func (v *Virtualizer) TotalContentHeight() float64 {
	if v.TotalCount <= 0 {
		return 0
	}
	if v.ItemHeightFn == nil {
		return float64(v.TotalCount) * v.ItemHeight
	}
	total := 0.0
	for i := 0; i < v.TotalCount; i++ {
		total += v.ItemHeightFn(i)
	}
	return total
}

// VisibleRange represents the slice of indices currently visible in the viewport.
type VisibleRange struct {
	StartIndex  int
	EndIndex    int
	StartOffset float64
}

// ComputeVisibleRange calculates which items fall within the viewport plus overscan buffer.
func (v *Virtualizer) ComputeVisibleRange(scrollOffset, viewportHeight float64) VisibleRange {
	if v.TotalCount <= 0 || viewportHeight <= 0 {
		return VisibleRange{}
	}

	if v.ItemHeightFn == nil && v.ItemHeight > 0 {
		// O(1) direct calculation for fixed item heights
		firstVisible := int(math.Floor(scrollOffset / v.ItemHeight))
		visibleCount := int(math.Ceil(viewportHeight / v.ItemHeight))

		start := int(math.Max(0, float64(firstVisible-v.Overscan)))
		end := int(math.Min(float64(v.TotalCount-1), float64(firstVisible+visibleCount+v.Overscan)))

		startOffset := float64(start) * v.ItemHeight

		return VisibleRange{
			StartIndex:  start,
			EndIndex:    end,
			StartOffset: startOffset,
		}
	}

	// Dynamic variable heights
	curY := 0.0
	start := -1
	end := 0
	startOffset := 0.0

	for i := 0; i < v.TotalCount; i++ {
		h := v.ItemHeight
		if v.ItemHeightFn != nil {
			h = v.ItemHeightFn(i)
		}

		if curY+h >= scrollOffset && start == -1 {
			start = int(math.Max(0, float64(i-v.Overscan)))
			startOffset = curY
		}

		if curY > scrollOffset+viewportHeight {
			end = int(math.Min(float64(v.TotalCount-1), float64(i+v.Overscan)))
			break
		}

		curY += h
		end = i
	}

	if start == -1 {
		start = 0
	}

	return VisibleRange{
		StartIndex:  start,
		EndIndex:    end,
		StartOffset: startOffset,
	}
}
