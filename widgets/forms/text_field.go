package forms

import (
	"fmt"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/input"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/text"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
)

// TextFieldComponent is a professional single-line text input.
type TextFieldComponent struct {
	ui.BaseComponent
	Label       string
	Placeholder string
	Value       *state.Value[string]
	Error       string
	Disabled    bool
	Width       float64
	OnChanged   func(string)
}

// TextField creates a single-line text input field.
func TextField(placeholder string) *TextFieldComponent {
	return &TextFieldComponent{
		Placeholder: placeholder,
		Value:       state.String(""),
		Width:       240,
	}
}

func (tf *TextFieldComponent) WithLabel(label string) *TextFieldComponent {
	tf.Label = label
	return tf
}

func (tf *TextFieldComponent) Bind(val *state.Value[string]) *TextFieldComponent {
	tf.Value = val
	return tf
}

func (tf *TextFieldComponent) WithError(err string) *TextFieldComponent {
	tf.Error = err
	return tf
}

func (tf *TextFieldComponent) WithWidth(w float64) *TextFieldComponent {
	tf.Width = w
	return tf
}

func (tf *TextFieldComponent) OnChange(fn func(string)) *TextFieldComponent {
	tf.OnChanged = fn
	return tf
}

func (tf *TextFieldComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := 40.0
	if tf.Label != "" {
		h += 20.0
	}
	if tf.Error != "" {
		h += 18.0
	}

	w := tf.Width
	if constraints.HasBoundedWidth() && tf.Width > constraints.MaxWidth {
		w = constraints.MaxWidth
	}

	node.OnKeyDown = func(e *event.KeyEvent) {
		if tf.Disabled || tf.Value == nil {
			return
		}
		cur := tf.Value.Get()
		if e.Key == input.KeyBackspace {
			if len(cur) > 0 {
				newVal := cur[:len(cur)-1]
				tf.Value.Set(newVal)
				if tf.OnChanged != nil {
					tf.OnChanged(newVal)
				}
			}
		} else if e.Rune >= 32 && e.Rune != 127 {
			newVal := cur + string(e.Rune)
			tf.Value.Set(newVal)
			if tf.OnChanged != nil {
				tf.OnChanged(newVal)
			}
		}
	}

	return constraints.Constrain(geom.Sz(w, h))
}

