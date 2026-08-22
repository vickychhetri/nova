package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

func main() {
	app := nova.New()

	win := app.Window(
		nova.Title("Nova — Hello World"),
		nova.Size(600, 400),
	)

	win.Content(
		ui.Center(
			widgets.Card("Hello Nova!",
				ui.Column(
					ui.Text("Build once with Go. Render natively. Run everywhere."),
					ui.Row(
						widgets.Badge("Go-Native").Success(),
						widgets.Badge("GPU Accelerated").Info(),
						widgets.Badge("Declarative").Warning(),
					).GapSpacing(8),
					widgets.Button("Click Me").OnClick(func() {
						fmt.Println("Button clicked in Hello World!")
					}),
				).GapSpacing(12),
			),
		),
	)

	fmt.Println("Running Hello World example...")
	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Export frame snapshot
	_ = win.SaveScreenshot("hello_world.png")
	fmt.Println("Saved screenshot to hello_world.png")
}

var _ = geom.Pt
