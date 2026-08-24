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
)

// Use Case: Action Triggers, Counter State, and Interactive Command Buttons.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 01 Buttons & Actions"),
		nova.Size(800, 600),
		nova.Theme(theme.Dark()),
	)

	clickCount := state.Int(0)
	lastAction := state.String("None")

	var rebuild func()

	buildUI := func() ui.Component {
		count := clickCount.Get()
		action := lastAction.Get()

		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("01. Buttons & Actions Showcase").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Interactive").Success(),
				).GapSpacing(10),
				ui.Text("Demonstrates Primary, Secondary, Ghost, and Destructive button styles with reactive state bindings.").Size(13).Col(color.Hex("#94A3B8")),

				// Button Variants Card
				widgets.Card("Button Variants & Styling",
					ui.Column(
						ui.Row(
							widgets.Button("Primary Action").Primary().OnClick(func() {
								clickCount.Set(clickCount.Get() + 1)
								lastAction.Set("Clicked Primary Action")
								rebuild()
							}),
							widgets.Button("Secondary Action").Secondary().OnClick(func() {
								clickCount.Set(clickCount.Get() + 1)
								lastAction.Set("Clicked Secondary Action")
								rebuild()
							}),
							widgets.Button("Ghost Action").Ghost().OnClick(func() {
								clickCount.Set(clickCount.Get() + 1)
								lastAction.Set("Clicked Ghost Action")
								rebuild()
							}),
						).GapSpacing(12),
					).GapSpacing(10),
				),

				// Counter & Action Controller Card
				widgets.Card("State Modification & Counter",
					ui.Column(
						ui.Row(
							widgets.Button("[ - Decrease ]").Secondary().OnClick(func() {
								clickCount.Set(clickCount.Get() - 1)
								lastAction.Set("Decreased counter")
								rebuild()
							}),
							widgets.Badge(fmt.Sprintf("Total Clicks: %d", count)).Info(),
							widgets.Button("[ + Increase ]").Primary().OnClick(func() {
								clickCount.Set(clickCount.Get() + 1)
								lastAction.Set("Increased counter")
								rebuild()
							}),
							widgets.Button("[ Reset (0) ]").Ghost().OnClick(func() {
								clickCount.Set(0)
								lastAction.Set("Reset counter to 0")
								rebuild()
							}),
						).GapSpacing(12),
						ui.Text(fmt.Sprintf("Last Action Triggered: %s", action)).Size(12).Col(color.Hex("#38BDF8")),
					).GapSpacing(10),
				),
			).GapSpacing(16),
		)
	}

	rebuild = func() {
		win.Content(buildUI())
	}

	rebuild()
	_ = win.SaveScreenshot("cookbook_01_buttons.png")
	_ = app.Run()
}
