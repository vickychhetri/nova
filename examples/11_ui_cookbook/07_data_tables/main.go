package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/data"
)

// Use Case: Server Fleet Inventory, Financial Ledgers, and Tabular Datasets.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 07 Data Tables"),
		nova.Size(950, 700),
		nova.Theme(theme.Dark()),
	)

	type ServerNode struct {
		ID       string
		Hostname string
		Region   string
		CPUUsage string
		Memory   string
		Status   string
	}

	servers := []ServerNode{
		{"SRV-101", "us-east-core-01", "us-east-1", "24.5%", "14.2 GB", "HEALTHY"},
		{"SRV-102", "us-east-core-02", "us-east-1", "88.1%", "28.9 GB", "HIGH_LOAD"},
		{"SRV-103", "eu-west-node-01", "eu-west-1", "12.0%", "8.1 GB", "HEALTHY"},
		{"SRV-104", "eu-west-node-02", "eu-west-1", "0.0%", "0.0 GB", "OFFLINE"},
		{"SRV-105", "ap-south-gateway-01", "ap-south-1", "45.8%", "18.6 GB", "HEALTHY"},
		{"SRV-106", "ap-south-gateway-02", "ap-south-1", "61.3%", "22.4 GB", "HEALTHY"},
		{"SRV-107", "sa-east-edge-01", "sa-east-1", "33.7%", "11.0 GB", "HEALTHY"},
	}

	cols := []data.TableColumn{
		{Title: "Node ID", Width: 110},
		{Title: "Hostname", Width: 200},
		{Title: "Region", Width: 130},
		{Title: "CPU Load", Width: 110},
		{Title: "RAM Usage", Width: 120},
		{Title: "Status", Width: 130},
	}

	table := data.Table(cols, len(servers), func(row, col int) string {
		s := servers[row]
		switch col {
		case 0:
			return s.ID
		case 1:
			return s.Hostname
		case 2:
			return s.Region
		case 3:
			return s.CPUUsage
		case 4:
			return s.Memory
		case 5:
			return s.Status
		default:
			return ""
		}
	})

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("07. Tabular Data & Virtualized Grid").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Data Grid").Info(),
					widgets.Badge(fmt.Sprintf("%d Nodes", len(servers))).Success(),
				).GapSpacing(10),
				ui.Text("High-performance structured table with column sizing, scroll support, and alternating row backgrounds.").Size(13).Col(color.Hex("#94A3B8")),

				// Table Card
				widgets.Card("Infrastructure Node Inventory",
					ui.Container().
						WithWidth(880).
						WithHeight(340).
						WithChild(table),
				),
			).GapSpacing(16),
		)
	}

	win.Content(buildUI())
	_ = win.SaveScreenshot("cookbook_07_tables.png")
	_ = app.Run()
}
