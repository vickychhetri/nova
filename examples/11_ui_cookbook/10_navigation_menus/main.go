package main

import (
	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/nav"
)

// Use Case: Multi-Pane Workspace, Application Sidebar, and Tab Navigation.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 10 Navigation & Workspaces"),
		nova.Size(950, 700),
		nova.Theme(theme.Dark()),
	)

	activeTab := state.Int(0)
	var rebuild func()

	sidebar := nav.Sidebar("Admin Console",
		nav.SidebarItem{Title: "Dashboard", Selected: true},
		nav.SidebarItem{Title: "Analytics", Badge: "Live"},
		nav.SidebarItem{Title: "Settings"},
		nav.SidebarItem{Title: "Security", Badge: "2FA"},
	)

	tabs := nav.Tabs(
		nav.TabItem{
			Title: "Overview Tab",
			Content: ui.Container().
				Bg(color.Hex("#0F172A")).
				Pad(geom.All(16)).
				Rounded(geom.RadiusUniform(8)).
				WithChild(
					ui.Column(
						ui.Text("Project Analytics & Performance Overview").Size(14).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
						ui.Text("System running with 99.99% uptime across 12 distributed worker nodes.").Size(12).Col(color.Hex("#94A3B8")),
					).GapSpacing(8),
				),
		},
		nav.TabItem{
			Title: "Logs & Traces",
			Content: ui.Container().
				Bg(color.Hex("#0F172A")).
				Pad(geom.All(16)).
				Rounded(geom.RadiusUniform(8)).
				WithChild(
					ui.Column(
						ui.Text("Real-Time Application Log Stream").Size(14).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
						ui.Text("[00:24:12] INFO: Request dispatched to auth-microservice-01 (latency: 12ms)").Size(11).Col(color.Hex("#38BDF8")),
						ui.Text("[00:24:15] INFO: Database connection pool recycled (active: 18, idle: 32)").Size(11).Col(color.Hex("#10B981")),
					).GapSpacing(6),
				),
		},
	).Bind(activeTab)

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("10. Navigation & Workspace Layouts").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Navigation").Info(),
				).GapSpacing(10),
				ui.Text("Demonstrates structured Sidebar navigation with item badges and multi-view Tab panels.").Size(13).Col(color.Hex("#94A3B8")),

				// Main Split Workspace
				ui.Row(
					sidebar,
					ui.Container().
						WithWidth(620).
						WithHeight(460).
						WithChild(
							ui.Column(
								widgets.Card("Workspace Content Switcher", tabs),
								ui.Row(
									widgets.Button("Switch to Overview").Primary().OnClick(func() {
										activeTab.Set(0)
										rebuild()
									}),
									widgets.Button("Switch to Logs").Secondary().OnClick(func() {
										activeTab.Set(1)
										rebuild()
									}),
								).GapSpacing(10),
							).GapSpacing(12),
						),
				).GapSpacing(16),
			).GapSpacing(16),
		)
	}

	rebuild = func() {
		win.Content(buildUI())
	}

	rebuild()
	_ = win.SaveScreenshot("cookbook_10_navigation.png")
	_ = app.Run()
}
