package ui_test

import (
	"testing"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
)

func TestUIComponentTreeAndHitTest(t *testing.T) {
	clicked := false

	tree := ui.Column(
		ui.Text("Welcome to Nova"),
		ui.Row(
			ui.Button("Click Me").OnClick(func() {
				clicked = true
			}),
		).GapSpacing(8),
	).GapSpacing(12)

	node := ui.NewNode(tree)
	node.Mount(ui.BuildContext{
		Theme: theme.Dark(),
		Scale: 1.0,
	})

	sz := node.Layout(layout.Tight(geom.Sz(400, 300)))
	if sz.Width != 400 || sz.Height != 300 {
		t.Fatalf("expected layout size 400x300, got %s", sz)
	}

	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)
	node.Paint(canvas)

	if buf.Len() == 0 {
		t.Fatal("expected paint commands recorded into buffer")
	}

	// Test hit testing on the button
	// Find button's absolute location
	rowNode := node.Children[1]
	btnNode := rowNode.Children[0]

	btnCenter := geom.Pt(
		node.Bounds.X+rowNode.Bounds.X+btnNode.Bounds.X+btnNode.Bounds.Width/2,
		node.Bounds.Y+rowNode.Bounds.Y+btnNode.Bounds.Y+btnNode.Bounds.Height/2,
	)

	hit := node.HitTest(btnCenter)
	if hit == nil {
		t.Fatal("expected hit test to find button node")
	}

	if hit.OnClick != nil {
		hit.OnClick()
	}

	if !clicked {
		t.Fatal("expected button click handler to execute")
	}
}
