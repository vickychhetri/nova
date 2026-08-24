package ui_test

import (
	"testing"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
)

// TestUIComponentTreeAndHitTest verifies the main public UI pipeline from
// component construction through mounting, layout, painting, hit testing, and
// click-handler execution.
func TestUIComponentTreeAndHitTest(t *testing.T) {
	// The callback changes this flag so the test can prove that hit testing
	// reached the button and that its handler was invoked.
	clicked := false

	// Build a nested tree with a text header and a row containing one button.
	// Gap values make the tree exercise real layout offsets rather than a flat
	// single-node case.
	tree := ui.Column(
		ui.Text("Welcome to Nova"),
		ui.Row(
			ui.Button("Click Me").OnClick(func() {
				clicked = true
			}),
		).GapSpacing(8),
	).GapSpacing(12)

	node := ui.NewNode(tree)
	// Mount reconciles the declarative components into retained Nodes using the
	// supplied theme and logical-to-device scale context.
	node.Mount(ui.BuildContext{
		Theme: theme.Dark(),
		Scale: 1.0,
	})

	// Tight constraints require the root to occupy exactly 400 by 300.
	sz := node.Layout(layout.Tight(geom.Sz(400, 300)))
	if sz.Width != 400 || sz.Height != 300 {
		t.Fatalf("expected layout size 400x300, got %s", sz)
	}

	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)
	// Paint records commands instead of drawing directly to a native window.
	node.Paint(canvas)

	if buf.Len() == 0 {
		t.Fatal("expected paint commands recorded into buffer")
	}

	// Locate the retained row and button nodes so the expected center can be
	// assembled from each node's parent-relative bounds.
	rowNode := node.Children[1]
	btnNode := rowNode.Children[0]

	// Convert the button center from nested local coordinates into the root
	// coordinate space by adding each ancestor's origin.
	btnCenter := geom.Pt(
		node.Bounds.X+rowNode.Bounds.X+btnNode.Bounds.X+btnNode.Bounds.Width/2,
		node.Bounds.Y+rowNode.Bounds.Y+btnNode.Bounds.Y+btnNode.Bounds.Height/2,
	)

	// HitTest should select the deepest interactive node under this point.
	hit := node.HitTest(btnCenter)
	if hit == nil {
		t.Fatal("expected hit test to find button node")
	}

	// Invoke the discovered callback to verify that the hit target is the
	// configured button rather than merely a containing layout node.
	if hit.OnClick != nil {
		hit.OnClick()
	}

	if !clicked {
		t.Fatal("expected button click handler to execute")
	}
}
