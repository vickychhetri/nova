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
	"github.com/vickychhetri/nova/widgets/editor"
)

// Use Case: IDE Code Viewer, Script Scratchpad, and Syntax Highlighted Text.
func main() {
	app := nova.New()
	win := app.Window(
		nova.Title("UI Cookbook - 11 Code Editor"),
		nova.Size(950, 700),
		nova.Theme(theme.Dark()),
	)

	sampleCode := `package main

import "fmt"

func CalculateMetrics(load float64) string {
    if load > 0.85 {
        return "HIGH_UTILIZATION"
    }
    return "NOMINAL"
}

func main() {
    status := CalculateMetrics(0.42)
    fmt.Printf("Cluster status: %s\n", status)
}`

	codeState := state.String(sampleCode)
	codeEditor := editor.CodeEditor(sampleCode, "go").
		Bind(codeState)

	var rebuild func()

	buildUI := func() ui.Component {
		return ui.Padding(geom.All(24),
			ui.Column(
				// Header
				ui.Row(
					ui.Text("11. Interactive Code Editor").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Go Syntax").Success(),
					widgets.Badge("Interactive").Info(),
				).GapSpacing(10),
				ui.Text("Code editor with line numbers, syntax highlighting (keywords, strings, functions), and cursor typing.").Size(13).Col(color.Hex("#94A3B8")),

				// Editor Card
				widgets.Card("Go Source Code Scratchpad",
					ui.Column(
						ui.Container().
							WithWidth(880).
							WithHeight(360).
							WithChild(codeEditor),
						ui.Row(
							widgets.Button("Reset Sample Code").Secondary().OnClick(func() {
								codeState.Set(sampleCode)
								rebuild()
							}),
							ui.Spacer(),
							widgets.Badge(fmt.Sprintf("Character Count: %d", len(codeState.Get()))).Info(),
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
	_ = win.SaveScreenshot("cookbook_11_editor.png")
	_ = app.Run()
}
