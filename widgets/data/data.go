package data

import (
	"fmt"
	"math"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/text"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/virtualization"
)

// --- VirtualList ---

type VirtualListComponent struct {
	ui.BaseComponent
	ItemCount    int
	ItemHeight   float64
	ScrollOffset *state.Value[float64]
	RenderItem   func(index int) ui.Component
	virtualizer  *virtualization.Virtualizer
}

// VirtualList creates a viewport-virtualized list capable of handling 1,000,000 items at 60+ FPS.
func VirtualList(itemCount int, itemHeight float64, renderItem func(index int) ui.Component) *VirtualListComponent {
	return &VirtualListComponent{
		ItemCount:    itemCount,
		ItemHeight:   itemHeight,
		ScrollOffset: state.Float(0),
		RenderItem:   renderItem,
		virtualizer:  virtualization.NewVirtualizer(itemCount, itemHeight),
	}
}

func (vl *VirtualListComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := constraints.MaxHeight
	if h >= layout.Infinity {
		h = 400
	}
	w := constraints.MaxWidth
	if w >= layout.Infinity {
		w = 300
	}

	node.OnScroll = func(e *event.ScrollEvent) {
		if vl.ScrollOffset != nil {
			maxScroll := math.Max(0, vl.virtualizer.TotalContentHeight()-h)
			cur := vl.ScrollOffset.Get()
			newScroll := math.Max(0, math.Min(maxScroll, cur+e.DeltaY*20))
			vl.ScrollOffset.Set(newScroll)
		}
	}

	return constraints.Constrain(geom.Sz(w, h))
}

func (vl *VirtualListComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	w := node.Bounds.Width
	h := node.Bounds.Height
	scroll := 0.0
	if vl.ScrollOffset != nil {
		scroll = vl.ScrollOffset.Get()
	}

	vis := vl.virtualizer.ComputeVisibleRange(scroll, h)

	canvas.PushClip(geom.NewRect(0, 0, w, h))

	for i := vis.StartIndex; i <= vis.EndIndex; i++ {
		itemComp := vl.RenderItem(i)
		itemY := float64(i)*vl.ItemHeight - scroll

		canvas.Save()
		canvas.Translate(0, itemY)
		itemNode := ui.NewNode(itemComp)
		itemNode.Bounds = geom.NewRect(0, 0, w, vl.ItemHeight)
		itemNode.Paint(canvas)
		canvas.Restore()
	}

	// Scrollbar
	totalH := vl.virtualizer.TotalContentHeight()
	if totalH > h {
		barH := math.Max(24, (h/totalH)*h)
		barY := (scroll / (totalH - h)) * (h - barH)
		barRect := geom.NewRect(w-6, barY, 4, barH)
		canvas.FillRoundedRect(barRect, geom.RadiusUniform(2), theme.Current().Palette.BorderHover)
	}

	canvas.PopClip()
}

// --- Virtualized Table ---

type TableColumn struct {
	Title string
	Width float64
	Field string
}

type TableComponent struct {
	ui.BaseComponent
	Columns      []TableColumn
	RowCount     int
	RowHeight    float64
	ScrollOffset *state.Value[float64]
	SelectedIndex *state.Value[int]
	GetCell      func(row int, col int) string
	virtualizer  *virtualization.Virtualizer
}

// Table creates a virtualized, sorting-ready tabular data view.
func Table(columns []TableColumn, rowCount int, getCell func(row int, col int) string) *TableComponent {
	return &TableComponent{
		Columns:       columns,
		RowCount:      rowCount,
		RowHeight:     32.0,
		ScrollOffset:  state.Float(0),
		SelectedIndex: state.Int(-1),
		GetCell:       getCell,
		virtualizer:   virtualization.NewVirtualizer(rowCount, 32.0),
	}
}

func (tc *TableComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := constraints.MaxHeight
	if h >= layout.Infinity {
		h = 400
	}
	w := constraints.MaxWidth
	if w >= layout.Infinity {
		totalColW := 0.0
		for _, c := range tc.Columns {
			totalColW += c.Width
		}
		w = totalColW
	}

	node.OnScroll = func(e *event.ScrollEvent) {
		if tc.ScrollOffset != nil {
			bodyH := h - 36.0 // 36px header
			maxScroll := math.Max(0, tc.virtualizer.TotalContentHeight()-bodyH)
			cur := tc.ScrollOffset.Get()
			newScroll := math.Max(0, math.Min(maxScroll, cur+e.DeltaY*20))
			tc.ScrollOffset.Set(newScroll)
		}
	}

	return constraints.Constrain(geom.Sz(w, h))
}

