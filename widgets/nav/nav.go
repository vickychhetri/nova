package nav

// Package nav provides tabs, sidebars, split panes, menus, toolbars, and status
// bars for application navigation and structure.

import (
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

	node.OnPointerDown = func(e *event.PointerEvent) {
		if e.Position.Y <= 44 {
			curX := 16.0
			for i, tab := range tc.Tabs {
				txtSz := text.MeasureText(tab.Title, 13, font.WeightMedium)
				tabW := txtSz.Width + 24.0
				if e.Position.X >= curX && e.Position.X <= curX+tabW {
					if tc.ActiveIndex != nil {
						tc.ActiveIndex.Set(i)
					}
					break
				}
				curX += tabW + 8.0
			}
		}
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
		Width:       230,
	}
}

func (s *SidebarComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := float64(len(s.Items))*42.0 + 64.0
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

	// Header section
	if s.HeaderTitle != "" {
		headerH := 48.0
		canvas.DrawText(s.HeaderTitle, geom.Pt(16, 16), 14, font.WeightBold, t.Palette.TextPrimary)
		canvas.DrawLine(geom.Pt(0, headerH), geom.Pt(w, headerH), t.Palette.Border, 1.0)
	}

	curY := 56.0
	for _, item := range s.Items {
		itemRect := geom.NewRect(8, curY, w-16, 36)
		textCol := t.Palette.TextSecondary
		fontWeight := font.WeightMedium

		if item.Selected {
			canvas.FillRoundedRect(itemRect, geom.RadiusUniform(6), t.Palette.Primary.WithAlpha(0.12))
			// Left indicator bar
			indicatorRect := geom.NewRect(8, curY+6, 3.5, 24)
			canvas.FillRoundedRect(indicatorRect, geom.RadiusUniform(2), t.Palette.Primary)
			textCol = t.Palette.Primary
			fontWeight = font.WeightBold
		}

		iconPrefix := item.Icon
		if iconPrefix != "" {
			iconPrefix += "  "
		}
		canvas.DrawText(iconPrefix+item.Title, geom.Pt(18, curY+10), 13, fontWeight, textCol)

		if item.Badge != "" {
			badgeSz := text.MeasureText(item.Badge, 11, font.WeightMedium)
			bRect := geom.NewRect(w-20-badgeSz.Width-10, curY+8, badgeSz.Width+10, 20)
			canvas.FillRoundedRect(bRect, geom.RadiusUniform(10), t.Palette.Secondary)
			canvas.DrawText(item.Badge, geom.Pt(bRect.X+5, curY+11), 11, font.WeightMedium, t.Palette.TextMuted)
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
	Direction  SplitDirection
	SplitRatio *state.Value[float64] // 0.0 to 1.0
	FixedFirst float64               // >0 to fix first pane width/height
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

func (sp *SplitPaneComponent) WithRatio(r float64) *SplitPaneComponent {
	if sp.SplitRatio != nil {
		sp.SplitRatio.Set(r)
	}
	return sp
}

func (sp *SplitPaneComponent) WithFixedFirst(w float64) *SplitPaneComponent {
	sp.FixedFirst = w
	return sp
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

	totalW := constraints.MaxWidth
	totalH := constraints.MaxHeight

	if sp.Direction == SplitHorizontal {
		splitW := totalW * 0.5
		if sp.FixedFirst > 0 {
			splitW = sp.FixedFirst
		} else if sp.SplitRatio != nil {
			splitW = totalW * sp.SplitRatio.Get()
		}
		if splitW > totalW-20 {
			splitW = totalW - 20
		}
		node.Children[0].Bounds = geom.NewRect(0, 0, splitW, totalH)
		node.Children[1].Bounds = geom.NewRect(splitW+1, 0, totalW-splitW-1, totalH)
		node.Children[0].Layout(layout.Tight(geom.Sz(splitW, totalH)))
		node.Children[1].Layout(layout.Tight(geom.Sz(totalW-splitW-1, totalH)))
	} else {
		splitH := totalH * 0.5
		if sp.FixedFirst > 0 {
			splitH = sp.FixedFirst
		} else if sp.SplitRatio != nil {
			splitH = totalH * sp.SplitRatio.Get()
		}
		if splitH > totalH-20 {
			splitH = totalH - 20
		}
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

// --- Enterprise MenuBar (Qt QMenuBar Style with Dropdowns & Events) ---

type MenuItem struct {
	Title    string
	Shortcut string
	Disabled bool
	Divider  bool
	OnClick  func()
	Children []MenuItem
}

// ActionItem creates a standard clickable menu item.
func ActionItem(title string, onClick func()) MenuItem {
	return MenuItem{Title: title, OnClick: onClick}
}

// ShortcutItem creates a clickable menu item with a keyboard shortcut hint.
func ShortcutItem(title, shortcut string, onClick func()) MenuItem {
	return MenuItem{Title: title, Shortcut: shortcut, OnClick: onClick}
}

// DividerItem creates a visual separator between menu item groups.
func DividerItem() MenuItem {
	return MenuItem{Divider: true}
}

type Menu struct {
	Title   string
	Items   []MenuItem
	OnClick func()
}

// NewMenu creates a dropdown menu containing menu items.
func NewMenu(title string, items ...MenuItem) Menu {
	return Menu{Title: title, Items: items}
}

// SimpleMenu creates a single-click menu button without a dropdown.
func SimpleMenu(title string, onClick func()) Menu {
	return Menu{Title: title, OnClick: onClick}
}

// MenuBarItem is maintained for backward compatibility.
type MenuBarItem struct {
	Title   string
	OnClick func()
}

type MenuBarComponent struct {
	ui.BaseComponent
	Menus      []Menu
	ActiveMenu *state.Value[int] // -1 = closed, 0..N = active dropdown menu
	HoverIndex *state.Value[int] // hover index for bar items
	HoverSub   *state.Value[int] // hover index for popup items
}

// MenuBar creates a native desktop top application menu bar with dropdown menus and event handlers.
func MenuBar(menus ...Menu) *MenuBarComponent {
	return &MenuBarComponent{
		Menus:      menus,
		ActiveMenu: state.Int(-1),
		HoverIndex: state.Int(-1),
		HoverSub:   state.Int(-1),
	}
}

// MenuBarFromItems creates a menu bar from simple MenuBarItem list (backward compatibility).
func MenuBarFromItems(items ...MenuBarItem) *MenuBarComponent {
	menus := make([]Menu, len(items))
	for i, it := range items {
		menus[i] = Menu{Title: it.Title, OnClick: it.OnClick}
	}
	return MenuBar(menus...)
}

func (mb *MenuBarComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	// Menu bar layout height is ALWAYS fixed at 28px (never pushes content down)
	h := 28.0

	node.OnPointerDown = func(e *event.PointerEvent) {
		activeIdx := -1
		if mb.ActiveMenu != nil {
			activeIdx = mb.ActiveMenu.Get()
		}

		if e.Position.Y <= 28.0 {
			curX := 10.0
			for i, m := range mb.Menus {
				txtW := text.MeasureText(m.Title, 12, font.WeightMedium).Width
				itemW := txtW + 20.0
				if e.Position.X >= curX && e.Position.X <= curX+itemW {
					if len(m.Items) > 0 {
						if activeIdx == i {
							mb.ActiveMenu.Set(-1)
						} else {
							mb.ActiveMenu.Set(i)
						}
					} else {
						if mb.ActiveMenu != nil {
							mb.ActiveMenu.Set(-1)
						}
						if m.OnClick != nil {
							m.OnClick()
						}
					}
					return
				}
				curX += itemW + 4.0
			}
			if mb.ActiveMenu != nil {
				mb.ActiveMenu.Set(-1)
			}
		}
	}

	node.OnPointerMove = func(e *event.PointerEvent) {
		activeIdx := -1
		if mb.ActiveMenu != nil {
			activeIdx = mb.ActiveMenu.Get()
		}

		if e.Position.Y <= 28.0 {
			curX := 10.0
			for i, m := range mb.Menus {
				txtW := text.MeasureText(m.Title, 12, font.WeightMedium).Width
				itemW := txtW + 20.0
				if e.Position.X >= curX && e.Position.X <= curX+itemW {
					if mb.HoverIndex != nil && mb.HoverIndex.Get() != i {
						mb.HoverIndex.Set(i)
					}
					if activeIdx >= 0 && activeIdx != i && len(m.Items) > 0 {
						mb.ActiveMenu.Set(i)
					}
					return
				}
				curX += itemW + 4.0
			}
			if mb.HoverIndex != nil && mb.HoverIndex.Get() != -1 {
				mb.HoverIndex.Set(-1)
			}
		}
	}

	node.OnPointerLeave = func() {
		if mb.HoverIndex != nil {
			mb.HoverIndex.Set(-1)
		}
		if mb.HoverSub != nil {
			mb.HoverSub.Set(-1)
		}
	}

	return constraints.Constrain(geom.Sz(constraints.MaxWidth, h))
}

func (mb *MenuBarComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width
	barH := 28.0

	// Menu bar background
	canvas.FillRect(geom.NewRect(0, 0, w, barH), t.Palette.Surface)
	canvas.DrawLine(geom.Pt(0, barH), geom.Pt(w, barH), t.Palette.Border, 1.0)

	activeIdx := -1
	if mb.ActiveMenu != nil {
		activeIdx = mb.ActiveMenu.Get()
	}
	hoverIdx := -1
	if mb.HoverIndex != nil {
		hoverIdx = mb.HoverIndex.Get()
	}

	// Render Top Menu Items
	curX := 10.0
	menuOffsets := make([]float64, len(mb.Menus))
	for i, m := range mb.Menus {
		menuOffsets[i] = curX
		txtSz := text.MeasureText(m.Title, 12, font.WeightMedium)
		itemW := txtSz.Width + 20.0

		if i == activeIdx {
			canvas.FillRoundedRect(geom.NewRect(curX, 3, itemW, barH-6), geom.RadiusUniform(4), t.Palette.Primary)
			canvas.DrawText(m.Title, geom.Pt(curX+10, 7), 12, font.WeightMedium, color.White)
		} else if i == hoverIdx {
			canvas.FillRoundedRect(geom.NewRect(curX, 3, itemW, barH-6), geom.RadiusUniform(4), t.Palette.SecondaryHover)
			canvas.DrawText(m.Title, geom.Pt(curX+10, 7), 12, font.WeightMedium, t.Palette.TextPrimary)
		} else {
			canvas.DrawText(m.Title, geom.Pt(curX+10, 7), 12, font.WeightMedium, t.Palette.TextPrimary)
		}

		curX += itemW + 4.0
	}

	// Configure Floating Overlay for active dropdown
	if activeIdx >= 0 && activeIdx < len(mb.Menus) && len(mb.Menus[activeIdx].Items) > 0 {
		m := mb.Menus[activeIdx]
		menuX := menuOffsets[activeIdx]
		popupW := 220.0
		popupH := float64(len(m.Items))*26.0 + 12.0

		// Register overlay painter (painted ON TOP of all other widgets)
		node.PaintOverlay = func(c *render.Canvas) {
			popupRect := geom.NewRect(menuX, barH+1, popupW, popupH)
			c.FillRoundedRect(popupRect, geom.RadiusUniform(6), t.Palette.Surface)
			c.StrokeRoundedRect(popupRect, geom.RadiusUniform(6), t.Palette.Border, 1.0)

			hoverSub := -1
			if mb.HoverSub != nil {
				hoverSub = mb.HoverSub.Get()
			}

			itemY := barH + 6.0
			for j, item := range m.Items {
				itemRect := geom.NewRect(menuX+4, itemY, popupW-8, 24)

				if item.Divider {
					c.DrawLine(geom.Pt(menuX+8, itemY+12), geom.Pt(menuX+popupW-8, itemY+12), t.Palette.Border, 1.0)
				} else {
					if j == hoverSub && !item.Disabled {
						c.FillRoundedRect(itemRect, geom.RadiusUniform(4), t.Palette.SecondaryHover)
					}

					textCol := t.Palette.TextPrimary
					if item.Disabled {
						textCol = t.Palette.TextMuted
					}

					c.DrawText(item.Title, geom.Pt(menuX+12, itemY+5), 12, font.WeightRegular, textCol)

					if item.Shortcut != "" {
						scSz := text.MeasureText(item.Shortcut, 11, font.WeightRegular)
						c.DrawText(item.Shortcut, geom.Pt(menuX+popupW-scSz.Width-12, itemY+6), 11, font.WeightRegular, t.Palette.TextSecondary)
					}
				}

				itemY += 26.0
			}
		}

		// Register overlay pointer down handler
		node.OnOverlayPointerDown = func(p geom.Point) bool {
			// If clicked inside dropdown popup
			if p.X >= menuX && p.X <= menuX+popupW && p.Y >= barH && p.Y <= barH+popupH {
				itemIdx := int((p.Y - (barH + 6.0)) / 26.0)
				if itemIdx >= 0 && itemIdx < len(m.Items) {
					item := m.Items[itemIdx]
					if !item.Disabled && !item.Divider && item.OnClick != nil {
						item.OnClick()
					}
				}
				mb.ActiveMenu.Set(-1)
				return true
			}

			// If clicked on top bar, let regular bar handler process it
			if p.Y <= barH {
				return false
			}

			// Clicked outside on window content -> dismiss menu without triggering content click
			mb.ActiveMenu.Set(-1)
			return true
		}

		// Register overlay pointer move handler
		node.OnOverlayPointerMove = func(p geom.Point) bool {
			if p.X >= menuX && p.X <= menuX+popupW && p.Y >= barH && p.Y <= barH+popupH {
				itemIdx := int((p.Y - (barH + 6.0)) / 26.0)
				if mb.HoverSub != nil && mb.HoverSub.Get() != itemIdx {
					mb.HoverSub.Set(itemIdx)
				}
				return true
			} else {
				if mb.HoverSub != nil && mb.HoverSub.Get() != -1 {
					mb.HoverSub.Set(-1)
				}
			}

			if p.Y <= barH {
				curX := 10.0
				for i, barMenu := range mb.Menus {
					txtW := text.MeasureText(barMenu.Title, 12, font.WeightMedium).Width
					itemW := txtW + 20.0
					if p.X >= curX && p.X <= curX+itemW {
						if mb.HoverIndex != nil && mb.HoverIndex.Get() != i {
							mb.HoverIndex.Set(i)
						}
						if activeIdx != i && len(barMenu.Items) > 0 {
							mb.ActiveMenu.Set(i)
						}
						return true
					}
					curX += itemW + 4.0
				}
			}

			return false
		}
	} else {
		node.PaintOverlay = nil
		node.OnOverlayPointerDown = nil
		node.OnOverlayPointerMove = nil
	}
}

// --- Enterprise Toolbar (Qt QToolBar Style) ---

type ToolbarComponent struct {
	ui.BaseComponent
	Children []ui.Component
}

// Toolbar creates a horizontal actions toolbar.
func Toolbar(children ...ui.Component) *ToolbarComponent {
	return &ToolbarComponent{Children: children}
}

func (tb *ToolbarComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := 46.0
	curX := 14.0

	if len(node.Children) == 0 {
		for _, child := range tb.Children {
			cn := ui.NewNode(child)
			cn.Parent = node
			node.Children = append(node.Children, cn)
		}
	}

	for i, cn := range node.Children {
		cn.Component = tb.Children[i]
		cSz := cn.Layout(layout.Loose(geom.Sz(constraints.MaxWidth-curX, h-12)))
		cn.Bounds = geom.NewRect(curX, 6, cSz.Width, cSz.Height)
		curX += cSz.Width + 10.0
	}

	return constraints.Constrain(geom.Sz(constraints.MaxWidth, h))
}

func (tb *ToolbarComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width
	h := 46.0

	canvas.FillRect(geom.NewRect(0, 0, w, h), t.Palette.Surface)
	canvas.DrawLine(geom.Pt(0, h), geom.Pt(w, h), t.Palette.Border, 1.0)

	for _, child := range node.Children {
		child.Paint(canvas)
	}
}

// --- Enterprise StatusBar (Qt QStatusBar Style) ---

type StatusSegment struct {
	Text  string
	Width float64
}

type StatusBarComponent struct {
	ui.BaseComponent
	MainMessage string
	Segments    []StatusSegment
}

// StatusBar creates a native desktop bottom status bar with panels.
func StatusBar(mainMsg string, segments ...StatusSegment) *StatusBarComponent {
	return &StatusBarComponent{
		MainMessage: mainMsg,
		Segments:    segments,
	}
}

func (sb *StatusBarComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(constraints.MaxWidth, 26))
}

func (sb *StatusBarComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width
	h := 26.0

	canvas.FillRect(geom.NewRect(0, 0, w, h), t.Palette.Surface)
	canvas.DrawLine(geom.Pt(0, 0), geom.Pt(w, 0), t.Palette.Border, 1.0)

	// Status indicator dot
	canvas.FillCircle(geom.Pt(16, 13), 3.5, t.Palette.Success)

	// Main message on left
	canvas.DrawText(sb.MainMessage, geom.Pt(26, 6), 11, font.WeightRegular, t.Palette.TextSecondary)

	// Right aligned status segments
	curRight := w - 10.0
	for i := len(sb.Segments) - 1; i >= 0; i-- {
		seg := sb.Segments[i]
		segW := seg.Width
		if segW <= 0 {
			segW = text.MeasureText(seg.Text, 11, font.WeightRegular).Width + 16.0
		}
		curRight -= segW

		// Vertical divider
		canvas.DrawLine(geom.Pt(curRight, 4), geom.Pt(curRight, h-4), t.Palette.Border, 1.0)
		canvas.DrawText(seg.Text, geom.Pt(curRight+8, 6), 11, font.WeightRegular, t.Palette.TextMuted)
	}
}

// Ensure color import used
var _ = color.White
