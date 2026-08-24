package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/data"
)

// Use Case: Infinite Log Stream, Contact Directory, and Viewport-Virtualized Feeds.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 08 Virtualized Lists"),
		nova.Size(900, 700),
		nova.Theme(theme.Dark()),
	)

	const totalItems = 50000

	vList := data.VirtualList(totalItems, 48.0, func(idx int) ui.Component {
		statusBadge := widgets.Badge("INFO").Info()
		if idx%7 == 0 {
			statusBadge = widgets.Badge("WARN").Warning()
		} else if idx%13 == 0 {
			statusBadge = widgets.Badge("CRIT").Error()
		}

		return ui.Container().
			Bg(color.Hex("#0F172A")).
			Border(color.Hex("#1E293B"), 1.0).
			Pad(geom.Insets{Top: 8, Bottom: 8, Left: 14, Right: 14}).
			Rounded(geom.RadiusUniform(6)).
			WithChild(
				ui.Row(
					widgets.Badge(fmt.Sprintf("#%05d", idx+1)).Secondary(),
					ui.Column(
						ui.Text(fmt.Sprintf("Event Trace Log Item: transaction_ref_0x%08x", idx*31337)).Size(12).Weight(font.WeightMedium).Col(color.Hex("#F8FAFC")),
						ui.Text(fmt.Sprintf("Latency: %dms | Shard: cluster-node-%02d", 12+(idx%45), idx%8)).Size(10).Col(color.Hex("#64748B")),
					).GapSpacing(2),
					ui.Spacer(),
					statusBadge,
				).GapSpacing(10),
			)
	})

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("08. Viewport-Virtualized List").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge(fmt.Sprintf("%d,000 Elements", totalItems/1000)).Success(),
					widgets.Badge("60 FPS").Info(),
				).GapSpacing(10),
				ui.Text("Renders only elements currently inside viewport bounding box. Scroll with mouse wheel.").Size(13).Col(color.Hex("#94A3B8")),

				// Virtual List Container
				widgets.Card("System Audit Log (50,000 Virtualized Records)",
					ui.Container().
						WithWidth(820).
						WithHeight(400).
						WithChild(vList),
				),
			).GapSpacing(16),
		)
	}

	win.Content(buildUI())
	_ = win.SaveScreenshot("cookbook_08_virtual_list.png")
	_ = app.Run()
}
