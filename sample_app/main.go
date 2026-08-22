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
	// 1. Initialize Nova Desktop Application
	app := nova.New()

	// 2. Configure Desktop Window (Title, Dimensions, Dark Theme)
	win := app.Window(
		nova.Title("Nova Enterprise Desktop — Sample Application"),
		nova.Size(1200, 820),
		nova.Theme(theme.Dark()),
	)

	// 3. Define Reactive State Signals
	activeNavTab := state.Int(0)
	showDialog := state.Bool(false)
	username := state.String("alex_developer")
	email := state.String("alex@enterprise.corp")
	password := state.String("SuperSecret123!")
	bio := state.String("Senior systems engineer building native Go desktop applications with Nova.")
	age := state.Float(29)
	country := state.String("us")
	volume := state.Float(85)
	enableNotifications := state.Bool(true)
	agreeTerms := state.Bool(true)
	favColor := state.New(color.Hex("#3B82F6"))
	sqlQuery := state.String("SELECT id, name, email, department, salary, status\nFROM enterprise_users\nWHERE active = true\nORDER BY id ASC;")

	// Form Validation State
	formState := forms.NewFormState()
	formState.RegisterField("username", forms.Required(), forms.MinLength(3))
	formState.RegisterField("email", forms.Required(), forms.Email())
	formState.RegisterField("password", forms.Required(), forms.MinLength(6))

	formState.OnSubmit = func(vals map[string]any) {
		fmt.Printf("✅ Form Submitted with data: %+v\n", vals)
	}

	// 4. Build Left Sidebar Navigation
	sidebar := widgets.Sidebar("Nova Desktop",
		widgets.SidebarItem{Title: "Dashboard", Icon: "📊", Selected: true},
		widgets.SidebarItem{Title: "Form Inputs", Icon: "📝", Badge: "12 inputs"},
		widgets.SidebarItem{Title: "Data Explorer", Icon: "🗄️", Badge: "100K"},
		widgets.SidebarItem{Title: "SQL Studio", Icon: "⚡"},
		widgets.SidebarItem{Title: "Settings", Icon: "⚙️"},
	)

	// 5. Build Tab 1: Dashboard & Analytics Overview
	dashboardTab := ui.Padding(geom.All(20),
		ui.Column(
			// Stat Metrics Row
			ui.Row(
				widgets.Card("Active Users",
					ui.Column(
						ui.Text("124,850").Size(22).Weight(700),
						ui.Row(
							widgets.Badge("+14.2%").Success(),
							ui.Text("vs last week"),
						).GapSpacing(8),
					).GapSpacing(6),
				),
				widgets.Card("Query Throughput",
					ui.Column(
						ui.Text("8.4M req/s").Size(22).Weight(700),
						ui.Row(
							widgets.Badge("GPU Accelerated").Info(),
							ui.Text("1.2ms latency"),
						).GapSpacing(8),
					).GapSpacing(6),
				),
				widgets.Card("Memory Baseline",
					ui.Column(
						ui.Text("32.4 MB").Size(22).Weight(700),
						ui.Row(
							widgets.Badge("Lightweight").Success(),
							ui.Text("Zero Chrome RAM"),
						).GapSpacing(8),
					).GapSpacing(6),
				),
			).GapSpacing(16),

			// System Status Banner & Progress Card
			widgets.Card("Framework Engine Status",
				ui.Column(
					widgets.Alert("GPU & Render Pipeline Active", "Pure Go rendering engine initialized with sub-millisecond layout passes.", widgets.AlertSuccess),
					ui.Row(
						ui.Column(
							ui.Text("CPU Utilization"),
							widgets.Progress(0.24),
						).GapSpacing(4),
						ui.Column(
							ui.Text("Memory Efficiency"),
							widgets.Progress(0.88),
						).GapSpacing(4),
					).GapSpacing(24),
				).GapSpacing(12),
			),

			// Quick Action Buttons
			ui.Row(
				widgets.Button("Open Confirmation Dialog").OnClick(func() {
					showDialog.Set(true)
				}),
				widgets.Button("Export PDF Report").Secondary(),
				widgets.Button("System Diagnostics").Outline(),
			).GapSpacing(12),
		).GapSpacing(16),
	)

	// 6. Build Tab 2: All Form Controls & Inputs Suite
	formsTab := ui.Padding(geom.All(20),
		widgets.Card("Enterprise Form Inputs & Live Validation",
			ui.Column(
				// Row 1: Text, Email, Password
				ui.Row(
					widgets.TextField("Enter username").
						WithLabel("Username").
						Bind(username),

					widgets.TextField("Enter email address").
						WithLabel("Work Email").
						Bind(email),

					widgets.PasswordField("Enter password").
						WithLabel("Security Password").
						Bind(password),
				).GapSpacing(16),

				// Row 2: TextArea, Select, Number Stepper
				ui.Row(
					widgets.TextArea("Enter profile biography...").
						WithLabel("User Biography").
						Bind(bio),

					ui.Column(
						widgets.Select(
							forms.SelectOption{Label: "United States (US)", Value: "us"},
							forms.SelectOption{Label: "United Kingdom (UK)", Value: "uk"},
							forms.SelectOption{Label: "Germany (DE)", Value: "de"},
							forms.SelectOption{Label: "India (IN)", Value: "in"},
							forms.SelectOption{Label: "Japan (JP)", Value: "jp"},
						).Bind(country),

						widgets.NumberInput(age.Get()).WithMinMax(18, 100),
					).GapSpacing(12),
				).GapSpacing(16),

				// Row 3: Sliders, DatePicker, ColorPicker
				ui.Row(
					ui.Column(
						ui.Text("Volume / Threshold Level"),
						widgets.Slider(0, 100).Bind(volume),
					).GapSpacing(4),

					widgets.DatePicker(),

					widgets.ColorPicker(favColor.Get()),
				).GapSpacing(16),

				// Row 4: Checkboxes, Switches, File Dropzone
				ui.Row(
					widgets.Checkbox("I agree to enterprise policies").Bind(agreeTerms),
					widgets.Switch("Enable real-time push alerts").Bind(enableNotifications),
					widgets.FilePicker("Upload config file (JSON/YAML)..."),
				).GapSpacing(20),

				// Row 5: Action Submit Button
				ui.Row(
					widgets.Button("Save & Submit Form").OnClick(func() {
						formState.Set("username", username.Get())
						formState.Set("email", email.Get())
						formState.Set("password", password.Get())
						if formState.Submit() {
							fmt.Println("🚀 Form validated and submitted successfully!")
						} else {
							fmt.Printf("❌ Validation errors: %+v\n", formState.Errors())
						}
					}),
					widgets.Button("Reset Fields").Secondary().OnClick(func() {
						username.Set("")
						email.Set("")
						password.Set("")
						formState.Reset()
					}),
				).GapSpacing(12),
			).GapSpacing(16),
		),
	)

	// 7. Build Tab 3: Big Data Table Explorer (100,000 Rows)
	tableCols := []widgets.TableColumn{
		{Title: "ID", Width: 70, Field: "id"},
		{Title: "Full Name", Width: 180, Field: "name"},
		{Title: "Email Address", Width: 240, Field: "email"},
		{Title: "Department", Width: 150, Field: "dept"},
		{Title: "Annual Salary", Width: 130, Field: "salary"},
		{Title: "Account Status", Width: 110, Field: "status"},
	}

	dataExplorerTab := ui.Padding(geom.All(20),
		widgets.Card("100,000 Rows Virtualized Table (13ns viewport math)",
			widgets.Table(tableCols, 100_000, func(row, col int) string {
				switch col {
				case 0:
					return fmt.Sprintf("#%d", row+1)
				case 1:
					names := []string{"Alex Mercer", "Diana Prince", "Bruce Wayne", "Clark Kent", "Barry Allen"}
					return names[row%len(names)]
				case 2:
					return fmt.Sprintf("user_%d@enterprise.corp", row+1)
				case 3:
					depts := []string{"Engineering", "Security", "Infrastructure", "Finance", "AI Research"}
					return depts[row%len(depts)]
				case 4:
					return fmt.Sprintf("$%d,500", 110+(row%90))
				case 5:
					return "Active"
				default:
					return ""
				}
			}),
		),
	)

	// 8. Build Tab 4: SQL Studio & Syntax Highlighted Code Editor
	sqlStudioTab := ui.Padding(geom.All(20),
		widgets.Card("Native SQL Code Studio",
			ui.Column(
				widgets.CodeEditor(sqlQuery.Get(), "sql").Bind(sqlQuery),
				ui.Row(
					widgets.Button("▶ Run Query (Ctrl+Enter)").OnClick(func() {
						fmt.Println("Executing SQL Query on database engine...")
					}),
					widgets.Button("Format Query").Secondary(),
					widgets.Button("Explain Execution Plan").Secondary(),
					widgets.Badge("PostgreSQL 16").Info(),
					widgets.Badge("0.9ms Execution").Success(),
				).GapSpacing(10),
			).GapSpacing(12),
		),
	)

	// 9. Assemble Main Tabs
	mainTabs := widgets.Tabs(
		widgets.TabItem{Title: "Dashboard", Content: dashboardTab},
		widgets.TabItem{Title: "Form Controls", Content: formsTab},
		widgets.TabItem{Title: "100K Table Explorer", Content: dataExplorerTab},
		widgets.TabItem{Title: "SQL Studio", Content: sqlStudioTab},
	).Bind(activeNavTab)

	// 10. Split View: Sidebar + Main Tabs Content
	splitView := widgets.SplitPane(widgets.SplitHorizontal, sidebar, mainTabs)

	// 11. Root View with Overlay Modal Dialog
	rootView := ui.Stack(
		splitView,
		widgets.Dialog("Confirm System Action", "Are you sure you want to execute this operation?", showDialog).
			OnOk(func() {
				fmt.Println("Confirmed action!")
				showDialog.Set(false)
			}),
	)

	win.Content(rootView)

	// 12. Run Desktop Application Loop
	fmt.Println("==========================================================")
	fmt.Println("🚀 Nova Enterprise Desktop Sample Application Running...")
	fmt.Println("   • Window Dimensions: 1200 x 820")
	fmt.Println("   • Active Theme:      Dark Mode (Design Tokens)")
	fmt.Println("   • Features:          Forms, Inputs, Virtualized Table, SQL Studio")
	fmt.Println("==========================================================")

	if err := app.Run(); err != nil {
		fmt.Printf("Application Error: %v\n", err)
	}

	// Export initial frame preview
	_ = win.SaveScreenshot("app_preview.png")
	fmt.Println("📸 Initial window snapshot exported -> app_preview.png")
}