func (tc *TableComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width
	h := node.Bounds.Height
	scroll := 0.0
	if tc.ScrollOffset != nil {
		scroll = tc.ScrollOffset.Get()
	}

	headerH := 32.0
	bodyH := h - headerH

	// Table border box
	tableRect := geom.NewRect(0, 0, w, h)
	canvas.FillRoundedRect(tableRect, t.Radii.SM, t.Palette.Surface)
	canvas.StrokeRoundedRect(tableRect, t.Radii.SM, t.Palette.Border, 1.0)

	// 1. Header Bar (Qt QHeaderView Style)
	headerRect := geom.NewRect(0, 0, w, headerH)
	canvas.FillRoundedRect(headerRect, geom.RadiusSeparate(4, 4, 0, 0), t.Palette.Secondary)
	canvas.DrawLine(geom.Pt(0, headerH), geom.Pt(w, headerH), t.Palette.Border, 1.0)

	curX := 0.0
	for _, col := range tc.Columns {
		// Header text
		canvas.DrawText(col.Title, geom.Pt(curX+10, 8), 12, font.WeightBold, t.Palette.TextPrimary)
		curX += col.Width
		// Header vertical separator line
		canvas.DrawLine(geom.Pt(curX, 0), geom.Pt(curX, headerH), t.Palette.Border, 1.0)
	}

	// 2. Virtualized Body Rows
	canvas.Save()
	canvas.Translate(0, headerH)
	canvas.PushClip(geom.NewRect(0, 0, w, bodyH))

	vis := tc.virtualizer.ComputeVisibleRange(scroll, bodyH)
	selectedRow := -1
	if tc.SelectedIndex != nil {
		selectedRow = tc.SelectedIndex.Get()
	}

	for r := vis.StartIndex; r <= vis.EndIndex; r++ {
		rowY := float64(r)*tc.RowHeight - scroll
		rowRect := geom.NewRect(0, rowY, w, tc.RowHeight)

		// Alternating Row background & selection
		if r == selectedRow {
			canvas.FillRect(rowRect, t.Palette.Primary)
		} else if r%2 == 1 {
			canvas.FillRect(rowRect, t.Palette.Background.WithAlpha(0.4))
		}

		// Horizontal row grid line
		canvas.DrawLine(geom.Pt(0, rowY+tc.RowHeight), geom.Pt(w, rowY+tc.RowHeight), t.Palette.Border.WithAlpha(0.35), 0.5)

		// Cells & Vertical column grid lines
		cellX := 0.0
		for c, col := range tc.Columns {
			val := tc.GetCell(r, c)
			val = text.TruncateWithEllipsis(val, col.Width-16, 12, font.WeightRegular)
			textCol := t.Palette.TextPrimary
			if r == selectedRow {
				textCol = color.White
			}
			canvas.DrawText(val, geom.Pt(cellX+10, rowY+7), 12, font.WeightRegular, textCol)
			cellX += col.Width

			// Vertical cell grid line
			canvas.DrawLine(geom.Pt(cellX, rowY), geom.Pt(cellX, rowY+tc.RowHeight), t.Palette.Border.WithAlpha(0.25), 0.5)
		}
	}

	// Body scrollbar
	totalH := tc.virtualizer.TotalContentHeight()
	if totalH > bodyH {
		barH := math.Max(24, (bodyH/totalH)*bodyH)
		barY := (scroll / (totalH - bodyH)) * (bodyH - barH)
		barRect := geom.NewRect(w-6, barY, 4, barH)
		canvas.FillRoundedRect(barRect, geom.RadiusUniform(2), t.Palette.BorderHover)
	}

	canvas.PopClip()
	canvas.Restore()
}

// --- Tree View ---

type TreeNode struct {
	Label    string
	Children []TreeNode
	Expanded bool
}

type TreeComponent struct {
	ui.BaseComponent
	RootNodes []TreeNode
}

// Tree creates an expandable hierarchical tree view.
func Tree(roots ...TreeNode) *TreeComponent {
	return &TreeComponent{RootNodes: roots}
}

func (tc *TreeComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	totalRows := countTreeNodes(tc.RootNodes)
	return constraints.Constrain(geom.Sz(240, float64(totalRows)*28.0))
}

func (tc *TreeComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	curY := 4.0
	var drawNode func(n TreeNode, depth int)
	drawNode = func(n TreeNode, depth int) {
		indent := float64(depth) * 16.0
		prefix := "• "
		if len(n.Children) > 0 {
			if n.Expanded {
				prefix = "▼ "
			} else {
				prefix = "▶ "
			}
		}
		canvas.DrawText(prefix+n.Label, geom.Pt(indent+10, curY+4), 13, font.WeightRegular, t.Palette.TextPrimary)
		curY += 28.0

		if n.Expanded {
			for _, child := range n.Children {
				drawNode(child, depth+1)
			}
		}
	}

	for _, root := range tc.RootNodes {
		drawNode(root, 0)
	}
}

func countTreeNodes(nodes []TreeNode) int {
	total := 0
	for _, n := range nodes {
		total++
		if n.Expanded {
			total += countTreeNodes(n.Children)
		}
	}
	return total
}

var _ = fmt.Sprintf
