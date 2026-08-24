# 📖 Nova UI Component Cookbook & Developer Guide

A comprehensive, developer-focused reference and interactive gallery of standalone sample programs demonstrating **every UI component, layout pattern, and use case** available in the Nova GUI Engine.

---

## 📂 Cookbook Directory Structure

```
examples/11_ui_cookbook/
├── 01_buttons_actions/        # Button variants, ghost buttons, icon triggers, and click callbacks
├── 02_typography_text/        # Typography scales, weights (Regular/Medium/Bold), wrapping, and colors
├── 03_containers_cards/       # Custom container borders, padding, radii, backgrounds, and stat cards
├── 04_layout_flexbox/         # Flexbox Rows, Columns, Spacers, GapSpacing, and auto-alignment
├── 05_text_inputs_forms/      # Text fields, password masking, focus states, and form validation
├── 06_selection_controls/     # Checkboxes, boolean state bindings, continuous and stepped sliders
├── 07_data_tables/            # Tabular data grids, column widths, and alternating row styling
├── 08_virtual_lists/          # 50,000+ element viewport virtualized lists with 60 FPS scrolling
├── 09_feedback_modals/        # Status badges, progress meters, alerts (Success/Warning/Error), and dialogs
├── 10_navigation_menus/       # Workspace sidebars, multi-view tab switchers, and split panes
├── 11_code_editor/            # Interactive code editor with line numbers, syntax highlighting, and typing
├── 12_canvas_graphics/        # 2D vector canvas rendering, custom geometric charts, and sparklines
├── main.go                    # Master Interactive Component Gallery (all-in-one browser)
├── cookbook_test.go           # Automated test suite covering component rendering and event dispatching
└── README.md                  # This complete developer manual and API reference
```

---

## 🚀 Quick Start: Running Examples

### 1. Run the Master Interactive Gallery:
```bash
go run ./examples/11_ui_cookbook
```

### 2. Run Any Individual Use Case Program:
```bash
# 01. Buttons & Actions
go run ./examples/11_ui_cookbook/01_buttons_actions

# 05. Text Inputs & Forms
go run ./examples/11_ui_cookbook/05_text_inputs_forms

# 07. Data Tables
go run ./examples/11_ui_cookbook/07_data_tables

# 12. Custom Vector Graphics & Canvas
go run ./examples/11_ui_cookbook/12_canvas_graphics
```

---

## 🛠️ Complete Component API & Use Case Reference

---

### 1. Buttons & Action Triggers (`01_buttons_actions`)
**Use Case**: Primary actions, form submitters, toolbars, and counter controls.

```go
// Primary Action
widgets.Button("Submit Order").Primary().OnClick(func() {
    fmt.Println("Order submitted!")
})

// Secondary Action
widgets.Button("Cancel").Secondary().OnClick(func() { ... })

// Ghost Button
widgets.Button("Settings").Ghost().OnClick(func() { ... })
```

---

### 2. Typography & Text Hierarchy (`02_typography_text`)
**Use Case**: Headings, paragraphs, code labels, and muted subtitles.

```go
// Heading 1 (24px Bold)
ui.Text("Dashboard Overview").
    Size(24).
    Weight(font.WeightBold).
    Col(color.Hex("#38BDF8"))

// Body text with automatic layout constraints
ui.Text("Detailed operational metrics for worker nodes.").
    Size(13).
    Weight(font.WeightRegular).
    Col(color.Hex("#94A3B8"))
```

---

### 3. Containers & Cards (`03_containers_cards`)
**Use Case**: KPI metric cards, bordered sections, glassmorphism panels.

```go
// Custom Container with background, border, and rounded corners
ui.Container().
    Bg(color.Hex("#1E293B")).
    Border(color.Hex("#38BDF8"), 1.5).
    Pad(geom.All(16)).
    Rounded(geom.RadiusUniform(8)).
    WithWidth(280).
    WithChild(
        ui.Column(
            ui.Text("Active Subscriptions").Size(12).Col(color.Hex("#94A3B8")),
            ui.Text("1,842").Size(22).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
        ).GapSpacing(4),
    )

// Themed Card Component
widgets.Card("System Health", ui.Text("Nominal")).
    WithSubtitle("Updated 1 min ago")
```

---

### 4. Flexbox Layout Engine (`04_layout_flexbox`)
**Use Case**: Responsive multi-column forms, navigation bars, and space distribution.

```go
// Horizontal Row with automatic Spacer distribution
ui.Row(
    widgets.Button("Left Tool").Secondary(),
    ui.Spacer(), // Expands to push remaining items across available width
    widgets.Badge("Live Status").Success(),
    ui.Spacer(),
    widgets.Button("Right Action").Primary(),
).GapSpacing(12)

// Vertical Column with uniform child gap
ui.Column(
    itemA,
    itemB,
    itemC,
).GapSpacing(8)
```

---

### 5. Text Inputs & Forms (`05_text_inputs_forms`)
**Use Case**: User authentication, search bars, and validated data entry.

