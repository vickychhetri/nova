package virtualization

import (
	"math"
)

// Virtualizer calculates the visible window of items in large or infinite
// lists and tables.
//
// It can use one constant ItemHeight for O(1) range lookup or ItemHeightFn for
// variable-height items. Overscan expands the returned range so nearby items
// can be prepared before they enter the viewport.
type Virtualizer struct {
	// TotalCount is the total number of logical items.
	TotalCount int
	// ItemHeight is the fixed item height, or the fallback height for dynamic
	// mode when ItemHeightFn is nil for an item.
	ItemHeight float64
	// ItemHeightFn returns the height of an item at index.
	ItemHeightFn func(index int) float64
	// Overscan is the number of nearby indices included beyond the viewport.
	Overscan int
}

// NewVirtualizer creates a fixed-height virtualizer with the default overscan
// of three items.
//
// A non-positive itemHeight is retained in the struct; ComputeVisibleRange will
// use its dynamic/fallback path rather than performing fixed-height division.
func NewVirtualizer(totalCount int, itemHeight float64) *Virtualizer {
	return &Virtualizer{
		TotalCount: totalCount,
		ItemHeight: itemHeight,
		Overscan:   3,
	}
}

// TotalContentHeight returns the sum of all item heights.
//
// Fixed-height lists are calculated directly. Dynamic lists call ItemHeightFn
// once per item, so this operation is O(TotalCount) in that mode.
func (v *Virtualizer) TotalContentHeight() float64 {
	if v.TotalCount <= 0 {
		return 0
	}
	if v.ItemHeightFn == nil {
		// A constant height avoids scanning every item.
		return float64(v.TotalCount) * v.ItemHeight
	}
	total := 0.0
	for i := 0; i < v.TotalCount; i++ {
		// Dynamic totals intentionally follow the same per-index height source
		// used by visible-range calculation.
		total += v.ItemHeightFn(i)
	}
	return total
}

// VisibleRange describes the inclusive item-index range prepared for a
// viewport, including overscan items.
//
// An empty range is represented by the zero value. StartOffset is the logical
// y-position of StartIndex in the complete content, before the viewport's
// scroll offset is applied.
type VisibleRange struct {
	// StartIndex is the first prepared item index, inclusive.
	StartIndex int
	// EndIndex is the last prepared item index, inclusive.
	EndIndex int
	// StartOffset is the content-space vertical offset of StartIndex.
	StartOffset float64
}

// ComputeVisibleRange calculates the inclusive item range intersecting the
// viewport, expanded by Overscan on both sides.
//
// scrollOffset and viewportHeight are expressed in the same logical units as
// item heights. Empty input, a non-positive viewport, or no items returns the
// zero VisibleRange.
func (v *Virtualizer) ComputeVisibleRange(scrollOffset, viewportHeight float64) VisibleRange {
	if v.TotalCount <= 0 || viewportHeight <= 0 {
		return VisibleRange{}
	}

	if v.ItemHeightFn == nil && v.ItemHeight > 0 {
		// Fixed-height items can map scroll position directly to an index without
		// scanning preceding rows.
		firstVisible := int(math.Floor(scrollOffset / v.ItemHeight))
		visibleCount := int(math.Ceil(viewportHeight / v.ItemHeight))

		start := int(math.Max(0, float64(firstVisible-v.Overscan)))
		end := int(math.Min(float64(v.TotalCount-1), float64(firstVisible+visibleCount+v.Overscan)))

		// In fixed mode every item before start has the same height.
		startOffset := float64(start) * v.ItemHeight

		return VisibleRange{
			StartIndex:  start,
			EndIndex:    end,
			StartOffset: startOffset,
		}
	}

	// Dynamic heights require a forward scan because the offset of item i
	// depends on the sum of all preceding item heights.
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
			// The first item reaching the viewport becomes the visible start;
			// move backward by Overscan while clamping to index zero.
			start = int(math.Max(0, float64(i-v.Overscan)))
			startOffset = curY
		}

		if curY > scrollOffset+viewportHeight {
			// Once the scan passes the viewport, include trailing overscan and
			// stop; later items cannot be visible in this query.
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
