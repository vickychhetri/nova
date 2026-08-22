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

func BenchmarkAppStartupAndMount(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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
		_ = app.Run()
	}
}

func BenchmarkFlexLayout1000Children(b *testing.B) {
	children := make([]layout.FlexChildInput, 1000)
	for i := range children {
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
		_ = layout.ComputeFlex(constraints, cfg, children)
	}
}

func BenchmarkReactiveStateMutation(b *testing.B) {
	val := state.Int(0)
	derived := state.Compute(func() string {
		return fmt.Sprintf("Val: %d", val.Get())
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Set(i)
		_ = derived.Get()
	}
}

func BenchmarkVirtualizer100kRows(b *testing.B) {
	v := virtualization.NewVirtualizer(100_000, 32.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ComputeVisibleRange(float64(i%10000)*32.0, 768.0)
	}
}

func BenchmarkRenderPipelineBatch(b *testing.B) {
	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)

	for i := 0; i < 500; i++ {
		canvas.FillRoundedRect(geom.NewRect(float64(i*2), float64(i*2), 50, 50), geom.RadiusUniform(4), color.Blue)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		for j := 0; j < 500; j++ {
			canvas.FillRect(geom.NewRect(float64(j), float64(j), 20, 20), color.Blue)
		}
	}
}
