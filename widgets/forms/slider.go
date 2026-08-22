package forms

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
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
)

// SliderComponent is an interactive continuous/stepped slider control.
type SliderComponent struct {
	ui.BaseComponent
	Value     *state.Value[float64]
	Min       float64
	Max       float64
	Step      float64
	Width     float64
	OnChanged func(float64)
}

// Slider creates a slider input.
func Slider(min, max float64) *SliderComponent {
	return &SliderComponent{
		Value: state.Float(min),
		Min:   min,
		Max:   max,
		Step:  1.0,
		Width: 200,
	}
}

func (sc *SliderComponent) Bind(val *state.Value[float64]) *SliderComponent {
	sc.Value = val
	return sc
}

func (sc *SliderComponent) WithStep(step float64) *SliderComponent {
	sc.Step = step
	return sc
}

func (sc *SliderComponent) WithWidth(w float64) *SliderComponent {
	sc.Width = w
	return sc
}

func (sc *SliderComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	updateValFromPos := func(x float64) {
		if sc.Value == nil || sc.Max <= sc.Min {
			return
		}
		ratio := math.Max(0, math.Min(1, x/node.Bounds.Width))
		rawVal := sc.Min + ratio*(sc.Max-sc.Min)
		if sc.Step > 0 {
			rawVal = math.Round(rawVal/sc.Step) * sc.Step
		}
		sc.Value.Set(rawVal)
		if sc.OnChanged != nil {
			sc.OnChanged(rawVal)
		}
	}

	node.OnPointerDown = func(e *event.PointerEvent) {
		updateValFromPos(e.Position.X)
	}

	node.OnPointerMove = func(e *event.PointerEvent) {
		if e.Button != 0 {
			updateValFromPos(e.Position.X)
		}
	}

	return constraints.Constrain(geom.Sz(sc.Width, 32))
}

func (sc *SliderComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width
	trackY := 14.0
	trackH := 6.0

	// Background track
	trackRect := geom.NewRect(0, trackY, w, trackH)
	canvas.FillRoundedRect(trackRect, geom.RadiusUniform(3), t.Palette.Secondary)

	val := sc.Min
	if sc.Value != nil {
		val = sc.Value.Get()
	}

	ratio := 0.0
	if sc.Max > sc.Min {
		ratio = (val - sc.Min) / (sc.Max - sc.Min)
	}
	ratio = math.Max(0, math.Min(1, ratio))

	// Active filled track
	activeRect := geom.NewRect(0, trackY, w*ratio, trackH)
	canvas.FillRoundedRect(activeRect, geom.RadiusUniform(3), t.Palette.Primary)

	// Thumb knob
	thumbX := w * ratio
	canvas.FillCircle(geom.Pt(thumbX, trackY+3), 8, color.White)
	canvas.StrokeCircle(geom.Pt(thumbX, trackY+3), 8, t.Palette.Primary, 2.0)
}

// --- Select / Dropdown ---

type SelectOption struct {
	Label string
	Value string
}

type SelectComponent struct {
	ui.BaseComponent
	Label       string
	Placeholder string
	Options     []SelectOption
	Selected    *state.Value[string]
	IsOpen      *state.Value[bool]
	Width       float64
	OnChanged   func(string)
}

// Select creates a dropdown select control.
func Select(options ...SelectOption) *SelectComponent {
	return &SelectComponent{
		Placeholder: "Select an option...",
		Options:     options,
		Selected:    state.String(""),
		IsOpen:      state.Bool(false),
		Width:       220,
	}
}

func (sc *SelectComponent) Bind(val *state.Value[string]) *SelectComponent {
	sc.Selected = val
	return sc
}

func (sc *SelectComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	node.OnClick = func() {
		if sc.IsOpen != nil {
			sc.IsOpen.Set(!sc.IsOpen.Get())
		}
	}

	h := 40.0
	if sc.IsOpen != nil && sc.IsOpen.Get() {
		h += float64(len(sc.Options))*32.0 + 8.0
	}

	return constraints.Constrain(geom.Sz(sc.Width, h))
}