func (tf *TextFieldComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	yOffset := 0.0

	// Draw label if present
	if tf.Label != "" {
		canvas.DrawText(tf.Label, geom.Pt(0, 0), 13, font.WeightMedium, t.Palette.TextSecondary)
		yOffset += 20.0
	}

	fieldRect := geom.NewRect(0, yOffset, node.Bounds.Width, 40)

	// Border color
	borderCol := t.Palette.Border
	if tf.Error != "" {
		borderCol = t.Palette.Error
	} else if node.IsFocused {
		borderCol = t.Palette.BorderFocus
	} else if node.IsHovered {
		borderCol = t.Palette.BorderHover
	}

	canvas.FillRoundedRect(fieldRect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(fieldRect, t.Radii.MD, borderCol, 1.5)

	valStr := ""
	if tf.Value != nil {
		valStr = tf.Value.Get()
	}

	textY := yOffset + 12
	if valStr != "" {
		canvas.DrawText(valStr, geom.Pt(12, textY), 14, font.WeightRegular, t.Palette.TextPrimary)
		// Draw cursor if focused
		if node.IsFocused {
			curSz := text.MeasureText(valStr, 14, font.WeightRegular)
			canvas.DrawLine(geom.Pt(12+curSz.Width+2, yOffset+8), geom.Pt(12+curSz.Width+2, yOffset+32), t.Palette.Primary, 1.5)
		}
	} else {
		canvas.DrawText(tf.Placeholder, geom.Pt(12, textY), 14, font.WeightRegular, t.Palette.TextMuted)
	}

	// Draw error text
	if tf.Error != "" {
		canvas.DrawText(tf.Error, geom.Pt(2, yOffset+44), 11, font.WeightRegular, t.Palette.Error)
	}
}

// --- PasswordField ---

type PasswordFieldComponent struct {
	ui.BaseComponent
	Label       string
	Placeholder string
	Value       *state.Value[string]
	ShowText    *state.Value[bool]
	Error       string
	Width       float64
}

// PasswordField creates a password input field with show/hide toggle.
func PasswordField(placeholder string) *PasswordFieldComponent {
	return &PasswordFieldComponent{
		Placeholder: placeholder,
		Value:       state.String(""),
		ShowText:    state.Bool(false),
		Width:       240,
	}
}

func (pf *PasswordFieldComponent) WithLabel(label string) *PasswordFieldComponent {
	pf.Label = label
	return pf
}

func (pf *PasswordFieldComponent) Bind(val *state.Value[string]) *PasswordFieldComponent {
	pf.Value = val
	return pf
}

func (pf *PasswordFieldComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := 40.0
	if pf.Label != "" {
		h += 20.0
	}
	if pf.Error != "" {
		h += 18.0
	}
	return constraints.Constrain(geom.Sz(pf.Width, h))
}

func (pf *PasswordFieldComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	yOffset := 0.0
	if pf.Label != "" {
		canvas.DrawText(pf.Label, geom.Pt(0, 0), 13, font.WeightMedium, t.Palette.TextSecondary)
		yOffset += 20.0
	}

	fieldRect := geom.NewRect(0, yOffset, node.Bounds.Width, 40)
	canvas.FillRoundedRect(fieldRect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(fieldRect, t.Radii.MD, t.Palette.Border, 1.5)

	valStr := ""
	if pf.Value != nil {
		valStr = pf.Value.Get()
	}

	textY := yOffset + 12
	if valStr != "" {
		display := valStr
		if !pf.ShowText.Get() {
			display = ""
			for range valStr {
				display += "•"
			}
		}
		canvas.DrawText(display, geom.Pt(12, textY), 14, font.WeightRegular, t.Palette.TextPrimary)
	} else {
		canvas.DrawText(pf.Placeholder, geom.Pt(12, textY), 14, font.WeightRegular, t.Palette.TextMuted)
	}
}

// --- TextArea ---

type TextAreaComponent struct {
	ui.BaseComponent
	Label       string
	Placeholder string
	Value       *state.Value[string]
	Rows        int
	Width       float64
}

// TextArea creates a multiline text area.
func TextArea(placeholder string) *TextAreaComponent {
	return &TextAreaComponent{
		Placeholder: placeholder,
		Value:       state.String(""),
		Rows:        4,
		Width:       320,
	}
}

func (ta *TextAreaComponent) WithLabel(label string) *TextAreaComponent {
	ta.Label = label
	return ta
}

func (ta *TextAreaComponent) Bind(val *state.Value[string]) *TextAreaComponent {
	ta.Value = val
	return ta
}

func (ta *TextAreaComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := float64(ta.Rows)*22.0 + 24.0
	if ta.Label != "" {
		h += 20.0
	}
	return constraints.Constrain(geom.Sz(ta.Width, h))
}

func (ta *TextAreaComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	yOffset := 0.0
	if ta.Label != "" {
		canvas.DrawText(ta.Label, geom.Pt(0, 0), 13, font.WeightMedium, t.Palette.TextSecondary)
		yOffset += 20.0
	}

	boxH := float64(ta.Rows)*22.0 + 24.0
	fieldRect := geom.NewRect(0, yOffset, node.Bounds.Width, boxH)
	canvas.FillRoundedRect(fieldRect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(fieldRect, t.Radii.MD, t.Palette.Border, 1.5)

	valStr := ""
	if ta.Value != nil {
		valStr = ta.Value.Get()
	}

	if valStr != "" {
		lines := text.WrapLines(valStr, node.Bounds.Width-24, 13, font.WeightRegular)
		curY := yOffset + 12
		for _, line := range lines {
			canvas.DrawText(line, geom.Pt(12, curY), 13, font.WeightRegular, t.Palette.TextPrimary)
			curY += 20
		}
	} else {
		canvas.DrawText(ta.Placeholder, geom.Pt(12, yOffset+12), 13, font.WeightRegular, t.Palette.TextMuted)
	}
}

// --- NumberInput ---

type NumberInputComponent struct {
	ui.BaseComponent
	Label    string
	Value    *state.Value[float64]
	Min      float64
	Max      float64
	Step     float64
	Width    float64
}

// NumberInput creates a numeric stepper input.
func NumberInput(initial float64) *NumberInputComponent {
	return &NumberInputComponent{
		Value: state.Float(initial),
		Min:   -1000000,
		Max:   1000000,
		Step:  1.0,
		Width: 160,
	}
}

func (ni *NumberInputComponent) WithMinMax(min, max float64) *NumberInputComponent {
	ni.Min = min
	ni.Max = max
	return ni
}

func (ni *NumberInputComponent) WithStep(step float64) *NumberInputComponent {
	ni.Step = step
	return ni
}

func (ni *NumberInputComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(ni.Width, 40))
}

func (ni *NumberInputComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	rect := geom.NewRect(0, 0, node.Bounds.Width, 40)
	canvas.FillRoundedRect(rect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(rect, t.Radii.MD, t.Palette.Border, 1.5)

	val := 0.0
	if ni.Value != nil {
		val = ni.Value.Get()
	}

	valStr := fmt.Sprintf("%.0f", val)
	canvas.DrawText(valStr, geom.Pt(12, 12), 14, font.WeightRegular, t.Palette.TextPrimary)

	// Step buttons
	btnW := 28.0
	minusRect := geom.NewRect(node.Bounds.Width-btnW*2-4, 4, btnW, 32)
	plusRect := geom.NewRect(node.Bounds.Width-btnW-2, 4, btnW, 32)

	canvas.FillRoundedRect(minusRect, t.Radii.SM, t.Palette.Secondary)
	canvas.DrawText("-", geom.Pt(minusRect.X+10, minusRect.Y+8), 16, font.WeightBold, t.Palette.SecondaryText)

	canvas.FillRoundedRect(plusRect, t.Radii.SM, t.Palette.Secondary)
	canvas.DrawText("+", geom.Pt(plusRect.X+9, plusRect.Y+8), 16, font.WeightBold, t.Palette.SecondaryText)
}
