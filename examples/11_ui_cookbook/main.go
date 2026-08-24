package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/data"
	"github.com/vickychhetri/nova/widgets/editor"
	"github.com/vickychhetri/nova/widgets/feedback"
	"github.com/vickychhetri/nova/widgets/forms"
	"github.com/vickychhetri/nova/widgets/nav"
)

func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("Nova UI Cookbook - Master Interactive Component Gallery"),
		nova.Size(1160, 840),
		nova.Theme(theme.Dark()),
	)

	activeSection := state.String("buttons") // "buttons", "typography", "containers", "layout", "forms", "selection", "tables", "lists", "feedback", "navigation", "editor", "canvas"

	var rebuildView func()

	// Demo State variables
	btnClicks := state.Int(0)
	emailInput := state.String("developer@nova.dev")
	passInput := state.String("SecretPass123!")
	chkState := state.Bool(true)
	sliderVal := state.Float(68.0)
	progressVal := state.Float(0.72)
	codeSample := `// Nova Declarative Component Tree
func BuildDashboard() ui.Component {
    return ui.Padding(geom.All(16),
        ui.Row(
            widgets.Card("System Metrics", ui.Text("Nominal")),
            widgets.Button("Action").Primary(),
        ).GapSpacing(12),
    )
}`
	codeState := state.String(codeSample)

	buildMainUI := func() ui.Component {
		curSec := activeSection.Get()

		makeNavBtn := func(label, id string) *ui.ButtonComponent {
			if curSec == id {
				return widgets.Button(label).Primary()
			}
			btn := widgets.Button(label).Ghost()
			btn.OnClick(func() {
				activeSection.Set(id)
				rebuildView()
			})
			return btn
		}

		// Top Header
		headerBar := widgets.Card("",
			ui.Row(
				ui.Column(
					ui.Row(
						ui.Text("Nova UI Component Cookbook").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
						widgets.Badge("12 Complete Use Cases").Success(),
						widgets.Badge("Cookbook Gallery").Info(),
					).GapSpacing(8),
					ui.Text("Master reference guide & live sandbox for all Nova GUI widgets and layout patterns.").Size(12).Col(color.Hex("#94A3B8")),
				).GapSpacing(2),
				ui.Spacer(),
				widgets.Badge("v1.1.2 Engine").Secondary(),
			).GapSpacing(12),
		)

		// Sidebar Navigation
		sidebarNav := ui.Container().
			Bg(color.Hex("#0F172A")).
			Border(color.Hex("#1E293B"), 1.0).
			Pad(geom.All(12)).
			Rounded(geom.RadiusUniform(8)).
			WithWidth(250).
			WithChild(
				ui.Column(
					ui.Text("UI COMPONENTS").Size(11).Weight(font.WeightBold).Col(color.Hex("#64748B")),
					makeNavBtn("01. Buttons & Actions", "buttons"),
					makeNavBtn("02. Typography & Text", "typography"),
					makeNavBtn("03. Containers & Cards", "containers"),
					makeNavBtn("04. Layout & Flexbox", "layout"),
					makeNavBtn("05. Text Inputs & Forms", "forms"),
					makeNavBtn("06. Selection & Sliders", "selection"),
					makeNavBtn("07. Data Tables & Grids", "tables"),
					makeNavBtn("08. Virtualized Lists", "lists"),
					makeNavBtn("09. Feedback & Alerts", "feedback"),
					makeNavBtn("10. Navigation & Tabs", "navigation"),
					makeNavBtn("11. Code Editor", "editor"),
					makeNavBtn("12. Canvas & Graphics", "canvas"),
				).GapSpacing(4),
			)

		// Dynamic Section Content
		var sectionContent ui.Component

		switch curSec {
		case "buttons":
			sectionContent = ui.Column(
				ui.Row(
					ui.Text("01. Buttons & Actions").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: Command Bars & Triggers").Info(),
				).GapSpacing(8),
				widgets.Card("Interactive Button Variants",
					ui.Column(
						ui.Row(
							widgets.Button("Primary Button").Primary().OnClick(func() {
								btnClicks.Set(btnClicks.Get() + 1)
								rebuildView()
							}),
							widgets.Button("Secondary Button").Secondary().OnClick(func() {
								btnClicks.Set(btnClicks.Get() + 1)
								rebuildView()
							}),
							widgets.Button("Ghost Button").Ghost().OnClick(func() {
								btnClicks.Set(btnClicks.Get() + 1)
								rebuildView()
							}),
							widgets.Button("[ Reset ]").Secondary().OnClick(func() {
								btnClicks.Set(0)
								rebuildView()
							}),
						).GapSpacing(10),
						ui.Row(
							widgets.Badge(fmt.Sprintf("Total Clicks: %d", btnClicks.Get())).Success(),
							ui.Text("Click any button to trigger reactive state dispatch.").Size(12).Col(color.Hex("#94A3B8")),
						).GapSpacing(8),
					).GapSpacing(12),
				),
				widgets.Card("Developer Code Recipe",
					ui.Container().
						Bg(color.Hex("#0B0F19")).
						Pad(geom.All(12)).
						Rounded(geom.RadiusUniform(6)).
						WithWidth(800).
						WithChild(
							ui.Text(`// Button API Usage:
widgets.Button("Save Changes").
    Primary().
    OnClick(func() {
        // Handle action
    })`).Size(12).Col(color.Hex("#38BDF8")),
						),
				),
			).GapSpacing(12)

		case "typography":
			sectionContent = ui.Column(
				ui.Row(
					ui.Text("02. Typography & Text Hierarchy").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: Text Layouts & Labels").Info(),
				).GapSpacing(8),
				widgets.Card("Type Scale & Weights",
					ui.Column(
						ui.Text("Heading 1 (24px Bold)").Size(24).Weight(font.WeightBold).Col(color.Hex("#38BDF8")),
						ui.Text("Heading 2 (18px Semi-Bold)").Size(18).Weight(font.WeightBold).Col(color.Hex("#E2E8F0")),
						ui.Text("Body Text (13px Regular)").Size(13).Col(color.Hex("#94A3B8")),
						ui.Text("Caption Text (11px Muted)").Size(11).Col(color.Hex("#64748B")),
					).GapSpacing(6),
				),
				widgets.Card("Developer Code Recipe",
					ui.Container().
						Bg(color.Hex("#0B0F19")).
						Pad(geom.All(12)).
						Rounded(geom.RadiusUniform(6)).
						WithWidth(800).
						WithChild(
							ui.Text(`ui.Text("Title").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC"))`).Size(12).Col(color.Hex("#38BDF8")),
						),
				),
			).GapSpacing(12)

		case "containers":
			sectionContent = ui.Column(
				ui.Row(
					ui.Text("03. Containers & Cards").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: Metric Dashboards & Panels").Info(),
				).GapSpacing(8),
				ui.Row(
					ui.Container().
						Bg(color.Hex("#1E293B")).
						Border(color.Hex("#38BDF8"), 1.5).
						Pad(geom.All(16)).
						Rounded(geom.RadiusUniform(8)).
						WithWidth(260).
						WithChild(
							ui.Column(
								ui.Text("Active Nodes").Size(12).Col(color.Hex("#94A3B8")),
								ui.Text("1,248").Size(22).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
							).GapSpacing(4),
						),
					ui.Container().
						Bg(color.Hex("#1E293B")).
						Border(color.Hex("#10B981"), 1.5).
						Pad(geom.All(16)).
						Rounded(geom.RadiusUniform(8)).
						WithWidth(260).
						WithChild(
							ui.Column(
								ui.Text("Network Throughput").Size(12).Col(color.Hex("#94A3B8")),
								ui.Text("48.6 Gbps").Size(22).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
							).GapSpacing(4),
						),
				).GapSpacing(12),
			).GapSpacing(12)

		case "layout":
			sectionContent = ui.Column(
				ui.Row(
					ui.Text("04. Flexbox Layout Engine").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: Responsive Grid & Alignment").Info(),
				).GapSpacing(8),
				widgets.Card("Row & Column Distribution",
					ui.Column(
						ui.Row(
							ui.Container().Bg(color.Hex("#6366F1")).Pad(geom.All(10)).Rounded(geom.RadiusUniform(6)).WithChild(ui.Text("Left").Col(color.White)),
							ui.Spacer(),
							ui.Container().Bg(color.Hex("#38BDF8")).Pad(geom.All(10)).Rounded(geom.RadiusUniform(6)).WithChild(ui.Text("Center Spacer").Col(color.White)),
							ui.Spacer(),
							ui.Container().Bg(color.Hex("#10B981")).Pad(geom.All(10)).Rounded(geom.RadiusUniform(6)).WithChild(ui.Text("Right").Col(color.White)),
						),
					).GapSpacing(10),
				),
			).GapSpacing(12)

		case "forms":
			sectionContent = ui.Column(
				ui.Row(
					ui.Text("05. Text Inputs & Form Validation").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: User Input & Authentication").Info(),
				).GapSpacing(8),
				widgets.Card("Form Input Controls",
					ui.Column(
						ui.Text("Email Address").Size(12).Weight(font.WeightMedium).Col(color.Hex("#CBD5E1")),
						forms.TextField("user@domain.com").WithWidth(360).Bind(emailInput),
						ui.Text("Password").Size(12).Weight(font.WeightMedium).Col(color.Hex("#CBD5E1")),
						forms.PasswordField("Password").WithWidth(360).Bind(passInput),
						ui.Row(
							widgets.Button("Validate & Submit").Primary().OnClick(func() {
								rebuildView()
							}),
							widgets.Badge(fmt.Sprintf("Bound: %s", emailInput.Get())).Info(),
						).GapSpacing(10),
					).GapSpacing(8),
				),
			).GapSpacing(12)

		case "selection":
			slider := forms.Slider(0, 100).WithWidth(360).WithStep(1).Bind(sliderVal)
			slider.OnChanged = func(v float64) {
				if rebuildView != nil {
					rebuildView()
				}
			}

			sectionContent = ui.Column(
				ui.Row(
					ui.Text("06. Selection Controls & Sliders").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: Settings & Equalizers").Info(),
				).GapSpacing(8),
				widgets.Card("Checkbox & Slider Preferences",
					ui.Column(
						forms.Checkbox("Enable System Notifications").Bind(chkState).OnChange(func(v bool) {
							chkState.Set(v)
							rebuildView()
						}),
						ui.Row(
							ui.Text("Intensity Slider:").Size(13).Col(color.Hex("#CBD5E1")),
							widgets.Badge(fmt.Sprintf("%.0f%%", sliderVal.Get())).Success(),
						).GapSpacing(10),
						slider,
					).GapSpacing(10),
				),
			).GapSpacing(12)

		case "tables":
			type Item struct {
				ID, Name, Category, Price, Stock string
			}
			items := []Item{
				{"SKU-01", "Titanium Frame Case", "Hardware", "$249.00", "42 in stock"},
				{"SKU-02", "Nova GUI Developer License", "Software", "$199.00", "Instant"},
				{"SKU-03", "Precision Mechanical Switch", "Hardware", "$14.50", "128 in stock"},
				{"SKU-04", "Ultra-wide OLED Monitor", "Hardware", "$799.00", "8 in stock"},
			}
			cols := []data.TableColumn{
				{Title: "Item ID", Width: 100},
				{Title: "Product Name", Width: 220},
				{Title: "Category", Width: 120},
				{Title: "Price", Width: 100},
				{Title: "Stock Status", Width: 140},
			}
			table := data.Table(cols, len(items), func(r, c int) string {
				it := items[r]
				switch c {
				case 0:
					return it.ID
				case 1:
					return it.Name
				case 2:
					return it.Category
				case 3:
					return it.Price
				case 4:
					return it.Stock
				}
				return ""
			})

			sectionContent = ui.Column(
				ui.Row(
					ui.Text("07. Tabular Data & Grid").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: Inventory & Ledgers").Info(),
				).GapSpacing(8),
				widgets.Card("Product Inventory Table",
					ui.Container().
						WithWidth(820).
						WithHeight(260).
						WithChild(table),
				),
			).GapSpacing(12)

		case "lists":
			vList := data.VirtualList(10000, 44.0, func(idx int) ui.Component {
				return ui.Container().
					Bg(color.Hex("#0F172A")).
					Pad(geom.Insets{Top: 6, Bottom: 6, Left: 12, Right: 12}).
					Rounded(geom.RadiusUniform(4)).
					WithChild(
						ui.Row(
							widgets.Badge(fmt.Sprintf("#%d", idx+1)).Info(),
							ui.Text(fmt.Sprintf("Virtualized Record Item trace_seq_%06d", idx)).Size(12).Col(color.Hex("#F8FAFC")),
							ui.Spacer(),
							widgets.Badge("ACTIVE").Success(),
						).GapSpacing(10),
					)
			})

			sectionContent = ui.Column(
				ui.Row(
					ui.Text("08. Viewport-Virtualized Lists").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: 10,000+ Items at 60 FPS").Success(),
				).GapSpacing(8),
				widgets.Card("High-Performance Virtualized Stream",
					ui.Container().
						WithWidth(820).
						WithHeight(300).
						WithChild(vList),
				),
			).GapSpacing(12)

		case "feedback":
			sectionContent = ui.Column(
				ui.Row(
					ui.Text("09. Feedback, Badges & Alerts").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: User Alerts & Meters").Info(),
				).GapSpacing(8),
				widgets.Card("Status Badges & Progress",
					ui.Column(
						ui.Row(
							widgets.Badge("SUCCESS").Success(),
							widgets.Badge("WARNING").Warning(),
							widgets.Badge("ERROR").Error(),
							widgets.Badge("INFO").Info(),
						).GapSpacing(10),
						widgets.Progress(progressVal.Get()),
					).GapSpacing(10),
				),
				feedback.Alert("Cluster Deployment Complete", "All services successfully migrated to v1.1.2 without downtime.", feedback.AlertSuccess),
			).GapSpacing(12)

		case "navigation":
			sidebar := nav.Sidebar("Admin Portal",
				nav.SidebarItem{Title: "Overview", Icon: "📊", Selected: true},
				nav.SidebarItem{Title: "Security", Icon: "🔒", Badge: "2FA"},
			)
			sectionContent = ui.Column(
				ui.Row(
					ui.Text("10. Navigation & Sidebars").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: App Workspace Layout").Info(),
				).GapSpacing(8),
				widgets.Card("Navigation Sidebar Component",
					ui.Row(
						sidebar,
						ui.Container().
							Bg(color.Hex("#0F172A")).
							Pad(geom.All(16)).
							Rounded(geom.RadiusUniform(6)).
							WithWidth(520).
							WithChild(ui.Text("Active Workspace Body Pane").Size(13).Col(color.Hex("#94A3B8"))),
					).GapSpacing(12),
				),
			).GapSpacing(12)

		case "editor":
			codeEd := editor.CodeEditor(codeSample, "go").Bind(codeState)
			sectionContent = ui.Column(
				ui.Row(
					ui.Text("11. Interactive Code Editor").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: Syntax Highlighting & Scratchpads").Info(),
				).GapSpacing(8),
				widgets.Card("Code Editor Widget",
					ui.Container().
						WithWidth(820).
						WithHeight(300).
						WithChild(codeEd),
				),
			).GapSpacing(12)

		case "canvas":
			chart := widgets.Canvas(820, 240, func(canvas *render.Canvas, bounds geom.Rect) {
				canvas.PushClip(bounds)
				canvas.FillRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#0B0F19"))
				canvas.StrokeRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#1E293B"), 1.5)

				// Draw sine wave curve
				points := 40
				for i := 0; i < points-1; i++ {
					x1 := float64(i) * (820.0 / float64(points-1))
					y1 := 120.0 + 70.0*mathSin(float64(i)*0.25)
					x2 := float64(i+1) * (820.0 / float64(points-1))
					y2 := 120.0 + 70.0*mathSin(float64(i+1)*0.25)

					canvas.DrawLine(geom.Pt(x1, y1), geom.Pt(x2, y2), color.Hex("#38BDF8"), 2.5)
					canvas.FillCircle(geom.Pt(x1, y1), 3.5, color.Hex("#F8FAFC"))
				}
				canvas.PopClip()
			})

			sectionContent = ui.Column(
				ui.Row(
					ui.Text("12. Custom Canvas & Vector Graphics").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Use Case: Charts, Diagrams & 2D Engines").Info(),
				).GapSpacing(8),
				widgets.Card("Immediate Mode Vector Canvas", chart),
			).GapSpacing(12)
		}

		return ui.Padding(geom.All(20),
			ui.Column(
				headerBar,
				ui.Row(
					sidebarNav,
					ui.Container().
						WithWidth(850).
						WithChild(sectionContent),
				).GapSpacing(16),
			).GapSpacing(12),
		)
	}

	rebuildView = func() {
		win.Content(buildMainUI())
	}

	rebuildView()
	_ = win.SaveScreenshot("cookbook_master_gallery.png")

	fmt.Println("🚀 Nova UI Cookbook Master Gallery is running...")
	if err := app.Run(); err != nil {
		fmt.Printf("Error running gallery: %v\n", err)
	}
}

func mathSin(x float64) float64 {
	for x > 3.14159265 {
		x -= 2 * 3.14159265
	}
	for x < -3.14159265 {
		x += 2 * 3.14159265
	}
	return x - (x*x*x)/6.0 + (x*x*x*x*x)/120.0
}
