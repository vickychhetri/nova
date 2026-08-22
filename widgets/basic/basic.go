package basic

import (
	"fmt"
	"math"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/text"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
)

// --- Badge / Tag ---

type BadgeComponent struct {
	ui.BaseComponent
	Text    string
	Color   color.Color
	Variant BadgeVariant
}

type BadgeVariant int

const (
	BadgeDefault BadgeVariant = iota
	BadgeSuccess
	BadgeWarning
	BadgeError
	BadgeInfo
)

// Badge creates a small pill/badge label.
func Badge(text string) *BadgeComponent {
	return &BadgeComponent{
		Text:    text,
		Variant: BadgeDefault,
	}
}

func (b *BadgeComponent) Success() *BadgeComponent { b.Variant = BadgeSuccess; return b }
func (b *BadgeComponent) Warning() *BadgeComponent { b.Variant = BadgeWarning; return b }
func (b *BadgeComponent) Error() *BadgeComponent   { b.Variant = BadgeError; return b }
func (b *BadgeComponent) Info() *BadgeComponent    { b.Variant = BadgeInfo; return b }

func (b *BadgeComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	txtSz := text.MeasureText(b.Text, 11, font.WeightMedium)
	return constraints.Constrain(geom.Sz(txtSz.Width+16, 20))
}

func (b *BadgeComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	rect := geom.NewRect(0, 0, node.Bounds.Width, node.Bounds.Height)

	bgCol := t.Palette.Secondary
	textCol := t.Palette.SecondaryText

	switch b.Variant {
	case BadgeSuccess:
		bgCol = t.Palette.Success.WithAlpha(0.2)
		textCol = t.Palette.Success
	case BadgeWarning:
		bgCol = t.Palette.Warning.WithAlpha(0.2)
		textCol = t.Palette.Warning
	case BadgeError:
		bgCol = t.Palette.Error.WithAlpha(0.2)
		textCol = t.Palette.Error
	case BadgeInfo:
		bgCol = t.Palette.Info.WithAlpha(0.2)
		textCol = t.Palette.Info
	}

	canvas.FillRoundedRect(rect, geom.RadiusUniform(10), bgCol)
	txtSz := text.MeasureText(b.Text, 11, font.WeightMedium)
	tx := (node.Bounds.Width - txtSz.Width) / 2.0
	ty := (node.Bounds.Height - txtSz.Height) / 2.0
	canvas.DrawText(b.Text, geom.Pt(tx, ty), 11, font.WeightMedium, textCol)
}

// --- Avatar ---

type AvatarComponent struct {
	ui.BaseComponent
	Initials string
	Size     float64
}

// Avatar creates a user avatar circle.
func Avatar(initials string) *AvatarComponent {
	return &AvatarComponent{
		Initials: initials,
		Size:     36,
	}
}

func (a *AvatarComponent) WithSize(sz float64) *AvatarComponent {
	a.Size = sz
	return a
}

func (a *AvatarComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(a.Size, a.Size))
}

func (a *AvatarComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	center := geom.Pt(a.Size/2.0, a.Size/2.0)
	canvas.FillCircle(center, a.Size/2.0, t.Palette.Primary)

	fontSize := a.Size * 0.4
	txtSz := text.MeasureText(a.Initials, fontSize, font.WeightBold)
	tx := (a.Size - txtSz.Width) / 2.0
	ty := (a.Size - txtSz.Height) / 2.0
	canvas.DrawText(a.Initials, geom.Pt(tx, ty), fontSize, font.WeightBold, color.White)
}

// --- Spinner ---

type SpinnerComponent struct {
	ui.BaseComponent
	Size float64
}

// Spinner creates an animated loading spinner.
func Spinner() *SpinnerComponent {
	return &SpinnerComponent{Size: 24}
}

func (s *SpinnerComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(s.Size, s.Size))
}

func (s *SpinnerComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	center := geom.Pt(s.Size/2.0, s.Size/2.0)
	canvas.StrokeCircle(center, s.Size/2.0-2, t.Palette.Secondary, 3.0)
	canvas.StrokeCircle(center, s.Size/2.0-2, t.Palette.Primary, 3.0)
}

// --- ProgressBar ---

type ProgressBarComponent struct {
	ui.BaseComponent
	Progress float64 // 0.0 to 1.0
	Width    float64
	Height   float64
}

// Progress creates a linear progress bar.
func Progress(value float64) *ProgressBarComponent {
	return &ProgressBarComponent{
		Progress: math.Max(0, math.Min(1, value)),
		Width:    200,
		Height:   8,
	}
}

func (p *ProgressBarComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(p.Width, p.Height))
}

func (p *ProgressBarComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width
	h := node.Bounds.Height
	r := geom.RadiusUniform(h / 2.0)

	// Background
	canvas.FillRoundedRect(geom.NewRect(0, 0, w, h), r, t.Palette.Secondary)

	// Filled portion
	if p.Progress > 0 {
		fillW := w * p.Progress
		canvas.FillRoundedRect(geom.NewRect(0, 0, fillW, h), r, t.Palette.Primary)
	}
}

// --- Card ---

type CardComponent struct {
	ui.BaseComponent
	Title    string
	Subtitle string
	Child    ui.Component
	Padding  geom.Insets
}

// Card creates a container card with optional title and shadow.
func Card(title string, child ui.Component) *CardComponent {
	return &CardComponent{
		Title:   title,
		Child:   child,
		Padding: geom.All(16),
	}
}

func (c *CardComponent) WithSubtitle(sub string) *CardComponent {
	c.Subtitle = sub
	return c
}

func (c *CardComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	var children []ui.Component
	if c.Title != "" {
		children = append(children, ui.Text(c.Title).Size(16).Weight(font.WeightBold))
	}
	if c.Subtitle != "" {
		children = append(children, ui.Text(c.Subtitle).Size(13).Col(theme.Current().Palette.TextMuted))
	}
	if c.Child != nil {
		children = append(children, c.Child)
	}

	flex := ui.Column(children...).GapSpacing(8)
	container := ui.Container().
		Pad(c.Padding).
		Bg(theme.Current().Palette.Card).
		Border(theme.Current().Palette.Border, 1.0).
		Rounded(theme.Current().Radii.LG).
		WithChild(flex)

	node.Children = []*ui.Node{ui.NewNode(container)}
	node.Children[0].Parent = node
	return node.Children[0].Layout(constraints)
}

func (c *CardComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	if len(node.Children) > 0 {
		node.Children[0].Paint(canvas)
	}
}

// --- Skeleton Loader ---

type SkeletonComponent struct {
	ui.BaseComponent
	Width  float64
	Height float64
}

// Skeleton creates a placeholder loader box.
func Skeleton(width, height float64) *SkeletonComponent {
	return &SkeletonComponent{Width: width, Height: height}
}

func (s *SkeletonComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(s.Width, s.Height))
}

func (s *SkeletonComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	canvas.FillRoundedRect(geom.NewRect(0, 0, node.Bounds.Width, node.Bounds.Height), t.Radii.MD, t.Palette.SurfaceHover)
}

// Format number helper
func FormatCount(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
