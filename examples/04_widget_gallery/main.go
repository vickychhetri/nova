package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

func main() {
	app := nova.New()

	win := app.Window(
		nova.Title("Nova — Full Widget Gallery"),
		nova.Size(1100, 800),
		nova.Theme(theme.Dark()),
	)

	activeTab := state.Int(0)
	dialogOpen := state.Bool(false)

	// Left Sidebar
	sidebar := widgets.Sidebar("Nova Gallery",
		widgets.SidebarItem{Title: "Components", Icon: "🧩", Selected: true},
		widgets.SidebarItem{Title: "Navigation", Icon: "🧭"},
		widgets.SidebarItem{Title: "Feedback & Modals", Icon: "💬"},
		widgets.SidebarItem{Title: "Data & Tables", Icon: "📊", Badge: "Fast"},
		widgets.SidebarItem{Title: "Settings", Icon: "⚙️"},
	)

	// Main tabbed content
	tabs := widgets.Tabs(
		widgets.TabItem{
			Title: "Basic Components",
			Content: ui.Padding(geom.All(16),
				ui.Column(
					widgets.Card("Buttons & Badges",
						ui.Row(
							widgets.Button("Primary"),
							widgets.Button("Secondary").Secondary(),
							widgets.Button("Outline").Outline(),
							widgets.Button("Danger").Danger(),
							widgets.Badge("Active").Success(),
							widgets.Badge("Pending").Warning(),
							widgets.Avatar("NC"),
						).GapSpacing(12),
					),
					widgets.Card("Feedback Banners",
						ui.Column(
							widgets.Alert("Information Note", "Nova features zero external C dependencies for tests.", widgets.AlertInfo),
							widgets.Alert("Success Message", "GPU pipeline initialization complete.", widgets.AlertSuccess),
						).GapSpacing(8),
					),
				).GapSpacing(12),
			),
		},
		widgets.TabItem{
			Title: "Data & Virtualization",
			Content: ui.Padding(geom.All(16),
				widgets.Card("10,000 Row Virtual Table",
					widgets.Table(
						[]widgets.TableColumn{
							{Title: "ID", Width: 60},
							{Title: "Name", Width: 160},
							{Title: "Role", Width: 120},
							{Title: "Department", Width: 140},
							{Title: "Status", Width: 100},
						},
						10_000,
						func(row, col int) string {
							switch col {
							case 0:
								return fmt.Sprintf("#%05d", row+1)
							case 1:
								return fmt.Sprintf("Employee %d", row+1)
							case 2:
								roles := []string{"Engineer", "Designer", "Product Manager", "DevOps"}
								return roles[row%len(roles)]
							case 3:
								return "Engineering"
							case 4:
								return "Active"
							default:
								return ""
							}
						},
					),
				),
			),
		},
	).Bind(activeTab)

	split := widgets.SplitPane(widgets.SplitHorizontal, sidebar, tabs)

	win.Content(
		ui.Stack(
			split,
			widgets.Dialog("Nova Dialog", "This is an overlay modal component.", dialogOpen),
		),
	)

	fmt.Println("Running Widget Gallery example...")
	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	_ = win.SaveScreenshot("widget_gallery.png")
	fmt.Println("Saved screenshot to widget_gallery.png")
}
