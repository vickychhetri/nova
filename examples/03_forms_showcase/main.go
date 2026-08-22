package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/forms"
)

func main() {
	app := nova.New()

	win := app.Window(
		nova.Title("Nova — Complete Form Controls Showcase"),
		nova.Size(900, 750),
		nova.Theme(theme.Dark()),
	)

	// State signals
	username := state.String("")
	email := state.String("")
	password := state.String("")
	bio := state.String("")
	volume := state.Float(80)
	country := state.String("us")
	agreeTerms := state.Bool(false)
	newsletter := state.Bool(true)
	favColor := state.New(color.Hex("#3B82F6"))

	formState := forms.NewFormState()
	formState.RegisterField("username", forms.Required(), forms.MinLength(3))
	formState.RegisterField("email", forms.Required(), forms.Email())
	formState.RegisterField("password", forms.Required(), forms.MinLength(6))

	formState.OnSubmit = func(vals map[string]any) {
		fmt.Printf("🎉 Form submitted successfully! Data: %+v\n", vals)
	}

	win.Content(
		ui.Padding(geom.All(24),
			ui.Column(
				widgets.Card("User Registration & Preferences Form",
					ui.Column(
						// Row 1: Text Fields
						ui.Row(
							widgets.TextField("john_doe").
								WithLabel("Username").
								Bind(username),

							widgets.TextField("john@example.com").
								WithLabel("Email Address").
								Bind(email),

							widgets.PasswordField("••••••••").
								WithLabel("Password").
								Bind(password),
						).GapSpacing(16),

						// Row 2: Text Area & Stepper
						ui.Row(
							widgets.TextArea("Tell us about yourself...").
								WithLabel("Biography").
								Bind(bio),

							ui.Column(
								widgets.Select(
									forms.SelectOption{Label: "United States", Value: "us"},
									forms.SelectOption{Label: "United Kingdom", Value: "uk"},
									forms.SelectOption{Label: "Germany", Value: "de"},
									forms.SelectOption{Label: "India", Value: "in"},
									forms.SelectOption{Label: "Japan", Value: "jp"},
								).Bind(country),

								widgets.NumberInput(25).WithMinMax(18, 120),
							).GapSpacing(12),
						).GapSpacing(16),

						// Row 3: Sliders, Checkboxes & Toggles
						ui.Row(
							ui.Column(
								ui.Text("Volume / Sensitivity"),
								widgets.Slider(0, 100).Bind(volume),
							).GapSpacing(4),

							widgets.ColorPicker(favColor.Get()),

							widgets.DatePicker(),
						).GapSpacing(16),

						// Row 4: Toggles
						ui.Row(
							widgets.Checkbox("I agree to terms & conditions").Bind(agreeTerms),
							widgets.Switch("Subscribe to weekly newsletter").Bind(newsletter),
						).GapSpacing(24),

						// Row 5: Action buttons
						ui.Row(
							widgets.Button("Reset Form").Secondary().OnClick(func() {
								username.Set("")
								email.Set("")
								password.Set("")
								bio.Set("")
								agreeTerms.Set(false)
								formState.Reset()
							}),

							widgets.Button("Submit Application").OnClick(func() {
								formState.Set("username", username.Get())
								formState.Set("email", email.Get())
								formState.Set("password", password.Get())
								if formState.Submit() {
									fmt.Println("Form validated and processed!")
								} else {
									fmt.Printf("Validation errors: %+v\n", formState.Errors())
								}
							}),
						).GapSpacing(12),
					).GapSpacing(16),
				),
			).GapSpacing(16),
		),
	)

	fmt.Println("Running Forms Showcase...")
	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	_ = win.SaveScreenshot("forms_showcase.png")
	fmt.Println("Saved screenshot to forms_showcase.png")
}
