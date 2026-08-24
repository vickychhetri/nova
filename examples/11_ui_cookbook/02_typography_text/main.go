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

// Use Case: Typography Hierarchy, Font Weights, Colors, and Paragraph Layouts.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 02 Typography & Text"),
		nova.Size(850, 650),
		nova.Theme(theme.Dark()),
	)

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("02. Typography & Text Hierarchy").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Typography").Info(),
				).GapSpacing(10),
				ui.Text("Showcases scale levels, weight variations, color highlights, and auto-wrapping text.").Size(13).Col(color.Hex("#94A3B8")),

				// Font Scale Card
				widgets.Card("Type Hierarchy & Sizing Scale",
					ui.Column(
						ui.Text("Heading 1 (24px Bold)").Size(24).Weight(font.WeightBold).Col(color.Hex("#38BDF8")),
						ui.Text("Heading 2 (18px Semi-Bold)").Size(18).Weight(font.WeightBold).Col(color.Hex("#E2E8F0")),
						ui.Text("Subheading (15px Medium)").Size(15).Weight(font.WeightMedium).Col(color.Hex("#CBD5E1")),
						ui.Text("Body Text (13px Regular) - Standard reading text with high contrast for clean desktop UI.").Size(13).Col(color.Hex("#94A3B8")),
						ui.Text("Caption / Footnote (11px Light)").Size(11).Col(color.Hex("#64748B")),
					).GapSpacing(6),
				),

				// Font Weights & Emphasis Card
				widgets.Card("Weights & Emphasis",
					ui.Row(
						ui.Text("Regular Weight").Size(13).Weight(font.WeightRegular).Col(color.Hex("#E2E8F0")),
						ui.Text("Medium Weight").Size(13).Weight(font.WeightMedium).Col(color.Hex("#38BDF8")),
						ui.Text("Bold Weight").Size(13).Weight(font.WeightBold).Col(color.Hex("#10B981")),
						ui.Text("Accent Color").Size(13).Weight(font.WeightBold).Col(color.Hex("#F59E0B")),
						ui.Text("Error Alert").Size(13).Weight(font.WeightBold).Col(color.Hex("#EF4444")),
					).GapSpacing(16),
				),

				// Text Wrapping Card
				widgets.Card("Paragraph Layout & Text Wrapping",
					ui.Container().
						WithWidth(760).
						WithChild(
							ui.Text("Nova provides ultra-fast immediate-mode software rasterized typography rendering. Each glyph is extracted from TrueType vector outlines, rendered with subpixel alpha antialiasing, and laid out within bounding box constraints with automatic line wrapping and line breaks.").Size(12).Col(color.Hex("#CBD5E1")),
						),
				),
			).GapSpacing(16),
		)
	}

	win.Content(buildUI())
	_ = win.SaveScreenshot("cookbook_02_typography.png")
	_ = app.Run()
}
