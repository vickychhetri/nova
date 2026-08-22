package nova_test

import (
	"testing"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/ui"
)

func TestNovaAppLifecycle(t *testing.T) {
	app := nova.New()

	win := app.Window(
		nova.Title("Test Nova App"),
		nova.Size(800, 600),
	)

	win.Content(
		ui.Column(
			ui.Text("Hello Nova!"),
			ui.Button("Click Me"),
		),
	)

	err := app.Run()
	if err != nil {
		t.Fatalf("app.Run() failed: %v", err)
	}

	if win.Title() != "Test Nova App" {
		t.Fatalf("unexpected title: %s", win.Title())
	}
	if win.Size().Width != 800 || win.Size().Height != 600 {
		t.Fatalf("unexpected window size: %s", win.Size())
	}
}
