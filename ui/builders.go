package ui

import (
	"fmt"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/text"
	"github.com/vickychhetri/nova/theme"
)

// --- Flex Components (Column & Row) ---

// FlexComponent implements a flexbox container (Column or Row).
type FlexComponent struct {
	BaseComponent
	Direction AxisDirection
	MainAxis  layout.MainAxisAlignment
	CrossAxis layout.CrossAxisAlignment
	Gap       float64
	Children  []Component
}

type AxisDirection int

const (
	DirHorizontal AxisDirection = iota
	DirVertical
)

// Column creates a vertical flex column layout.
func Column(children ...Component) *FlexComponent {
	return &FlexComponent{
		Direction: DirVertical,
		MainAxis:  layout.MainStart,
		CrossAxis: layout.CrossStart,
		Children:  children,
	}
}

// Row creates a horizontal flex row layout.
func Row(children ...Component) *FlexComponent {
	return &FlexComponent{
		Direction: DirHorizontal,
		MainAxis:  layout.MainStart,
		CrossAxis: layout.CrossCenter,
		Children:  children,
	}
}

func (f *FlexComponent) GapSpacing(gap float64) *FlexComponent {
	f.Gap = gap
	return f
}

func (f *FlexComponent) AlignMain(align layout.MainAxisAlignment) *FlexComponent {
	f.MainAxis = align
	return f
}

func (f *FlexComponent) AlignCross(align layout.CrossAxisAlignment) *FlexComponent {
	f.CrossAxis = align
	return f
}

func (f *FlexComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	// Sync children
	if len(node.Children) != len(f.Children) {
		node.Children = make([]*Node, len(f.Children))
		for i, ch := range f.Children {
			node.Children[i] = NewNode(ch)
			node.Children[i].Parent = node
		}
	} else {
		for i, ch := range f.Children {
			node.Children[i].Component = ch
		}
	}

	flexInputs := make([]layout.FlexChildInput, len(node.Children))
	for i, chNode := range node.Children {
		childComp := chNode.Component
		var flexFactor float64
		if exp, ok := childComp.(*ExpandedComponent); ok {
			flexFactor = exp.FlexFactor
		} else if _, ok := childComp.(*SpacerComponent); ok {
			flexFactor = 1.0
		}

		cNode := chNode
		flexInputs[i] = layout.FlexChildInput{
			Flex: flexFactor,
			Measure: func(c layout.BoxConstraints) geom.Size {
				return cNode.Layout(c)
			},
		}
	}

	dir := layout.AxisVertical
	if f.Direction == DirHorizontal {
		dir = layout.AxisHorizontal
	}

	res := layout.ComputeFlex(constraints, layout.FlexConfig{
		Direction: dir,
		MainAxis:  f.MainAxis,
		CrossAxis: f.CrossAxis,
		Gap:       f.Gap,
	}, flexInputs)

	for i, bounds := range res.ChildBounds {
		node.Children[i].Bounds = bounds
	}

	return res.Size
}

func (f *FlexComponent) Paint(node *Node, canvas *render.Canvas) {
	for _, child := range node.Children {
		child.Paint(canvas)
	}
}

// --- Stack Component ---

// StackComponent implements a layered stack layout.
type StackComponent struct {
	BaseComponent
	Alignment layout.Alignment
	Children  []Component
}

// Stack creates a layered stack layout.
func Stack(children ...Component) *StackComponent {
	return &StackComponent{
		Alignment: layout.AlignTopLeft,
		Children:  children,
	}
}

func (s *StackComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	if len(node.Children) != len(s.Children) {
		node.Children = make([]*Node, len(s.Children))
		for i, ch := range s.Children {
			node.Children[i] = NewNode(ch)
			node.Children[i].Parent = node
		}
	}

	stackInputs := make([]layout.StackChildInput, len(node.Children))
	for i, chNode := range node.Children {
		cNode := chNode
		stackInputs[i] = layout.StackChildInput{
			Measure: func(c layout.BoxConstraints) geom.Size {
				return cNode.Layout(c)
			},
		}
	}

	res := layout.ComputeStack(constraints, s.Alignment, stackInputs)
	for i, bounds := range res.ChildBounds {
		node.Children[i].Bounds = bounds
	}
	return res.Size
}

