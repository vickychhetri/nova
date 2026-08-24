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

// Use Case: User Registration, Input Validation, Passwords, and Form Submissions.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 05 Text Inputs & Forms"),
		nova.Size(850, 700),
		nova.Theme(theme.Dark()),
	)

	emailState := state.String("")
	passState := state.String("")
	feedbackState := state.String("Ready for input")

	emailField := forms.TextField("name@domain.com").
		WithWidth(360).
		Bind(emailState)

	passField := forms.PasswordField("Enter secure password").
		WithWidth(360).
		Bind(passState)

	var rebuild func()

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("05. Text Inputs & Form Validation").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Forms").Success(),
				).GapSpacing(10),
				ui.Text("Demonstrates text input fields, password masking, focus indicators, and form state.").Size(13).Col(color.Hex("#94A3B8")),

				// Form Card
				widgets.Card("User Authentication Form",
					ui.Column(
						ui.Text("Email Address").Size(12).Weight(font.WeightMedium).Col(color.Hex("#CBD5E1")),
						emailField,

						ui.Text("Password").Size(12).Weight(font.WeightMedium).Col(color.Hex("#CBD5E1")),
						passField,

						ui.Row(
							widgets.Button("Submit Form").Primary().OnClick(func() {
								email := emailState.Get()
								pass := passState.Get()
								if email == "" || pass == "" {
									feedbackState.Set("Error: Both email and password fields are required!")
								} else {
									feedbackState.Set(fmt.Sprintf("Success! Logged in as %s (Password: %d chars)", email, len(pass)))
								}
								rebuild()
							}),
							widgets.Button("Clear").Secondary().OnClick(func() {
								emailState.Set("")
								passState.Set("")
								feedbackState.Set("Form cleared")
								rebuild()
							}),
						).GapSpacing(10),

						ui.Container().
							Bg(color.Hex("#0F172A")).
							Pad(geom.All(10)).
							Rounded(geom.RadiusUniform(6)).
							WithChild(
								ui.Text(feedbackState.Get()).Size(12).Col(color.Hex("#38BDF8")),
							),
					).GapSpacing(10),
				),
			).GapSpacing(16),
		)
	}

	rebuild = func() {
		win.Content(buildUI())
	}

	rebuild()
	_ = win.SaveScreenshot("cookbook_05_forms.png")
	_ = app.Run()
}
