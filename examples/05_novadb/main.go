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
		nova.Title("NovaDB — Native Go Database Client"),
		nova.Size(1200, 800),
		nova.Theme(theme.Dark()),
	)

	// State
	queryText := state.String("SELECT id, name, email, department, salary, status\nFROM employees\nWHERE status = 'active'\nORDER BY id ASC\nLIMIT 100000;")
	rowCount := state.Int(100_000)
	statusMsg := state.String("Query executed in 1.4ms — 100,000 rows loaded.")

	// Sidebar tree: Connections and Tables
	sidebar := widgets.Sidebar("NovaDB",
		widgets.SidebarItem{Title: "postgres://prod-db", Icon: "#", Selected: true},
		widgets.SidebarItem{Title: "mysql://analytics", Icon: "#"},
		widgets.SidebarItem{Title: "sqlite://local.db", Icon: "$"},
		widgets.SidebarItem{Title: "Tables (4)", Icon: "="},
		widgets.SidebarItem{Title: "  | users", Icon: "-"},
		widgets.SidebarItem{Title: "  | employees", Icon: "-", Selected: true},
		widgets.SidebarItem{Title: "  | orders", Icon: "-"},
		widgets.SidebarItem{Title: "  | transactions", Icon: "-"},
	)

	// Top SQL Editor Box
	editorBox := widgets.Card("SQL Query Editor",
		ui.Column(
			widgets.CodeEditor(queryText.Get(), "sql").Bind(queryText),
			ui.Row(
				widgets.Button("▶ Run Query (Ctrl+Enter)").OnClick(func() {
					statusMsg.Set("Query re-executed in 0.9ms — 100,000 rows refreshed.")
				}),
				widgets.Button("Format SQL").Secondary(),
				widgets.Button("Explain Plan").Secondary(),
				widgets.Badge("100,000 Rows").Success(),
				widgets.Badge("1.4 ms").Info(),
			).GapSpacing(8),
		).GapSpacing(10),
	)

	// Bottom Virtualized Results Table (100,000 rows)
	tableCols := []widgets.TableColumn{
		{Title: "ID", Width: 70, Field: "id"},
		{Title: "Name", Width: 180, Field: "name"},
		{Title: "Email Address", Width: 220, Field: "email"},
		{Title: "Department", Width: 150, Field: "dept"},
		{Title: "Salary", Width: 120, Field: "salary"},
		{Title: "Status", Width: 100, Field: "status"},
	}

	resultsTable := widgets.Card("Query Results",
		widgets.Table(tableCols, rowCount.Get(), func(row, col int) string {
			switch col {
			case 0:
				return fmt.Sprintf("#%d", row+1)
			case 1:
				names := []string{"Alex Johnson", "Sarah Connor", "Michael Scott", "Emily Davis", "David Kim"}
				return names[row%len(names)]
			case 2:
				return fmt.Sprintf("user_%d@enterprise.corp", row+1)
			case 3:
				depts := []string{"Engineering", "Finance", "Product", "Security", "Marketing"}
				return depts[row%len(depts)]
			case 4:
				return fmt.Sprintf("$%d,000", 90+(row%80))
			case 5:
				return "Active"
			default:
				return ""
			}
		}),
	)

	// Main workspace: Split between SQL editor and results table
	mainWorkspace := ui.Padding(geom.All(16),
		ui.Column(
			editorBox,
			resultsTable,
		).GapSpacing(16),
	)

	// Full application layout: Sidebar + Main workspace
	appLayout := widgets.SplitPane(widgets.SplitHorizontal, sidebar, mainWorkspace)

	win.Content(appLayout)

	fmt.Println("🚀 Running NovaDB Database Client showcase...")
	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	_ = win.SaveScreenshot("novadb_showcase.png")
	fmt.Println("✅ Saved screenshot to novadb_showcase.png")
}
