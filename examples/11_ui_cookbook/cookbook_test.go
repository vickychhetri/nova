package main

import (
	"testing"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
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

func TestCookbook_ComponentsRendering(t *testing.T) {
	app := nova.New()
	win := app.Window(
		nova.Title("Test Cookbook"),
		nova.Size(800, 600),
		nova.Theme(theme.Dark()),
	)

	// Test 1: Buttons
	clicked := false
	btn := widgets.Button("Click Me").Primary().OnClick(func() {
		clicked = true
	})
	win.Content(btn)
	win.RenderFrame()
	win.DispatchPointerDown(geom.Pt(40, 20), 1)
	if !clicked {
		t.Errorf("Expected button click to trigger handler")
	}

	// Test 2: Typography & Containers
	container := ui.Container().
		Bg(color.Hex("#1E293B")).
		Border(color.Hex("#38BDF8"), 1.0).
		Pad(geom.All(16)).
		WithChild(ui.Text("Sample Heading").Size(18).Weight(font.WeightBold))
	win.Content(container)
	win.RenderFrame()

	// Test 3: Forms & Sliders
	strVal := state.String("test@domain.com")
	txtField := forms.TextField("Placeholder").Bind(strVal)
	passField := forms.PasswordField("Password")
	sliderVal := state.Float(50.0)
	slider := forms.Slider(0, 100).Bind(sliderVal)
	chk := forms.Checkbox("Check").Bind(state.Bool(true))
	win.Content(ui.Column(txtField, passField, slider, chk))
	win.RenderFrame()

	// Test 4: Tables & Lists
	cols := []data.TableColumn{{Title: "Col1", Width: 100}}
	tbl := data.Table(cols, 5, func(r, c int) string { return "Data" })
	vList := data.VirtualList(100, 30, func(idx int) ui.Component { return ui.Text("Item") })
	win.Content(ui.Column(tbl, vList))
	win.RenderFrame()

	// Test 5: Feedback, Nav, and Editor
	alert := feedback.Alert("Notice", "System ok", feedback.AlertSuccess)
	progress := widgets.Progress(0.5)
	sidebar := nav.Sidebar("App", nav.SidebarItem{Title: "Home"})
	codeEd := editor.CodeEditor("package main", "go")
	win.Content(ui.Column(alert, progress, sidebar, codeEd))
	win.RenderFrame()
}