func (sc *SelectComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width

	// Header box
	headerRect := geom.NewRect(0, 0, w, 40)
	canvas.FillRoundedRect(headerRect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(headerRect, t.Radii.MD, t.Palette.Border, 1.5)

	selVal := ""
	if sc.Selected != nil {
		selVal = sc.Selected.Get()
	}

	displayLabel := sc.Placeholder
	for _, opt := range sc.Options {
		if opt.Value == selVal {
			displayLabel = opt.Label
			break
		}
	}

	textCol := t.Palette.TextPrimary
	if selVal == "" {
		textCol = t.Palette.TextMuted
	}
	canvas.DrawText(displayLabel, geom.Pt(12, 12), 14, font.WeightRegular, textCol)

	// Down arrow indicator
	canvas.DrawText("▾", geom.Pt(w-24, 10), 14, font.WeightBold, t.Palette.TextSecondary)

	// Popup options menu if opened
	if sc.IsOpen != nil && sc.IsOpen.Get() {
		menuY := 44.0
		menuH := float64(len(sc.Options))*32.0 + 8.0
		menuRect := geom.NewRect(0, menuY, w, menuH)
		canvas.FillRoundedRect(menuRect, t.Radii.MD, t.Palette.Card)
		canvas.StrokeRoundedRect(menuRect, t.Radii.MD, t.Palette.Border, 1.5)

		for i, opt := range sc.Options {
			optY := menuY + 4.0 + float64(i)*32.0
			optCol := t.Palette.TextPrimary
			if opt.Value == selVal {
				optCol = t.Palette.Primary
				canvas.FillRoundedRect(geom.NewRect(4, optY, w-8, 28), t.Radii.SM, t.Palette.SurfaceHover)
			}
			canvas.DrawText(opt.Label, geom.Pt(12, optY+6), 13, font.WeightRegular, optCol)
		}
	}
}

// --- DatePicker & ColorPicker & FilePicker ---

type DatePickerComponent struct {
	ui.BaseComponent
	Value *state.Value[string]
	Width float64
}

// DatePicker creates a date selection widget.
func DatePicker() *DatePickerComponent {
	return &DatePickerComponent{
		Value: state.String("2026-08-23"),
		Width: 200,
	}
}

func (dp *DatePickerComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(dp.Width, 40))
}

func (dp *DatePickerComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	rect := geom.NewRect(0, 0, node.Bounds.Width, 40)
	canvas.FillRoundedRect(rect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(rect, t.Radii.MD, t.Palette.Border, 1.5)

	valStr := "YYYY-MM-DD"
	if dp.Value != nil && dp.Value.Get() != "" {
		valStr = dp.Value.Get()
	}
	canvas.DrawText("📅 "+valStr, geom.Pt(12, 12), 14, font.WeightRegular, t.Palette.TextPrimary)
}

type ColorPickerComponent struct {
	ui.BaseComponent
	SelectedColor *state.Value[color.Color]
	Width         float64
}

// ColorPicker creates an interactive color selector.
func ColorPicker(initial color.Color) *ColorPickerComponent {
	return &ColorPickerComponent{
		SelectedColor: state.New(initial),
		Width:         220,
	}
}

func (cp *ColorPickerComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(cp.Width, 40))
}

func (cp *ColorPickerComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	rect := geom.NewRect(0, 0, node.Bounds.Width, 40)
	canvas.FillRoundedRect(rect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(rect, t.Radii.MD, t.Palette.Border, 1.5)

	curCol := color.Blue
	if cp.SelectedColor != nil {
		curCol = cp.SelectedColor.Get()
	}

	// Swatch box
	swatch := geom.NewRect(8, 8, 24, 24)
	canvas.FillRoundedRect(swatch, t.Radii.SM, curCol)
	canvas.StrokeRoundedRect(swatch, t.Radii.SM, t.Palette.Border, 1.0)

	canvas.DrawText(curCol.HexString(), geom.Pt(40, 12), 14, font.WeightRegular, t.Palette.TextPrimary)
}

type FilePickerComponent struct {
	ui.BaseComponent
	FilePath *state.Value[string]
	Prompt   string
	Width    float64
}

// FilePicker creates a file upload/selection dropzone.
func FilePicker(prompt string) *FilePickerComponent {
	return &FilePickerComponent{
		FilePath: state.String(""),
		Prompt:   prompt,
		Width:    280,
	}
}

func (fp *FilePickerComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(fp.Width, 70))
}

func (fp *FilePickerComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	rect := geom.NewRect(0, 0, node.Bounds.Width, 70)

	canvas.FillRoundedRect(rect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(rect, t.Radii.MD, t.Palette.Border, 1.5)

	txt := fp.Prompt
	if fp.FilePath != nil && fp.FilePath.Get() != "" {
		txt = fmt.Sprintf("Selected: %s", fp.FilePath.Get())
	}

	canvas.DrawText("📁 "+txt, geom.Pt(20, 26), 13, font.WeightRegular, t.Palette.TextSecondary)
}
