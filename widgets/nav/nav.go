package nav

import (
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/text"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
)

// --- Tabs ---

type TabItem struct {
	Title   string
	Content ui.Component
}

type TabsComponent struct {
	ui.BaseComponent
	Tabs        []TabItem
	ActiveIndex *state.Value[int]
}

// Tabs creates a tabbed view container.
func Tabs(tabs ...TabItem) *TabsComponent {
	return &TabsComponent{
		Tabs:        tabs,
		ActiveIndex: state.Int(0),
	}
}

func (tc *TabsComponent) Bind(idx *state.Value[int]) *TabsComponent {
	tc.ActiveIndex = idx
	return tc
}

func (tc *TabsComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	activeIdx := 0
	if tc.ActiveIndex != nil {
		activeIdx = tc.ActiveIndex.Get()
	}
	if activeIdx < 0 || activeIdx >= len(tc.Tabs) {
		activeIdx = 0
	}

	var activeContent ui.Component
	if len(tc.Tabs) > 0 {
		activeContent = tc.Tabs[activeIdx].Content
	}

	// Reconcile child node for active tab body
	if activeContent != nil {
		if len(node.Children) == 0 {
			childNode := ui.NewNode(activeContent)
			childNode.Parent = node
			node.Children = []*ui.Node{childNode}
		} else {
			node.Children[0].Component = activeContent
		}

		bodyConstraints := constraints.Deflate(geom.TRBL(44, 0, 0, 0))
		bodySz := node.Children[0].Layout(bodyConstraints)
		node.Children[0].Bounds = geom.NewRect(0, 44, bodySz.Width, bodySz.Height)

		totalH := bodySz.Height + 44.0
		totalW := bodySz.Width
		return constraints.Constrain(geom.Sz(totalW, totalH))
	}

	return constraints.Constrain(geom.Sz(0, 44))
}

