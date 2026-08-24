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
	"github.com/vickychhetri/nova/widgets/forms"
)

// Use Case: Application Preferences, Checkboxes, Audio/Display Sliders.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 06 Selection Controls & Sliders"),
		nova.Size(850, 700),
		nova.Theme(theme.Dark()),
	)

	enableSounds := state.Bool(true)
	darkTheme := state.Bool(true)
	volumeVal := state.Float(75.0)

	var rebuild func()

	volumeSlider := forms.Slider(0, 100).
		WithWidth(360).
		WithStep(1).
		Bind(volumeVal)
	volumeSlider.OnChanged = func(v float64) {
		if rebuild != nil {
			rebuild()
		}
	}

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("06. Selection Controls & Sliders").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Selection").Info(),
				).GapSpacing(10),
				ui.Text("Showcases interactive Checkboxes, Toggle state bindings, and continuous Sliders.").Size(13).Col(color.Hex("#94A3B8")),

				// Checkbox Toggles Card
				widgets.Card("Checkbox Preferences",
					ui.Column(
						forms.Checkbox("Enable Background Sound Effects").
							Bind(enableSounds).
							OnChange(func(val bool) {
								enableSounds.Set(val)
								rebuild()
							}),
						forms.Checkbox("Enable High-Contrast Dark Theme").
							Bind(darkTheme).
							OnChange(func(val bool) {
								darkTheme.Set(val)
								rebuild()
							}),
					).GapSpacing(12),
				),

				// Continuous & Stepped Slider Card
				widgets.Card("Audio Master Volume Slider",
					ui.Column(
						ui.Row(
							ui.Text("Volume Level:").Size(13).Weight(font.WeightMedium).Col(color.Hex("#CBD5E1")),
							widgets.Badge(fmt.Sprintf("%.0f%%", volumeVal.Get())).Success(),
						).GapSpacing(10),
						volumeSlider,
					).GapSpacing(10),
				),

				// Summary Card
				widgets.Card("Active Configuration State",
					ui.Text(fmt.Sprintf("• Sound FX: %t\n• Dark Theme: %t\n• Volume: %.0f/100",
						enableSounds.Get(), darkTheme.Get(), volumeVal.Get())).Size(12).Col(color.Hex("#38BDF8")),
				),
			).GapSpacing(16),
		)
	}

	rebuild = func() {
		win.Content(buildUI())
	}

	rebuild()
	_ = win.SaveScreenshot("cookbook_06_selection.png")
	_ = app.Run()
}
