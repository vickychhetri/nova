package layout_test

import (
	"testing"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/layout"
)

func TestFlexRowLayout(t *testing.T) {
	constraints := layout.Tight(geom.Sz(300, 100))

	children := []layout.FlexChildInput{
		{
			Measure: func(c layout.BoxConstraints) geom.Size {
				return geom.Sz(50, 40)
			},
			Flex: 0,
		},
		{
			Measure: func(c layout.BoxConstraints) geom.Size {
				return geom.Sz(c.MinWidth, 40)
			},
			Flex: 1, // should take remaining width: 300 - 50 - 50 - 20 (gaps) = 180
		},
		{
			Measure: func(c layout.BoxConstraints) geom.Size {
				return geom.Sz(50, 40)
			},
			Flex: 0,
		},
	}

	result := layout.ComputeFlex(constraints, layout.FlexConfig{
		Direction: layout.AxisHorizontal,
		MainAxis:  layout.MainStart,
		CrossAxis: layout.CrossCenter,
		Gap:       10,
	}, children)

	if result.Size.Width != 300 || result.Size.Height != 100 {
		t.Fatalf("expected container size 300x100, got %s", result.Size)
	}

	if len(result.ChildBounds) != 3 {
		t.Fatalf("expected 3 child bounds, got %d", len(result.ChildBounds))
	}

	// Child 0: x=0, y=(100-40)/2=30, w=50, h=40
	b0 := result.ChildBounds[0]
	if b0.X != 0 || b0.Y != 30 || b0.Width != 50 || b0.Height != 40 {
		t.Fatalf("unexpected child 0 bounds: %s", b0)
	}

	// Child 1: x=50+10=60, y=30, w=180, h=40
	b1 := result.ChildBounds[1]
	if b1.X != 60 || b1.Y != 30 || b1.Width != 180 || b1.Height != 40 {
		t.Fatalf("unexpected child 1 bounds: %s", b1)
	}

	// Child 2: x=60+180+10=250, y=30, w=50, h=40
	b2 := result.ChildBounds[2]
	if b2.X != 250 || b2.Y != 30 || b2.Width != 50 || b2.Height != 40 {
		t.Fatalf("unexpected child 2 bounds: %s", b2)
	}
}

func TestStackLayout(t *testing.T) {
	constraints := layout.Loose(geom.Sz(400, 300))

	topRightPos := layout.StackPosition{
		IsPositioned: true,
		Top:          ptr(10.0),
		Right:        ptr(15.0),
		Width:        ptr(60.0),
		Height:       ptr(40.0),
	}

	children := []layout.StackChildInput{
		{
			Measure: func(c layout.BoxConstraints) geom.Size {
				return geom.Sz(200, 150)
			},
			Position: layout.StackPosition{IsPositioned: false},
		},
		{
			Measure: func(c layout.BoxConstraints) geom.Size {
				return geom.Sz(60, 40)
			},
			Position: topRightPos,
		},
	}

	res := layout.ComputeStack(constraints, layout.AlignCenter, children)

	if res.Size.Width != 200 || res.Size.Height != 150 {
		t.Fatalf("expected container size 200x150, got %s", res.Size)
	}

	// Centered child
	b0 := res.ChildBounds[0]
	if b0.X != 0 || b0.Y != 0 || b0.Width != 200 || b0.Height != 150 {
		t.Fatalf("unexpected child 0 bounds: %s", b0)
	}

	// Positioned child: top=10, right=15 => x = 200 - 15 - 60 = 125, y = 10
	b1 := res.ChildBounds[1]
	if b1.X != 125 || b1.Y != 10 || b1.Width != 60 || b1.Height != 40 {
		t.Fatalf("unexpected child 1 bounds: %s", b1)
	}
}

func TestGridLayout(t *testing.T) {
	constraints := layout.TightWidth(300)

	children := []layout.GridChildInput{
		{Measure: func(c layout.BoxConstraints) geom.Size { return geom.Sz(c.MaxWidth, 50) }},
		{Measure: func(c layout.BoxConstraints) geom.Size { return geom.Sz(c.MaxWidth, 50) }},
		{Measure: func(c layout.BoxConstraints) geom.Size { return geom.Sz(c.MaxWidth, 50) }},
	}

	res := layout.ComputeGrid(constraints, layout.GridConfig{
		Columns:    2,
		ColumnGap:  20,
		RowGap:     10,
		ItemHeight: 50,
	}, children)

	// Col width = (300 - 20) / 2 = 140
	// 2 rows => height = 50 + 10 + 50 = 110
	if res.Size.Width != 300 || res.Size.Height != 110 {
		t.Fatalf("expected grid size 300x110, got %s", res.Size)
	}

	b0 := res.ChildBounds[0]
	if b0.X != 0 || b0.Y != 0 || b0.Width != 140 || b0.Height != 50 {
		t.Fatalf("unexpected child 0 bounds: %s", b0)
	}

	b1 := res.ChildBounds[1]
	if b1.X != 160 || b1.Y != 0 || b1.Width != 140 || b1.Height != 50 {
		t.Fatalf("unexpected child 1 bounds: %s", b1)
	}

	b2 := res.ChildBounds[2]
	if b2.X != 0 || b2.Y != 60 || b2.Width != 140 || b2.Height != 50 {
		t.Fatalf("unexpected child 2 bounds: %s", b2)
	}
}

func ptr(v float64) *float64 {
	return &v
}