func (tc *TabsComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	activeIdx := 0
	if tc.ActiveIndex != nil {
		activeIdx = tc.ActiveIndex.Get()
	}

	// Tab header background
	headerRect := geom.NewRect(0, 0, node.Bounds.Width, 44)
	canvas.FillRect(headerRect, t.Palette.Surface)
	canvas.DrawLine(geom.Pt(0, 44), geom.Pt(node.Bounds.Width, 44), t.Palette.Border, 1.0)

	curX := 16.0
	for i, tab := range tc.Tabs {
		txtSz := text.MeasureText(tab.Title, 13, font.WeightMedium)
		tabW := txtSz.Width + 24.0

		isActive := i == activeIdx
		textCol := t.Palette.TextSecondary
		if isActive {
			textCol = t.Palette.Primary
			// Active underline indicator
			canvas.DrawLine(geom.Pt(curX, 42), geom.Pt(curX+tabW, 42), t.Palette.Primary, 2.5)
		}

		tx := curX + (tabW-txtSz.Width)/2.0
		canvas.DrawText(tab.Title, geom.Pt(tx, 14), 13, font.WeightMedium, textCol)

		curX += tabW + 8.0
	}

	// Paint active content
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

// --- Sidebar Navigation ---

type SidebarItem struct {
	Icon     string
	Title    string
	Badge    string
	Selected bool
	OnClick  func()
}

type SidebarComponent struct {
	ui.BaseComponent
	Items       []SidebarItem
	Width       float64
	HeaderTitle string
}

// Sidebar creates a left-rail navigation sidebar.
func Sidebar(headerTitle string, items ...SidebarItem) *SidebarComponent {
	return &SidebarComponent{
		HeaderTitle: headerTitle,
		Items:       items,
		Width:       220,
	}
}

func (s *SidebarComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := float64(len(s.Items))*40.0 + 60.0
	if constraints.HasBoundedHeight() {
		h = constraints.MaxHeight
	}
	return constraints.Constrain(geom.Sz(s.Width, h))
}

func (s *SidebarComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width
	h := node.Bounds.Height

	canvas.FillRect(geom.NewRect(0, 0, w, h), t.Palette.Surface)
	canvas.DrawLine(geom.Pt(w, 0), geom.Pt(w, h), t.Palette.Border, 1.0)

	// Header
	if s.HeaderTitle != "" {
		canvas.DrawText(s.HeaderTitle, geom.Pt(16, 18), 16, font.WeightBold, t.Palette.TextPrimary)
	}

	curY := 52.0
	for _, item := range s.Items {
		itemRect := geom.NewRect(8, curY, w-16, 36)
		textCol := t.Palette.TextSecondary
		if item.Selected {
			canvas.FillRoundedRect(itemRect, t.Radii.MD, t.Palette.Primary.WithAlpha(0.15))
			textCol = t.Palette.Primary
		}

		iconPrefix := item.Icon
		if iconPrefix != "" {
			iconPrefix += " "
		}
		canvas.DrawText(iconPrefix+item.Title, geom.Pt(18, curY+10), 13, font.WeightMedium, textCol)

		if item.Badge != "" {
			badgeSz := text.MeasureText(item.Badge, 11, font.WeightRegular)
			canvas.DrawText(item.Badge, geom.Pt(w-24-badgeSz.Width, curY+11), 11, font.WeightRegular, t.Palette.TextMuted)
		}

		curY += 40.0
	}
}

// --- SplitPane ---

type SplitDirection int

const (
	SplitHorizontal SplitDirection = iota
	SplitVertical
)

type SplitPaneComponent struct {
	ui.BaseComponent
	Direction SplitDirection
	SplitRatio *state.Value[float64] // 0.0 to 1.0
	First      ui.Component
	Second     ui.Component
}

// SplitPane creates a resizable two-pane view.
func SplitPane(dir SplitDirection, first, second ui.Component) *SplitPaneComponent {
	return &SplitPaneComponent{
		Direction:  dir,
		SplitRatio: state.Float(0.5),
		First:      first,
		Second:     second,
	}
}

func (sp *SplitPaneComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	if len(node.Children) != 2 {
		node.Children = []*ui.Node{ui.NewNode(sp.First), ui.NewNode(sp.Second)}
		node.Children[0].Parent = node
		node.Children[1].Parent = node
	} else {
		node.Children[0].Component = sp.First
		node.Children[1].Component = sp.Second
	}

	ratio := 0.5
	if sp.SplitRatio != nil {
		ratio = sp.SplitRatio.Get()
	}

	totalW := constraints.MaxWidth
	totalH := constraints.MaxHeight

	if sp.Direction == SplitHorizontal {
		splitW := totalW * ratio
		node.Children[0].Bounds = geom.NewRect(0, 0, splitW, totalH)
		node.Children[1].Bounds = geom.NewRect(splitW+1, 0, totalW-splitW-1, totalH)
		node.Children[0].Layout(layout.Tight(geom.Sz(splitW, totalH)))
		node.Children[1].Layout(layout.Tight(geom.Sz(totalW-splitW-1, totalH)))
	} else {
		splitH := totalH * ratio
		node.Children[0].Bounds = geom.NewRect(0, 0, totalW, splitH)
		node.Children[1].Bounds = geom.NewRect(0, splitH+1, totalW, totalH-splitH-1)
		node.Children[0].Layout(layout.Tight(geom.Sz(totalW, splitH)))
		node.Children[1].Layout(layout.Tight(geom.Sz(totalW, totalH-splitH-1)))
	}

	return constraints.Constrain(geom.Sz(totalW, totalH))
}

func (sp *SplitPaneComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	if len(node.Children) == 2 {
		node.Children[0].Paint(canvas)
		node.Children[1].Paint(canvas)

		// Divider line
		if sp.Direction == SplitHorizontal {
			splitX := node.Children[1].Bounds.X - 1
			canvas.DrawLine(geom.Pt(splitX, 0), geom.Pt(splitX, node.Bounds.Height), t.Palette.Border, 1.0)
		} else {
			splitY := node.Children[1].Bounds.Y - 1
			canvas.DrawLine(geom.Pt(0, splitY), geom.Pt(node.Bounds.Width, splitY), t.Palette.Border, 1.0)
		}
	}
}

// --- Breadcrumb ---

type BreadcrumbItem struct {
	Label   string
	OnClick func()
}

type BreadcrumbComponent struct {
	ui.BaseComponent
	Items []BreadcrumbItem
}

// Breadcrumb creates a path navigation trail.
func Breadcrumb(items ...BreadcrumbItem) *BreadcrumbComponent {
	return &BreadcrumbComponent{Items: items}
}

func (b *BreadcrumbComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	totalW := 0.0
	for _, item := range b.Items {
		sz := text.MeasureText(item.Label+" / ", 13, font.WeightRegular)
		totalW += sz.Width
	}
	return constraints.Constrain(geom.Sz(totalW, 24))
}

func (b *BreadcrumbComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	curX := 0.0
	for i, item := range b.Items {
		isLast := i == len(b.Items)-1
		col := t.Palette.TextMuted
		if isLast {
			col = t.Palette.TextPrimary
		}

		canvas.DrawText(item.Label, geom.Pt(curX, 4), 13, font.WeightMedium, col)
		curX += text.MeasureText(item.Label, 13, font.WeightMedium).Width

		if !isLast {
			canvas.DrawText(" / ", geom.Pt(curX, 4), 13, font.WeightRegular, t.Palette.TextDisabled)
			curX += text.MeasureText(" / ", 13, font.WeightRegular).Width
		}
	}
}

// Ensure color import used
var _ = color.White
