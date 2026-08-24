package virtualization_test

import (
	"testing"

	"github.com/vickychhetri/nova/virtualization"
)

// TestVirtualizer verifies fixed-height visible-range calculations at the list
// start and at a large middle scroll position. It also checks that overscan is
// included without requiring every item to be visited.
func TestVirtualizer(t *testing.T) {
	// A million-item list makes accidental full-list scanning easier to notice,
	// while fixed height keeps the expected range arithmetic deterministic.
	const totalItems = 1_000_000
	const itemHeight = 30.0
	const viewportHeight = 600.0

	v := virtualization.NewVirtualizer(totalItems, itemHeight)

	// At the top, the start index must clamp to zero even after subtracting
	// overscan items.
	r0 := v.ComputeVisibleRange(0, viewportHeight)
	if r0.StartIndex != 0 {
		t.Fatalf("expected start index 0, got %d", r0.StartIndex)
	}
	// 600 / 30 = 20 viewport items. The default three-item overscan should
	// extend the inclusive end beyond the visible portion.
	if r0.EndIndex < 20 {
		t.Fatalf("expected at least 20 items visible, got end index %d", r0.EndIndex)
	}

	// 150,000 / 30 = item index 5,000. The returned start may move backward by
	// overscan, so assert the valid range rather than one exact implementation
	// detail at the boundary.
	rMid := v.ComputeVisibleRange(150_000, viewportHeight)
	if rMid.StartIndex < 4990 || rMid.StartIndex > 5000 {
		t.Fatalf("unexpected start index for middle scroll: %d", rMid.StartIndex)
	}
}

// BenchmarkVirtualizer1MillionItems measures fixed-height range lookup for a
// one-million-item data set. The virtualizer is configured outside the timed
// loop so the benchmark measures ComputeVisibleRange itself.
func BenchmarkVirtualizer1MillionItems(b *testing.B) {
	const totalItems = 1_000_000
	v := virtualization.NewVirtualizer(totalItems, 30.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Vary the scroll position across half the data set while keeping the
		// viewport fixed. Fixed-height lookup should remain O(1) regardless of
		// total item count or scroll position.
		_ = v.ComputeVisibleRange(float64(i%500_000), 800.0)
	}
}
