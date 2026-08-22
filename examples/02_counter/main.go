package main

import (
	"fmt"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

func main() {
	app := nova.New()

	win := app.Window(
		nova.Title("Nova — Reactive Counter"),
		nova.Size(600, 450),
	)

	count := state.Int(0)
	doubleCount := state.Compute(func() string {
		return fmt.Sprintf("Doubled: %d", count.Get()*2)
	})

	win.Content(
		ui.Center(
			widgets.Card("Reactive Signal Counter",
				ui.Column(
					ui.Padding(geom.Symmetric(12, 0),
						ui.Column(
							ui.Text(state.Compute(func() string {
								return fmt.Sprintf("Current Count: %d", count.Get())
							})).Size(20).Weight(700),
							ui.Text(doubleCount).Size(14),
						).GapSpacing(4),
					),
					ui.Row(
						widgets.Button("- Decrement").Secondary().OnClick(func() {
							count.Update(func(c int) int { return c - 1 })
						}),
						widgets.Button("Reset").Danger().OnClick(func() {
							count.Set(0)
						}),
						widgets.Button("+ Increment").OnClick(func() {
							count.Update(func(c int) int { return c + 1 })
						}),
					).GapSpacing(8),
				).GapSpacing(12),
			),
		),
	)

	fmt.Println("Running Counter example...")
	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	_ = win.SaveScreenshot("counter.png")
	fmt.Println("Saved screenshot to counter.png")
}
