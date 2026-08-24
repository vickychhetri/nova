package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

// Use Case: Custom 2D Vector Visualizations, Financial Sparklines, and Real-time Telemetry Charts.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 12 Canvas & Vector Graphics"),
		nova.Size(950, 700),
		nova.Theme(theme.Dark()),
	)

	// Sample timeseries dataset (Stock / Metric values)
	dataPoints := []float64{120, 135, 128, 142, 160, 155, 178, 192, 185, 210, 230, 225, 248, 265, 258, 290}

	renderChart := func(w, h float64) ui.Component {
		return widgets.Canvas(w, h, func(canvas *render.Canvas, bounds geom.Rect) {
			canvas.PushClip(bounds)

			// Chart Canvas Background
			canvas.FillRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#0B0F19"))
			canvas.StrokeRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#1E293B"), 1.5)

			// Grid lines (horizontal)
			for i := 1; i <= 4; i++ {
				y := float64(i) * (h / 5.0)
				canvas.DrawLine(geom.Pt(0, y), geom.Pt(w, y), color.Hex("#1E293B").WithAlpha(0.6), 1.0)
			}

			// Render Trend Line & Data Nodes
			minVal := 100.0
			maxVal := 320.0
			stepX := w / float64(len(dataPoints)-1)

			var prevPt geom.Point
			for i, val := range dataPoints {
				normY := (val - minVal) / (maxVal - minVal)
				curY := h - 20 - normY*(h-40)
				curX := float64(i) * stepX
				curPt := geom.Pt(curX, curY)

				if i > 0 {
					// Draw glowing gradient line segment
					canvas.DrawLine(prevPt, curPt, color.Hex("#38BDF8"), 2.5)
				}

				// Data Node Point (Glow Circle)
				canvas.FillCircle(curPt, 4.5, color.Hex("#38BDF8"))
				canvas.FillCircle(curPt, 2.5, color.Hex("#F8FAFC"))

				prevPt = curPt
			}

			canvas.PopClip()
		})
	}

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("12. Custom Vector Graphics & Canvas API").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Vector Canvas").Success(),
					widgets.Badge("Immediate Mode").Info(),
				).GapSpacing(10),
				ui.Text("Direct software rasterizer canvas for custom geometry, charts, game engines, and diagrams.").Size(13).Col(color.Hex("#94A3B8")),

				// Chart Visualization Card
				widgets.Card("Real-Time Telemetry & Metric Trendline",
					ui.Column(
						renderChart(880, 280),
						ui.Row(
							widgets.Badge(fmt.Sprintf("Current Index: $%.2f", dataPoints[len(dataPoints)-1])).Success(),
							widgets.Badge(fmt.Sprintf("Min: $%.2f", 120.0)).Secondary(),
							widgets.Badge(fmt.Sprintf("Max: $%.2f", 290.0)).Secondary(),
							ui.Spacer(),
							ui.Text("Vector rendered at 60+ FPS via Nova CommandBuffer").Size(11).Col(color.Hex("#64748B")),
						).GapSpacing(8),
					).GapSpacing(12),
				),
			).GapSpacing(16),
		)
	}

	win.Content(buildUI())
	_ = win.SaveScreenshot("cookbook_12_canvas.png")
	_ = app.Run()
}
