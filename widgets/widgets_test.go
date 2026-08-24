package widgets_test

import (
	"fmt"
	"testing"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

// TestComprehensiveWidgetsSuite verifies that representative widgets from each
// facade category can be composed, mounted, laid out, and painted together.
// TestComprehensiveWidgetsSuite verifies that representative widgets from each
// facade category can be composed, mounted, laid out, and painted together.
func TestComprehensiveWidgetsSuite(t *testing.T) {
	// Reactive values exercise state-driven overlays and tab selection in the
	// same tree as static navigation and data widgets.
	dialogOpen := state.Bool(false)
	selectedTab := state.Int(0)

	// Virtualized table sample data (10,000 rows) keeps construction realistic
	// without allocating a component for every logical row.
	tableCols := []widgets.TableColumn{
		{Title: "ID", Width: 60, Field: "id"},
		{Title: "Username", Width: 140, Field: "username"},
		{Title: "Status", Width: 100, Field: "status"},
	}

	tableComp := widgets.Table(tableCols, 10_000, func(row, col int) string {
		switch col {
		case 0:
			return fmt.Sprintf("#%d", row+1)
		case 1:
			return fmt.Sprintf("user_%d@company.com", row+1)
		case 2:
			if row%2 == 0 {
				return "Active"
			}
			return "Offline"
		default:
			return ""
		}
	})

	// Combine several content modes so the facade is tested across cards, data,
	// and editor widgets.
	tabsComp := widgets.Tabs(
		widgets.TabItem{
			Title: "Overview",
			Content: widgets.Card("Metrics Overview",
				ui.Row(
					widgets.Badge("10,000 Users").Success(),
					widgets.Avatar("JD"),
					widgets.Progress(0.75),
				).GapSpacing(12),
			),
		},
		widgets.TabItem{
			Title:   "Data Explorer",
			Content: tableComp,
		},
		widgets.TabItem{
			Title:   "SQL Query",
			Content: widgets.CodeEditor("SELECT id, name FROM users WHERE active = true;", "sql"),
		},
	).Bind(selectedTab)

	// Build navigation separately, then compose it with the tab content in a
	// split pane.
	sidebarComp := widgets.Sidebar("Nova Admin",
		widgets.SidebarItem{Title: "Dashboard", Icon: "📊", Selected: true},
		widgets.SidebarItem{Title: "Users", Icon: "👥", Badge: "10K"},
		widgets.SidebarItem{Title: "Settings", Icon: "⚙️"},
	)

	appLayout := widgets.SplitPane(widgets.SplitHorizontal, sidebarComp, tabsComp)

	// Stack places the dialog above the main application layout.
	fullView := ui.Stack(
		appLayout,
		widgets.Dialog("Confirm Action", "Are you sure you want to proceed?", dialogOpen),
	)

	// Mount using an explicit theme and scale, matching normal application setup.
	node := ui.NewNode(fullView)
	node.Mount(ui.BuildContext{
		Theme: theme.Dark(),
		Scale: 1.0,
	})

	// Tight constraints verify that the full widget tree can occupy the expected
	// viewport dimensions.
	sz := node.Layout(layout.Tight(geom.Sz(1024, 768)))
	if sz.Width != 1024 || sz.Height != 768 {
		t.Fatalf("unexpected layout size: %s", sz)
	}

	// Painting should produce commands even though this test does not open a
	// native window or inspect rasterized pixels.
	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)
	node.Paint(canvas)

	if buf.Len() == 0 {
		t.Fatal("expected paint commands recorded for widget suite")
	}
}