```go
emailState := state.String("")
passwordState := state.String("")

// Text Field
emailField := forms.TextField().
    WithPlaceholder("user@domain.com").
    WithWidth(360).
    Bind(emailState)

// Password Field (Masked characters)
passField := forms.Password().
    WithPlaceholder("Enter password").
    WithWidth(360).
    Bind(passwordState)
```

---

### 6. Selection Controls & Sliders (`06_selection_controls`)
**Use Case**: Preferences, audio volume sliders, brightness controllers.

```go
// Checkbox Toggle
enableSounds := state.Bool(true)
forms.Checkbox("Enable Sound Effects").
    Bind(enableSounds).
    OnChanged(func(checked bool) {
        fmt.Printf("Sounds: %v\n", checked)
    })

// Continuous Stepped Slider (Min, Max, Step)
volumeVal := state.Float(75.0)
slider := forms.Slider(0, 100).
    WithWidth(360).
    WithStep(1).
    Bind(volumeVal)
slider.OnChanged = func(val float64) {
    fmt.Printf("Volume changed: %.0f\n", val)
}
```

---

### 7. Data Tables & Grids (`07_data_tables`)
**Use Case**: Inventory lists, user tables, financial ledgers.

```go
cols := []data.TableColumn{
    {Title: "Node ID", Width: 120},
    {Title: "Hostname", Width: 220},
    {Title: "Status", Width: 120},
}

table := data.Table(cols, len(rows), func(row, col int) string {
    item := rows[row]
    switch col {
    case 0: return item.ID
    case 1: return item.Hostname
    case 2: return item.Status
    }
    return ""
})
```

---

### 8. Viewport-Virtualized Lists (`08_virtual_lists`)
**Use Case**: Infinite chat feeds, event log streamers (10,000+ items at 60 FPS).

```go
// Only items visible in the viewport are rendered
vList := data.VirtualList(50000, 48.0, func(index int) ui.Component {
    return ui.Row(
        widgets.Badge(fmt.Sprintf("#%d", index+1)).Info(),
        ui.Text(fmt.Sprintf("Event Trace item %d", index)),
    ).GapSpacing(8)
})
```

---

### 9. Feedback, Badges & Alerts (`09_feedback_modals`)
**Use Case**: Notifications, progress meters, and operational health badges.

```go
// Badges
widgets.Badge("SUCCESS").Success()
widgets.Badge("WARNING").Warning()
widgets.Badge("CRITICAL").Error()
widgets.Badge("INFO").Info()

// Progress Bar (0.0 to 1.0)
widgets.ProgressBar(0.75)

// Inline Alert Banners
feedback.Alert("System Online", "All services operational.", feedback.AlertSuccess)
feedback.Alert("High CPU Load", "Usage exceeds 85%.", feedback.AlertWarning)
feedback.Alert("Connection Lost", "Retrying in 5s...", feedback.AlertError)
```

---

### 10. Navigation, Sidebars & Tabs (`10_navigation_menus`)
**Use Case**: Complex desktop application workspaces and multi-view switchers.

```go
// Sidebar Navigation
sidebar := nav.Sidebar("Admin Portal",
    nav.SidebarItem{Title: "Dashboard", Icon: "📊", Selected: true},
    nav.SidebarItem{Title: "Analytics", Icon: "📈", Badge: "Live"},
    nav.SidebarItem{Title: "Security", Icon: "🔒", Badge: "2FA"},
)

// Multi-Pane Tabs
tabs := nav.Tabs(
    nav.TabItem{Title: "Overview", Content: overviewPane},
    nav.TabItem{Title: "Log Stream", Content: logsPane},
).Bind(activeTabIndex)
```

---

### 11. Interactive Code Editor (`11_code_editor`)
**Use Case**: In-app IDEs, developer scratchpads, and config editors.

```go
codeState := state.String("package main\n\nfunc main() {\n    println(\"Hello Nova!\")\n}")

codeEditor := editor.CodeEditor(codeState.Get(), "go").
    Bind(codeState)
```

---

### 12. Custom 2D Vector Canvas (`12_canvas_graphics`)
**Use Case**: Financial candlestick charts, audio equalizers, custom geometry, and 2D games.

```go
chart := widgets.Canvas(800, 260, func(canvas *render.Canvas, bounds geom.Rect) {
    canvas.PushClip(bounds)
    
    // Background & Borders
    canvas.FillRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#0B0F19"))
    canvas.StrokeRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#1E293B"), 1.5)
    
    // Vector lines and glowing nodes
    canvas.DrawLine(geom.Pt(10, 100), geom.Pt(200, 40), color.Hex("#38BDF8"), 2.5)
    canvas.FillCircle(geom.Pt(200, 40), 4.0, color.Hex("#F8FAFC"))
    
    canvas.PopClip()
})
```

---

## 🧪 Testing the Cookbook

Run the automated test suite across all cookbook components:
```bash
go test -v ./examples/11_ui_cookbook/...
```
