package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/feedback"
)

// Use Case: User Notifications, Progress Meters, Status Badges, and Alert Banners.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 09 Feedback & Status Alerts"),
		nova.Size(900, 700),
		nova.Theme(theme.Dark()),
	)

	progressVal := state.Float(0.65)
	statusText := state.String("Database sync in progress...")

	var rebuild func()

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("09. Feedback & Status Alerts").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Feedback").Info(),
				).GapSpacing(10),
				ui.Text("Showcases status Badges, Progress bars, Alert banners, and user state feedback.").Size(13).Col(color.Hex("#94A3B8")),

				// Status Badges
				widgets.Card("Status Badges",
					ui.Row(
						widgets.Badge("SUCCESS").Success(),
						widgets.Badge("WARNING").Warning(),
						widgets.Badge("CRITICAL ERROR").Error(),
						widgets.Badge("INFORMATION").Info(),
						widgets.Badge("SECONDARY / NEUTRAL").Secondary(),
					).GapSpacing(12),
				),

				// Progress Bar Card
				widgets.Card("System Progress Meter",
					ui.Column(
						ui.Row(
							ui.Text("Task Completion:").Size(13).Weight(font.WeightMedium).Col(color.Hex("#CBD5E1")),
							widgets.Badge(fmt.Sprintf("%.0f%%", progressVal.Get()*100)).Success(),
							ui.Spacer(),
							widgets.Button("[ - 10% ]").Secondary().OnClick(func() {
								val := progressVal.Get() - 0.10
								if val < 0 {
									val = 0
								}
								progressVal.Set(val)
								statusText.Set(fmt.Sprintf("Progress decreased to %.0f%%", val*100))
								rebuild()
							}),
							widgets.Button("[ + 10% ]").Primary().OnClick(func() {
								val := progressVal.Get() + 0.10
								if val > 1.0 {
									val = 1.0
								}
								progressVal.Set(val)
								statusText.Set(fmt.Sprintf("Progress increased to %.0f%%", val*100))
								rebuild()
							}),
						).GapSpacing(10),
						widgets.Progress(progressVal.Get()),
						ui.Text(statusText.Get()).Size(11).Col(color.Hex("#64748B")),
					).GapSpacing(10),
				),

				// Alert Banners Card
				widgets.Card("Alert & Notification Banners",
					ui.Column(
						feedback.Alert("System Online", "All cluster microservices responding normally (p99 < 15ms).", feedback.AlertSuccess),
						feedback.Alert("Memory Pressure Warning", "Pod node memory usage exceeded 85% threshold.", feedback.AlertWarning),
						feedback.Alert("Database Replication Out of Sync", "Secondary node latency lag is currently 1,420ms.", feedback.AlertError),
					).GapSpacing(10),
				),
			).GapSpacing(16),
		)
	}

	rebuild = func() {
		win.Content(buildUI())
	}

	rebuild()
	_ = win.SaveScreenshot("cookbook_09_feedback.png")
	_ = app.Run()
}