func (s *StackComponent) Paint(node *Node, canvas *render.Canvas) {
	for _, child := range node.Children {
		child.Paint(canvas)
	}
}

// --- Expanded & Spacer ---

type ExpandedComponent struct {
	BaseComponent
	FlexFactor float64
	Child      Component
}

// Expanded expands a child within a flex Row or Column.
func Expanded(child Component) *ExpandedComponent {
	return &ExpandedComponent{FlexFactor: 1.0, Child: child}
}

func (e *ExpandedComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	if len(node.Children) == 0 {
		childNode := NewNode(e.Child)
		childNode.Parent = node
		node.Children = []*Node{childNode}
	} else {
		node.Children[0].Component = e.Child
	}
	sz := node.Children[0].Layout(constraints)
	node.Children[0].Bounds = geom.NewRect(0, 0, sz.Width, sz.Height)
	return sz
}

func (e *ExpandedComponent) Paint(node *Node, canvas *render.Canvas) {
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

type SpacerComponent struct {
	BaseComponent
}

// Spacer creates an empty expandable gap inside a Row or Column.
func Spacer() *SpacerComponent {
	return &SpacerComponent{}
}

func (s *SpacerComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(0, 0))
}

func (s *SpacerComponent) Paint(node *Node, canvas *render.Canvas) {}

// --- Center & Padding ---

type CenterComponent struct {
	BaseComponent
	Child Component
}

// Center centers a child within available space.
func Center(child Component) *CenterComponent {
	return &CenterComponent{Child: child}
}

func (c *CenterComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	if len(node.Children) == 0 {
		childNode := NewNode(c.Child)
		childNode.Parent = node
		node.Children = []*Node{childNode}
	} else {
		node.Children[0].Component = c.Child
	}

	childSz := node.Children[0].Layout(layout.Loose(constraints.MaxSize()))
	containerW := childSz.Width
	if constraints.HasBoundedWidth() {
		containerW = constraints.MaxWidth
	}
	containerH := childSz.Height
	if constraints.HasBoundedHeight() {
		containerH = constraints.MaxHeight
	}
	containerSz := constraints.Constrain(geom.Sz(containerW, containerH))

	offX := (containerSz.Width - childSz.Width) / 2.0
	offY := (containerSz.Height - childSz.Height) / 2.0
	node.Children[0].Bounds = geom.NewRect(offX, offY, childSz.Width, childSz.Height)

	return containerSz
}

