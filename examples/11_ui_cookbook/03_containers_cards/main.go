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

// Use Case: Dashboard Metric Cards, Glassmorphism Containers, and Nested Panels.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 03 Containers & Cards"),
		nova.Size(900, 650),
		nova.Theme(theme.Dark()),
	)

	buildUI := func() ui.Component {
		// Helper to create a stylish KPI stat metric card
		createStatCard := func(label, value, change string, isPositive bool) ui.Component {
			badge := widgets.Badge(change).Success()
			if !isPositive {
				badge = widgets.Badge(change).Error()
			}

			return ui.Container().
				Bg(color.Hex("#1E293B")).
				Border(color.Hex("#334155"), 1.0).
				Pad(geom.All(16)).
				Rounded(geom.RadiusUniform(10)).
				WithWidth(260).
				WithChild(
					ui.Column(
						ui.Row(
							ui.Text(label).Size(12).Weight(font.WeightMedium).Col(color.Hex("#94A3B8")),
							ui.Spacer(),
							badge,
						),
						ui.Text(value).Size(22).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					).GapSpacing(8),
				)
		}

		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("03. Containers & Cards").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Layout & Style").Info(),
				).GapSpacing(10),
				ui.Text("Demonstrates Container custom borders, radius, padding, background colors, and structured Cards.").Size(13).Col(color.Hex("#94A3B8")),

				// Metric Cards Row
				ui.Row(
					createStatCard("Total Revenue", "$124,500", "+14.2%", true),
					createStatCard("Active Subscriptions", "1,842", "+8.1%", true),
					createStatCard("Bounce Rate", "24.8%", "-3.4%", false),
				).GapSpacing(16),

				// Nested Cards Showcase
				widgets.Card("Standard Card Container with Subtitle",
					ui.Column(
						ui.Text("Cards provide pre-themed padding, borders, and structured title bars for standard application panels.").Size(13).Col(color.Hex("#E2E8F0")),
						ui.Row(
							ui.Container().
								Bg(color.Hex("#0F172A")).
								Border(color.Hex("#38BDF8"), 1.5).
								Pad(geom.All(12)).
								Rounded(geom.RadiusUniform(6)).
								WithChild(ui.Text("Custom Neon Cyan Border").Size(12).Col(color.Hex("#38BDF8"))),
							ui.Container().
								Bg(color.Hex("#0F172A")).
								Border(color.Hex("#10B981"), 1.5).
								Pad(geom.All(12)).
								Rounded(geom.RadiusUniform(6)).
								WithChild(ui.Text("Custom Emerald Border").Size(12).Col(color.Hex("#10B981"))),
							ui.Container().
								Bg(color.Hex("#0F172A")).
								Border(color.Hex("#F59E0B"), 1.5).
								Pad(geom.All(12)).
								Rounded(geom.RadiusUniform(6)).
								WithChild(ui.Text("Custom Amber Border").Size(12).Col(color.Hex("#F59E0B"))),
						).GapSpacing(12),
					).GapSpacing(10),
				).WithSubtitle("Pre-styled container with theme support"),
			).GapSpacing(16),
		)
	}

	win.Content(buildUI())
	_ = win.SaveScreenshot("cookbook_03_containers.png")
	_ = app.Run()
}
