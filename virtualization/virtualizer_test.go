package virtualization_test

import (
	"testing"

	"github.com/vickychhetri/nova/virtualization"
)

func TestVirtualizer(t *testing.T) {
	const totalItems = 1_000_000
	const itemHeight = 30.0
	const viewportHeight = 600.0

	v := virtualization.NewVirtualizer(totalItems, itemHeight)

	// Test top position
	r0 := v.ComputeVisibleRange(0, viewportHeight)
	if r0.StartIndex != 0 {
		t.Fatalf("expected start index 0, got %d", r0.StartIndex)
	}
	// 600 / 30 = 20 visible + 3 overscan = 23
	if r0.EndIndex < 20 {
		t.Fatalf("expected at least 20 items visible, got end index %d", r0.EndIndex)
	}

	// Test scrolled down to middle: scrollOffset = 150,000 (item index ~ 5000)
	rMid := v.ComputeVisibleRange(150_000, viewportHeight)
	if rMid.StartIndex < 4990 || rMid.StartIndex > 5000 {
		t.Fatalf("unexpected start index for middle scroll: %d", rMid.StartIndex)
	}
}

func BenchmarkVirtualizer1MillionItems(b *testing.B) {
	const totalItems = 1_000_000
	v := virtualization.NewVirtualizer(totalItems, 30.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ComputeVisibleRange(float64(i%500_000), 800.0)
	}
}