func (c *CenterComponent) Paint(node *Node, canvas *render.Canvas) {
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

type PaddingComponent struct {
	BaseComponent
	Insets geom.Insets
	Child  Component
}

// Padding adds insets around a child.
func Padding(insets geom.Insets, child Component) *PaddingComponent {
	return &PaddingComponent{Insets: insets, Child: child}
}

func (p *PaddingComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	if len(node.Children) == 0 {
		childNode := NewNode(p.Child)
		childNode.Parent = node
		node.Children = []*Node{childNode}
	} else {
		node.Children[0].Component = p.Child
	}

	innerConstraints := constraints.Deflate(p.Insets)
	childSz := node.Children[0].Layout(innerConstraints)

	node.Children[0].Bounds = geom.NewRect(p.Insets.Left, p.Insets.Top, childSz.Width, childSz.Height)
	totalW := childSz.Width + p.Insets.Horizontal()
	totalH := childSz.Height + p.Insets.Vertical()
	return constraints.Constrain(geom.Sz(totalW, totalH))
}

func (p *PaddingComponent) Paint(node *Node, canvas *render.Canvas) {
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

// --- Container & Card ---

type ContainerComponent struct {
	BaseComponent
	Width       *float64
	Height      *float64
	Padding     geom.Insets
	Margin      geom.Insets
	Background  color.Color
	BorderColor color.Color
	BorderWidth float64
	Radius      geom.CornerRadius
	Shadow      *render.ShadowParams
	Child       Component
}

// Container creates a styled rectangular container box.
func Container() *ContainerComponent {
	return &ContainerComponent{}
}

func (c *ContainerComponent) WithChild(child Component) *ContainerComponent {
	c.Child = child
	return c
}

func (c *ContainerComponent) Size(width, height float64) *ContainerComponent {
	c.Width = &width
	c.Height = &height
	return c
}

func (c *ContainerComponent) Pad(insets geom.Insets) *ContainerComponent {
	c.Padding = insets
	return c
}

func (c *ContainerComponent) Marg(insets geom.Insets) *ContainerComponent {
	c.Margin = insets
	return c
}

func (c *ContainerComponent) Bg(col color.Color) *ContainerComponent {
	c.Background = col
	return c
}

func (c *ContainerComponent) Border(col color.Color, width float64) *ContainerComponent {
	c.BorderColor = col
	c.BorderWidth = width
	return c
}

func (c *ContainerComponent) Rounded(radius geom.CornerRadius) *ContainerComponent {
	c.Radius = radius
	return c
}

func (c *ContainerComponent) DropShadow(shadow render.ShadowParams) *ContainerComponent {
	c.Shadow = &shadow
	return c
}

func (c *ContainerComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	if c.Child != nil {
		if len(node.Children) == 0 {
			childNode := NewNode(c.Child)
			childNode.Parent = node
			node.Children = []*Node{childNode}
		} else {
			node.Children[0].Component = c.Child
		}
	}

	var measureChild func(layout.BoxConstraints) geom.Size
	if len(node.Children) > 0 {
		measureChild = func(bc layout.BoxConstraints) geom.Size {
			return node.Children[0].Layout(bc)
		}
	}

	deco := layout.BoxDecoration{
		Padding: c.Padding,
		Margin:  c.Margin,
		Width:   c.Width,
		Height:  c.Height,
	}

	containerSz, childBounds := layout.ComputeBoxLayout(constraints, deco, measureChild)
	if len(node.Children) > 0 {
		node.Children[0].Bounds = childBounds
	}

	return containerSz
}

func (c *ContainerComponent) Paint(node *Node, canvas *render.Canvas) {
	innerBounds := geom.NewRect(
		c.Margin.Left,
		c.Margin.Top,
		node.Bounds.Width-c.Margin.Horizontal(),
		node.Bounds.Height-c.Margin.Vertical(),
	)

	// Draw shadow
	if c.Shadow != nil {
		canvas.DrawShadow(innerBounds, c.Radius, *c.Shadow)
	}

	// Draw background
	if c.Background.A > 0 {
		if c.Radius.TopLeft > 0 || c.Radius.TopRight > 0 || c.Radius.BottomRight > 0 || c.Radius.BottomLeft > 0 {
			canvas.FillRoundedRect(innerBounds, c.Radius, c.Background)
		} else {
			canvas.FillRect(innerBounds, c.Background)
		}
	}

	// Draw border
	if c.BorderWidth > 0 && c.BorderColor.A > 0 {
		if c.Radius.TopLeft > 0 || c.Radius.TopRight > 0 || c.Radius.BottomRight > 0 || c.Radius.BottomLeft > 0 {
			canvas.StrokeRoundedRect(innerBounds, c.Radius, c.BorderColor, c.BorderWidth)
		} else {
			canvas.StrokeRect(innerBounds, c.BorderColor, c.BorderWidth)
		}
	}

	// Paint child
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

// --- Text Component ---

type TextComponent struct {
	BaseComponent
	Content    string
	Signal     state.Signal[string]
	Color      color.Color
	FontSize   float64
	FontWeight int
}

// Text creates a text label with string or reactive state.
func Text(value any) *TextComponent {
	tc := &TextComponent{
		FontSize:   14,
		FontWeight: font.WeightRegular,
		Color:      color.White,
	}

	switch v := value.(type) {
	case string:
		tc.Content = v
	case state.Signal[string]:
		tc.Signal = v
	case *state.Value[int]:
		tc.Content = fmt.Sprintf("%d", v.Get())
	case fmt.Stringer:
		tc.Content = v.String()
	default:
		tc.Content = fmt.Sprintf("%v", value)
	}

	return tc
}

func (t *TextComponent) Size(sz float64) *TextComponent {
	t.FontSize = sz
	return t
}

func (t *TextComponent) Weight(w int) *TextComponent {
	t.FontWeight = w
	return t
}

func (t *TextComponent) Col(c color.Color) *TextComponent {
	t.Color = c
	return t
}

func (t *TextComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	str := t.Content
	if t.Signal != nil {
		str = t.Signal.Get()
	}
	sz := text.MeasureText(str, t.FontSize, t.FontWeight)
	return constraints.Constrain(sz)
}

func (t *TextComponent) Paint(node *Node, canvas *render.Canvas) {
	str := t.Content
	if t.Signal != nil {
		str = t.Signal.Get()
	}
	canvas.DrawText(str, geom.Pt(0, 0), t.FontSize, t.FontWeight, t.Color)
}

// --- Button Component ---

type ButtonComponent struct {
	BaseComponent
	Label       string
	OnClickFunc func()
	Variant     ButtonVariant
	Child       Component
}

type ButtonVariant int

const (
	ButtonPrimary ButtonVariant = iota
	ButtonSecondary
	ButtonOutline
	ButtonGhost
	ButtonDanger
)

// Button creates an interactive button.
func Button(label string) *ButtonComponent {
	return &ButtonComponent{
		Label:   label,
		Variant: ButtonPrimary,
	}
}

func (b *ButtonComponent) OnClick(fn func()) *ButtonComponent {
	b.OnClickFunc = fn
	return b
}

func (b *ButtonComponent) Danger() *ButtonComponent {
	b.Variant = ButtonDanger
	return b
}

func (b *ButtonComponent) Secondary() *ButtonComponent {
	b.Variant = ButtonSecondary
	return b
}

func (b *ButtonComponent) Outline() *ButtonComponent {
	b.Variant = ButtonOutline
	return b
}

func (b *ButtonComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	node.OnClick = b.OnClickFunc

	txtSz := text.MeasureText(b.Label, 14, font.WeightMedium)
	btnW := txtSz.Width + 24 // 12px padding each side
	btnH := mathMax(36, txtSz.Height+14)

	return constraints.Constrain(geom.Sz(btnW, btnH))
}

func (b *ButtonComponent) Paint(node *Node, canvas *render.Canvas) {
	t := theme.Current()
	rect := geom.NewRect(0, 0, node.Bounds.Width, node.Bounds.Height)

	bgCol := t.Palette.Primary
	textCol := t.Palette.PrimaryText
	borderCol := color.Transparent

	switch b.Variant {
	case ButtonSecondary:
		bgCol = t.Palette.Secondary
		textCol = t.Palette.SecondaryText
	case ButtonDanger:
		bgCol = t.Palette.Error
		textCol = color.White
	case ButtonOutline:
		bgCol = color.Transparent
		textCol = t.Palette.Primary
		borderCol = t.Palette.Primary
	case ButtonGhost:
		bgCol = color.Transparent
		textCol = t.Palette.TextPrimary
	}

	if node.IsHovered {
		bgCol = bgCol.Lighten(0.1)
	}

	canvas.FillRoundedRect(rect, t.Radii.MD, bgCol)
	if borderCol.A > 0 {
		canvas.StrokeRoundedRect(rect, t.Radii.MD, borderCol, 1.5)
	}

	// Center text
	txtSz := text.MeasureText(b.Label, 14, font.WeightMedium)
	tx := (node.Bounds.Width - txtSz.Width) / 2.0
	ty := (node.Bounds.Height - txtSz.Height) / 2.0
	canvas.DrawText(b.Label, geom.Pt(tx, ty), 14, font.WeightMedium, textCol)
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
