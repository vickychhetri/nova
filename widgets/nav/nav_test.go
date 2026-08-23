package nav_test

import (
	"testing"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets/nav"
)

func TestMenuBarOverlayBehavior(t *testing.T) {
	fileExportClicked := false

	menuBar := nav.MenuBar(
		nav.NewMenu("File",
			nav.ActionItem("New Document", nil),
			nav.ActionItem("Export PDF", func() { fileExportClicked = true }),
		),
		nav.NewMenu("Edit",
			nav.ActionItem("Undo", nil),
		),
	)

	node := ui.NewNode(menuBar)
	node.Mount(ui.BuildContext{
		Theme: theme.Light(),
		Scale: 1.0,
	})

	// 1. Initial layout height must be exactly 28px
	sz := node.Layout(layout.TightWidth(800))
	if sz.Height != 28.0 {
		t.Fatalf("expected MenuBar height to be 28.0, got %.2f", sz.Height)
	}

	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)
	node.Paint(canvas)

	// Initially no overlay is active
	if node.PaintOverlay != nil {
		t.Fatal("expected PaintOverlay to be nil when no menu is open")
	}

	// 2. Click "File" on top bar (e.g. x=20, y=10)
	node.OnPointerDown(&event.PointerEvent{
		Position: geom.Pt(20, 10),
	})

	// Layout height must STILL be 28.0px (never push content down!)
	szAfterOpen := node.Layout(layout.TightWidth(800))
	if szAfterOpen.Height != 28.0 {
		t.Fatalf("expected MenuBar height to remain 28.0 when menu is open, got %.2f", szAfterOpen.Height)
	}

	// Paint pass registers floating overlay
	buf.Clear()
	node.Paint(canvas)
	if node.PaintOverlay == nil {
		t.Fatal("expected PaintOverlay to be registered after opening menu")
	}
	if node.OnOverlayPointerDown == nil {
		t.Fatal("expected OnOverlayPointerDown to be registered after opening menu")
	}

	// 3. Click "Export PDF" inside floating dropdown popup (y=65, inside popup)
	handled := node.OnOverlayPointerDown(geom.Pt(30, 65))
	if !handled {
		t.Fatal("expected OnOverlayPointerDown to handle click inside popup")
	}
	if !fileExportClicked {
		t.Fatal("expected Export PDF action to be executed")
	}

	// Menu should now be closed
	node.Paint(canvas)
	if node.PaintOverlay != nil {
		t.Fatal("expected PaintOverlay to be nil after menu item click")
	}
}
