package forms

// Checkbox, radio, and switch controls provide boolean and mutually exclusive
// selection inputs backed by reactive state values.

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

// --- Checkbox ---

type CheckboxComponent struct {
	ui.BaseComponent
	Label         string
	Checked       *state.Value[bool]
	Indeterminate bool
	Disabled      bool
	OnChanged     func(bool)
}

// Checkbox creates a clickable checkbox with a label.
func Checkbox(label string) *CheckboxComponent {
	return &CheckboxComponent{
		Label:   label,
		Checked: state.Bool(false),
	}
}

func (cb *CheckboxComponent) Bind(checked *state.Value[bool]) *CheckboxComponent {
	cb.Checked = checked
	return cb
}

func (cb *CheckboxComponent) OnChange(fn func(bool)) *CheckboxComponent {
	cb.OnChanged = fn
	return cb
}

func (cb *CheckboxComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	node.OnClick = func() {
		if cb.Disabled || cb.Checked == nil {
			return
		}
		newVal := !cb.Checked.Get()
		cb.Checked.Set(newVal)
		if cb.OnChanged != nil {
			cb.OnChanged(newVal)
		}
	}

	txtSz := text.MeasureText(cb.Label, 14, font.WeightRegular)
	totalW := 20.0 + 8.0 + txtSz.Width
	return constraints.Constrain(geom.Sz(totalW, 24))
}

func (cb *CheckboxComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	boxRect := geom.NewRect(0, 2, 20, 20)

	isChecked := cb.Checked != nil && cb.Checked.Get()

	if isChecked {
		canvas.FillRoundedRect(boxRect, t.Radii.SM, t.Palette.Primary)
		// Draw checkmark
		canvas.DrawLine(geom.Pt(4, 12), geom.Pt(8, 16), color.White, 2.0)
		canvas.DrawLine(geom.Pt(8, 16), geom.Pt(16, 6), color.White, 2.0)
	} else if cb.Indeterminate {
		canvas.FillRoundedRect(boxRect, t.Radii.SM, t.Palette.Primary)
		canvas.DrawLine(geom.Pt(5, 12), geom.Pt(15, 12), color.White, 2.0)
	} else {
		canvas.FillRoundedRect(boxRect, t.Radii.SM, t.Palette.Surface)
		borderCol := t.Palette.Border
		if node.IsHovered {
			borderCol = t.Palette.BorderHover
		}
		canvas.StrokeRoundedRect(boxRect, t.Radii.SM, borderCol, 1.5)
	}

	// Label text
	canvas.DrawText(cb.Label, geom.Pt(28, 4), 14, font.WeightRegular, t.Palette.TextPrimary)
}

// --- Radio ---

type RadioComponent struct {
	ui.BaseComponent
	Label     string
	Value     string
	Selected  *state.Value[string]
	OnChanged func(string)
}

// Radio creates a single radio button option.
func Radio(label string, value string, selected *state.Value[string]) *RadioComponent {
	return &RadioComponent{
		Label:    label,
		Value:    value,
		Selected: selected,
	}
}

func (rc *RadioComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	node.OnClick = func() {
		if rc.Selected != nil {
			rc.Selected.Set(rc.Value)
			if rc.OnChanged != nil {
				rc.OnChanged(rc.Value)
			}
		}
	}

	txtSz := text.MeasureText(rc.Label, 14, font.WeightRegular)
	return constraints.Constrain(geom.Sz(20+8+txtSz.Width, 24))
}

func (rc *RadioComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	center := geom.Pt(10, 12)
	isSelected := rc.Selected != nil && rc.Selected.Get() == rc.Value

	if isSelected {
		canvas.StrokeCircle(center, 9, t.Palette.Primary, 2.0)
		canvas.FillCircle(center, 5, t.Palette.Primary)
	} else {
		borderCol := t.Palette.Border
		if node.IsHovered {
			borderCol = t.Palette.BorderHover
		}
		canvas.FillCircle(center, 9, t.Palette.Surface)
		canvas.StrokeCircle(center, 9, borderCol, 1.5)
	}

	canvas.DrawText(rc.Label, geom.Pt(28, 4), 14, font.WeightRegular, t.Palette.TextPrimary)
}

// --- Switch / Toggle ---

type SwitchComponent struct {
	ui.BaseComponent
	Label     string
	Checked   *state.Value[bool]
	OnChanged func(bool)
}

// Switch creates a toggle switch with smooth state representation.
func Switch(label string) *SwitchComponent {
	return &SwitchComponent{
		Label:   label,
		Checked: state.Bool(false),
	}
}

func (s *SwitchComponent) Bind(val *state.Value[bool]) *SwitchComponent {
	s.Checked = val
	return s
}

func (s *SwitchComponent) OnChange(fn func(bool)) *SwitchComponent {
	s.OnChanged = fn
	return s
}

func (s *SwitchComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	node.OnClick = func() {
		if s.Checked != nil {
			newVal := !s.Checked.Get()
			s.Checked.Set(newVal)
			if s.OnChanged != nil {
				s.OnChanged(newVal)
			}
		}
	}

	txtSz := text.MeasureText(s.Label, 14, font.WeightRegular)
	return constraints.Constrain(geom.Sz(44+10+txtSz.Width, 26))
}

func (s *SwitchComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	isOn := s.Checked != nil && s.Checked.Get()

	trackRect := geom.NewRect(0, 3, 44, 22)
	trackCol := t.Palette.Secondary
	if isOn {
		trackCol = t.Palette.Primary
	}

	canvas.FillRoundedRect(trackRect, geom.RadiusUniform(11), trackCol)

	// Thumb circle
	thumbX := 11.0
	if isOn {
		thumbX = 33.0
	}
	canvas.FillCircle(geom.Pt(thumbX, 14), 8, color.White)

	if s.Label != "" {
		canvas.DrawText(s.Label, geom.Pt(54, 5), 14, font.WeightRegular, t.Palette.TextPrimary)
	}
}
