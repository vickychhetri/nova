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

// Geometry helper re-exports keep common layout construction concise for users
// of the ui package while preserving the underlying core/geom value types.
var (
	All            = geom.All
	Symmetric      = geom.Symmetric
	TRBL           = geom.TRBL
	Pt             = geom.Pt
	Sz             = geom.Sz
	RadiusUniform  = geom.RadiusUniform
	RadiusSeparate = geom.RadiusSeparate
)

// --- Flex Components (Column & Row) ---

// FlexComponent implements a flexbox container for horizontal or vertical
// children. Layout delegates measurement and distribution to layout.ComputeFlex.
type FlexComponent struct {
	BaseComponent
	Direction AxisDirection
	MainAxis  layout.MainAxisAlignment
	CrossAxis layout.CrossAxisAlignment
	Gap       float64
	Children  []Component
}

// AxisDirection selects the main direction of a FlexComponent.
type AxisDirection int

const (
	// DirHorizontal lays children out from left to right.
	DirHorizontal AxisDirection = iota
	// DirVertical lays children out from top to bottom.
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

// GapSpacing sets the gap between adjacent flex children and returns f for
// fluent configuration.
func (f *FlexComponent) GapSpacing(gap float64) *FlexComponent {
	f.Gap = gap
	return f
}

// AlignMain sets the main-axis distribution policy and returns f.
func (f *FlexComponent) AlignMain(align layout.MainAxisAlignment) *FlexComponent {
	f.MainAxis = align
	return f
}

// AlignCross sets the cross-axis alignment policy and returns f.
func (f *FlexComponent) AlignCross(align layout.CrossAxisAlignment) *FlexComponent {
	f.CrossAxis = align
	return f
}

// Layout synchronizes component children with node children, measures them, and
// assigns the bounds returned by the flex layout engine.
func (f *FlexComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	// Reuse child nodes when the count is unchanged so their runtime state is
	// retained while the declarative component values are refreshed.
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
		// Expanded and Spacer contribute flex space; ordinary children are
		// measured without a flex factor.
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

// Paint paints flex children in their resolved order. The node's Paint method
// already establishes the node-local canvas translation.
func (f *FlexComponent) Paint(node *Node, canvas *render.Canvas) {
	for _, child := range node.Children {
		child.Paint(canvas)
	}
}

// --- Stack Component ---

// StackComponent implements a layered layout in which children share a
// container and may overlap.
type StackComponent struct {
	BaseComponent
	Alignment layout.Alignment
	Children  []Component
}

// Stack creates a layered stack layout with top-left alignment by default.
func Stack(children ...Component) *StackComponent {
	return &StackComponent{
		Alignment: layout.AlignTopLeft,
		Children:  children,
	}
}

// Layout synchronizes stack children, measures them, and assigns the bounds
// computed by layout.ComputeStack.
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

// Paint draws stack children in order, allowing later children to appear above
// earlier children in painter-style rendering.
func (s *StackComponent) Paint(node *Node, canvas *render.Canvas) {
	for _, child := range node.Children {
		child.Paint(canvas)
	}
}

// --- Expanded & Spacer ---

// ExpandedComponent gives its child a flex factor inside a FlexComponent.
type ExpandedComponent struct {
	BaseComponent
	FlexFactor float64
	Child      Component
}

// Expanded expands child within a flex Row or Column using flex factor 1.
func Expanded(child Component) *ExpandedComponent {
	return &ExpandedComponent{FlexFactor: 1.0, Child: child}
}

// Layout mounts or updates the wrapped child and lays it out using the supplied
// constraints. The flex parent controls the space passed to this component.
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

// Paint delegates drawing to the expanded child.
func (e *ExpandedComponent) Paint(node *Node, canvas *render.Canvas) {
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

// SpacerComponent is an empty component whose flex parent can allocate it
// remaining space.
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

// CenterComponent sizes itself to the available bounded space and positions its
// child at the resulting center.
type CenterComponent struct {
	BaseComponent
	Child Component
}

// Center centers child within the available space.
func Center(child Component) *CenterComponent {
	return &CenterComponent{Child: child}
}

// Layout measures the child loosely, chooses the available bounded container
// size, and assigns the centered child offset.
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

// Paint delegates drawing to the centered child.
func (c *CenterComponent) Paint(node *Node, canvas *render.Canvas) {
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

// PaddingComponent adds edge insets around a single child.
type PaddingComponent struct {
	BaseComponent
	Insets geom.Insets
	Child  Component
}

// Padding creates a component that adds insets around child.
func Padding(insets geom.Insets, child Component) *PaddingComponent {
	return &PaddingComponent{Insets: insets, Child: child}
}

// Layout deflates constraints for the child, then adds the insets back to the
// child's measured size and positions it inside the padding area.
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

// Paint delegates drawing to the padded child.
func (p *PaddingComponent) Paint(node *Node, canvas *render.Canvas) {
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

// --- Container & Card ---

// ContainerComponent describes a styled box with optional size, margin,
// padding, background, border, radius, shadow, and child content.
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

// Container creates an initially unstyled container box for fluent configuration.
func Container() *ContainerComponent {
	return &ContainerComponent{}
}

// WithChild sets the container's single child.
func (c *ContainerComponent) WithChild(child Component) *ContainerComponent {
	c.Child = child
	return c
}

// Size sets explicit width and height constraints for the container.
func (c *ContainerComponent) Size(width, height float64) *ContainerComponent {
	c.Width = &width
	c.Height = &height
	return c
}

// WithWidth sets an explicit container width.
func (c *ContainerComponent) WithWidth(w float64) *ContainerComponent {
	c.Width = &w
	return c
}

// WithHeight sets an explicit container height.
func (c *ContainerComponent) WithHeight(h float64) *ContainerComponent {
	c.Height = &h
	return c
}

// Pad sets the container's inner padding.
func (c *ContainerComponent) Pad(insets geom.Insets) *ContainerComponent {
	c.Padding = insets
	return c
}

// Marg sets the container's outer margin.
func (c *ContainerComponent) Marg(insets geom.Insets) *ContainerComponent {
	c.Margin = insets
	return c
}

// Bg sets the container background color.
func (c *ContainerComponent) Bg(col color.Color) *ContainerComponent {
	c.Background = col
	return c
}

// Border sets the container border color and stroke width.
func (c *ContainerComponent) Border(col color.Color, width float64) *ContainerComponent {
	c.BorderColor = col
	c.BorderWidth = width
	return c
}

// Rounded sets per-corner container radii.
func (c *ContainerComponent) Rounded(radius geom.CornerRadius) *ContainerComponent {
	c.Radius = radius
	return c
}

// DropShadow sets the container shadow parameters.
func (c *ContainerComponent) DropShadow(shadow render.ShadowParams) *ContainerComponent {
	c.Shadow = &shadow
	return c
}

// Layout mounts or updates the optional child, computes the decorated box size,
// and stores the child bounds returned by the box layout engine.
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

// Paint records shadow, background, border, and child commands in that visual
// order. Transparent or non-positive style values skip their corresponding
// drawing operation.
func (c *ContainerComponent) Paint(node *Node, canvas *render.Canvas) {
	innerBounds := geom.NewRect(
		c.Margin.Left,
		c.Margin.Top,
		node.Bounds.Width-c.Margin.Horizontal(),
		node.Bounds.Height-c.Margin.Vertical(),
	)

	// Draw shadow first so subsequent background and border content appears above it.
	if c.Shadow != nil {
		canvas.DrawShadow(innerBounds, c.Radius, *c.Shadow)
	}

	// Draw background only when it has visible alpha.
	if c.Background.A > 0 {
		if c.Radius.TopLeft > 0 || c.Radius.TopRight > 0 || c.Radius.BottomRight > 0 || c.Radius.BottomLeft > 0 {
			canvas.FillRoundedRect(innerBounds, c.Radius, c.Background)
		} else {
			canvas.FillRect(innerBounds, c.Background)
		}
	}

	// Draw border above the background.
	if c.BorderWidth > 0 && c.BorderColor.A > 0 {
		if c.Radius.TopLeft > 0 || c.Radius.TopRight > 0 || c.Radius.BottomRight > 0 || c.Radius.BottomLeft > 0 {
			canvas.StrokeRoundedRect(innerBounds, c.Radius, c.BorderColor, c.BorderWidth)
		} else {
			canvas.StrokeRect(innerBounds, c.BorderColor, c.BorderWidth)
		}
	}

	// Paint child content last so it appears inside the decorated container.
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

// --- Text Component ---

// TextComponent renders static text or a reactive string signal.
type TextComponent struct {
	BaseComponent
	Content        string
	Signal         state.Signal[string]
	Color          color.Color
	hasCustomColor bool
	FontSize       float64
	FontWeight     int
}

// Text creates a text label from a string, reactive string signal, integer state
// value, fmt.Stringer, or any value accepted by fmt.Sprintf.
func Text(value any) *TextComponent {
	tc := &TextComponent{
		FontSize:   14,
		FontWeight: font.WeightRegular,
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

// Size sets the text font size and returns t for fluent configuration.
func (t *TextComponent) Size(sz float64) *TextComponent {
	t.FontSize = sz
	return t
}

// Weight sets the text font weight and returns t.
func (t *TextComponent) Weight(w int) *TextComponent {
	t.FontWeight = w
	return t
}

// Col sets an explicit text color and returns t. Without Col, Paint uses the
// active theme's TextPrimary token.
func (t *TextComponent) Col(c color.Color) *TextComponent {
	t.Color = c
	t.hasCustomColor = true
	return t
}

// Layout measures the current static or signal-provided string and constrains
// the resulting size.
func (t *TextComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	str := t.Content
	if t.Signal != nil {
		str = t.Signal.Get()
	}
	sz := text.MeasureText(str, t.FontSize, t.FontWeight)
	return constraints.Constrain(sz)
}

// Paint reads the current text value and records it using either the custom
// color or the active theme's primary text color.
func (t *TextComponent) Paint(node *Node, canvas *render.Canvas) {
	str := t.Content
	if t.Signal != nil {
		str = t.Signal.Get()
	}
	col := theme.Current().Palette.TextPrimary
	if t.hasCustomColor {
		col = t.Color
	}
	canvas.DrawText(str, geom.Pt(0, 0), t.FontSize, t.FontWeight, col)
}

// --- Button Component ---

// ButtonComponent describes an interactive button with label, optional icon,
// click action, visual variant, and optional custom child content.
type ButtonComponent struct {
	BaseComponent
	Label       string
	Icon        string
	OnClickFunc func()
	Variant     ButtonVariant
	Child       Component
}

// ButtonVariant selects the visual treatment of a ButtonComponent.
type ButtonVariant int

const (
	// ButtonPrimary is the default emphasized button style.
	ButtonPrimary ButtonVariant = iota
	// ButtonSecondary is a supporting button style.
	ButtonSecondary
	// ButtonOutline is an outlined button style.
	ButtonOutline
	// ButtonGhost is a low-emphasis button style.
	ButtonGhost
	// ButtonDanger is intended for destructive actions.
	ButtonDanger
)

// Button creates an interactive primary button with label.
func Button(label string) *ButtonComponent {
	return &ButtonComponent{
		Label:   label,
		Variant: ButtonPrimary,
	}
}

// WithIcon adds a leading icon string to the button display text.
func (b *ButtonComponent) WithIcon(icon string) *ButtonComponent {
	b.Icon = icon
	return b
}

// OnClick registers the callback invoked for a click and returns b.
func (b *ButtonComponent) OnClick(fn func()) *ButtonComponent {
	b.OnClickFunc = fn
	return b
}

// Primary selects the primary button variant.
func (b *ButtonComponent) Primary() *ButtonComponent {
	b.Variant = ButtonPrimary
	return b
}

// Danger selects the destructive-action button variant.
func (b *ButtonComponent) Danger() *ButtonComponent {
	b.Variant = ButtonDanger
	return b
}

// Secondary selects the secondary button variant.
func (b *ButtonComponent) Secondary() *ButtonComponent {
	b.Variant = ButtonSecondary
	return b
}

// Outline selects the outlined button variant.
func (b *ButtonComponent) Outline() *ButtonComponent {
	b.Variant = ButtonOutline
	return b
}

// Ghost selects the low-emphasis ghost button variant.
func (b *ButtonComponent) Ghost() *ButtonComponent {
	b.Variant = ButtonGhost
	return b
}

// displayText returns the label with an optional icon prefix for rendering.
func (b *ButtonComponent) displayText() string {
	if b.Icon != "" {
		return b.Icon + " " + b.Label
	}
	return b.Label
}

func (b *ButtonComponent) Layout(node *Node, constraints layout.BoxConstraints) geom.Size {
	node.OnClick = b.OnClickFunc

	txt := b.displayText()
	txtSz := text.MeasureText(txt, 13, font.WeightMedium)
	btnW := txtSz.Width + 24 // 12px padding each side
	btnH := 34.0

	return constraints.Constrain(geom.Sz(btnW, btnH))
}

func (b *ButtonComponent) Paint(node *Node, canvas *render.Canvas) {
	t := theme.Current()
	rect := geom.NewRect(0, 0, node.Bounds.Width, node.Bounds.Height)
	radius := geom.RadiusUniform(6)

	bgCol := t.Palette.Primary
	textCol := t.Palette.PrimaryText
	borderCol := color.Transparent

	switch b.Variant {
	case ButtonSecondary:
		bgCol = t.Palette.Secondary
		textCol = t.Palette.SecondaryText
		borderCol = t.Palette.Border
		if node.IsHovered {
			bgCol = t.Palette.SecondaryHover
			borderCol = t.Palette.BorderHover
		}
	case ButtonDanger:
		bgCol = t.Palette.Error
		textCol = color.White
		if node.IsHovered {
			bgCol = bgCol.Lighten(0.08)
		}
	case ButtonOutline:
		bgCol = color.Transparent
		textCol = t.Palette.Primary
		borderCol = t.Palette.Border
		if node.IsHovered {
			bgCol = t.Palette.Primary.WithAlpha(0.08)
			borderCol = t.Palette.Primary
		}
	case ButtonGhost:
		bgCol = color.Transparent
		textCol = t.Palette.TextPrimary
		if node.IsHovered {
			bgCol = t.Palette.SurfaceHover
		}
	case ButtonPrimary:
		if node.IsHovered {
			bgCol = t.Palette.PrimaryHover
		}
	}

	if bgCol.A > 0 {
		canvas.FillRoundedRect(rect, radius, bgCol)
	}
	if borderCol.A > 0 {
		canvas.StrokeRoundedRect(rect, radius, borderCol, 1.0)
	}

	// Center text
	txt := b.displayText()
	txtSz := text.MeasureText(txt, 13, font.WeightMedium)
	tx := (node.Bounds.Width - txtSz.Width) / 2.0
	ty := (node.Bounds.Height - txtSz.Height) / 2.0
	canvas.DrawText(txt, geom.Pt(tx, ty), 13, font.WeightMedium, textCol)
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
