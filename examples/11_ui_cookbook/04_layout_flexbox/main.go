package main

import (
	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

// Use Case: Flexbox Layout Engine, Rows, Columns, Alignment, and Spacer elements.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 04 Layout & Flexbox"),
		nova.Size(900, 650),
		nova.Theme(theme.Dark()),
	)

	buildUI := func() ui.Component {
		makeBox := func(label string, bgCol color.Color, width float64) ui.Component {
			return ui.Container().
				Bg(bgCol).
				Pad(geom.All(10)).
				Rounded(geom.RadiusUniform(6)).
				WithWidth(width).
				WithChild(ui.Center(ui.Text(label).Size(12).Weight(font.WeightBold).Col(color.Hex("#FFFFFF"))))
		}

		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("04. Flexbox Layout Engine").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Flex Layout").Info(),
				).GapSpacing(10),
				ui.Text("Showcases horizontal Rows, vertical Columns, proportional Spacers, and GapSpacing.").Size(13).Col(color.Hex("#94A3B8")),

				// Horizontal Flex Row with Spacers
				widgets.Card("Horizontal Row with Spacer Proportions",
					ui.Column(
						ui.Text("Spacer elements automatically expand to distribute remaining free space across children.").Size(12).Col(color.Hex("#94A3B8")),
						ui.Container().
							Bg(color.Hex("#0F172A")).
							Pad(geom.All(12)).
							Rounded(geom.RadiusUniform(8)).
							WithChild(
								ui.Row(
									makeBox("Left Start", color.Hex("#6366F1"), 120),
									ui.Spacer(),
									makeBox("Middle", color.Hex("#38BDF8"), 120),
									ui.Spacer(),
									makeBox("Right End", color.Hex("#10B981"), 120),
								),
							),
					).GapSpacing(10),
				),

				// Vertical Column with Gap Spacing
				widgets.Card("Vertical Column with Uniform Gaps",
					ui.Container().
						Bg(color.Hex("#0F172A")).
						Pad(geom.All(12)).
						Rounded(geom.RadiusUniform(8)).
						WithChild(
							ui.Column(
								makeBox("Row Block A (Gap: 8px)", color.Hex("#F59E0B"), 300),
								makeBox("Row Block B (Gap: 8px)", color.Hex("#EC4899"), 300),
								makeBox("Row Block C (Gap: 8px)", color.Hex("#8B5CF6"), 300),
							).GapSpacing(8),
						),
				),
			).GapSpacing(16),
		)
	}

	win.Content(buildUI())
	_ = win.SaveScreenshot("cookbook_04_layout.png")
	_ = app.Run()
}
