package benchmarks_test

import (
	"fmt"
	"testing"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/virtualization"
	"github.com/vickychhetri/nova/widgets"
)

// BenchmarkAppStartupAndMount measures the cost of constructing an App,
// creating one configured Window, mounting a small component tree, and running
// the application lifecycle once.
//
// Setup occurs inside the timed loop because object creation and mounting are
// part of the startup path this benchmark is intended to measure.
func BenchmarkAppStartupAndMount(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Build a fresh application for every iteration so state from one
		// startup does not affect the next measurement.
		app := nova.New()
		win := app.Window(
			nova.Title("Bench App"),
			nova.Size(1024, 768),
		)
		win.Content(
			ui.Column(
				ui.Text("Benchmark Header"),
				widgets.Button("Action"),
			),
		)
		// Run may return immediately in a headless or unavailable-display
		// environment; the benchmark intentionally measures the call as part
		// of application startup.
		_ = app.Run()
	}
}

// BenchmarkFlexLayout1000Children measures flex-layout computation for a large
// vertical list of fixed-size children. The child definitions and constraints
// are prepared once so the timed section focuses on layout computation.
func BenchmarkFlexLayout1000Children(b *testing.B) {
	children := make([]layout.FlexChildInput, 1000)
	for i := range children {
		// Each child has a deterministic measurement result, isolating flex
		// distribution and positioning from expensive child measurement work.
		children[i] = layout.FlexChildInput{
			Flex: 0,
			Measure: func(c layout.BoxConstraints) geom.Size {
				return geom.Sz(100, 30)
			},
		}
	}

	constraints := layout.Tight(geom.Sz(1000, 30000))
	cfg := layout.FlexConfig{
		Direction: layout.AxisVertical,
		MainAxis:  layout.MainStart,
		CrossAxis: layout.CrossCenter,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Recompute the same 1000-child layout to measure repeatable throughput.
		_ = layout.ComputeFlex(constraints, cfg, children)
	}
}

// BenchmarkReactiveStateMutation measures repeated writes to a reactive
// integer followed by evaluation of a derived string value.
//
// The derived value is read after every mutation so the benchmark includes the
// invalidation/recomputation path rather than measuring Set alone.
func BenchmarkReactiveStateMutation(b *testing.B) {
	val := state.Int(0)
	derived := state.Compute(func() string {
		return fmt.Sprintf("Val: %d", val.Get())
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Each iteration changes the source value, then forces the derived value
		// to be observed.
		val.Set(i)
		_ = derived.Get()
	}
}

// BenchmarkVirtualizer100kRows measures visible-range calculation for a
// virtualized data set containing 100,000 rows. Only the range calculation is
// timed; row construction and rendering are intentionally out of scope.
func BenchmarkVirtualizer100kRows(b *testing.B) {
	v := virtualization.NewVirtualizer(100_000, 32.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Vary the scroll position while keeping the viewport height fixed so
		// the benchmark exercises different visible ranges.
		_ = v.ComputeVisibleRange(float64(i%10000)*32.0, 768.0)
	}
}

// BenchmarkRenderPipelineBatch measures repeated command recording for a
// 500-command render batch. The initial rounded-rectangle batch is setup data;
// the timed loop clears and records a fresh set of rectangles into the reused
// command buffer.
func BenchmarkRenderPipelineBatch(b *testing.B) {
	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)

	for i := 0; i < 500; i++ {
		canvas.FillRoundedRect(geom.NewRect(float64(i*2), float64(i*2), 50, 50), geom.RadiusUniform(4), color.Blue)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear retains the buffer's backing storage, allowing this benchmark to
		// focus on command construction and append throughput.
		buf.Clear()
		for j := 0; j < 500; j++ {
			canvas.FillRect(geom.NewRect(float64(j), float64(j), 20, 20), color.Blue)
		}
	}
}
