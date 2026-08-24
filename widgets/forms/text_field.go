package forms

// Text-field and editor-style input controls are implemented in this file,
// including clipboard integration and reactive value binding.

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

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

var (
	clipMu            sync.RWMutex
	internalClipboard string
)

// ReadClipboard retrieves text from the system clipboard or internal fallback.
func ReadClipboard() string {
	clipMu.RLock()
	fallback := internalClipboard
	clipMu.RUnlock()

	// 1. Wayland wl-paste
	if out, err := exec.Command("wl-paste", "--no-newline").Output(); err == nil && len(out) > 0 {
		return string(out)
	}
	if out, err := exec.Command("wl-paste").Output(); err == nil && len(out) > 0 {
		return strings.TrimRight(string(out), "\r\n")
	}
	if out, err := exec.Command("wl-paste", "-t", "text/plain").Output(); err == nil && len(out) > 0 {
		return string(out)
	}

	// 2. X11 xclip
	if out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output(); err == nil && len(out) > 0 {
		return string(out)
	}
	if out, err := exec.Command("xclip", "-selection", "primary", "-o").Output(); err == nil && len(out) > 0 {
		return string(out)
	}

	// 3. X11 xsel
	if out, err := exec.Command("xsel", "-b", "-o").Output(); err == nil && len(out) > 0 {
		return string(out)
	}
	if out, err := exec.Command("xsel", "-p", "-o").Output(); err == nil && len(out) > 0 {
		return string(out)
	}

	// 4. macOS pbpaste
	if out, err := exec.Command("pbpaste").Output(); err == nil && len(out) > 0 {
		return string(out)
	}

	// 5. Windows powershell
	if out, err := exec.Command("powershell", "-command", "Get-Clipboard").Output(); err == nil && len(out) > 0 {
		return strings.TrimRight(string(out), "\r\n")
	}

	return fallback
}

// WriteClipboard writes text to the system clipboard and internal fallback.
func WriteClipboard(val string) {
	clipMu.Lock()
	internalClipboard = val
	clipMu.Unlock()

	// 1. Try wl-copy
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(val)
	_ = cmd.Run()

	// 2. Try xclip clipboard
	cmd2 := exec.Command("xclip", "-selection", "clipboard")
	cmd2.Stdin = strings.NewReader(val)
	_ = cmd2.Run()

	// 3. Try xclip primary
	cmd2b := exec.Command("xclip", "-selection", "primary")
	cmd2b.Stdin = strings.NewReader(val)
	_ = cmd2b.Run()

	// 4. Try xsel
	cmd3 := exec.Command("xsel", "-b", "-i")
	cmd3.Stdin = strings.NewReader(val)
	_ = cmd3.Run()

	// 5. Try pbcopy
	cmd4 := exec.Command("pbcopy")
	cmd4.Stdin = strings.NewReader(val)
	_ = cmd4.Run()
}

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
	OnSubmitted func()
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

func (tf *TextFieldComponent) SetDisabled(d bool) *TextFieldComponent {
	tf.Disabled = d
	return tf
}

func (tf *TextFieldComponent) OnChange(fn func(string)) *TextFieldComponent {
	tf.OnChanged = fn
	return tf
}

func (tf *TextFieldComponent) OnSubmit(fn func()) *TextFieldComponent {
	tf.OnSubmitted = fn
	return tf
}

func (tf *TextFieldComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := 34.0
	if tf.Label != "" {
		h += 18.0
	}
	if tf.Error != "" {
		h += 16.0
	}

	w := tf.Width
	if constraints.HasBoundedWidth() && (tf.Width <= 0 || tf.Width > constraints.MaxWidth) {
		w = constraints.MaxWidth
	}

	node.OnPointerDown = func(e *event.PointerEvent) {}

	node.OnKeyDown = func(e *event.KeyEvent) {
		if tf.Disabled || tf.Value == nil {
			return
		}
		cur := tf.Value.Get()

		// Check for Paste (Ctrl+V, Cmd+V, or ASCII rune 22)
		isPaste := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyV || e.Rune == 'v' || e.Rune == 'V') || e.Rune == 22
		// Check for Copy (Ctrl+C, Cmd+C, or ASCII rune 3)
		isCopy := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyC || e.Rune == 'c' || e.Rune == 'C') || e.Rune == 3
		// Check for Cut (Ctrl+X, Cmd+X, or ASCII rune 24)
		isCut := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyX || e.Rune == 'x' || e.Rune == 'X') || e.Rune == 24
		// Check for Select All / Clear (Ctrl+A, Cmd+A, or ASCII rune 1)
		isSelectAll := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyA || e.Rune == 'a' || e.Rune == 'A') || e.Rune == 1

		if isPaste {
			clip := ReadClipboard()
			if clip != "" {
				newVal := cur + clip
				tf.Value.Set(newVal)
				if tf.OnChanged != nil {
					tf.OnChanged(newVal)
				}
			}
		} else if isCopy {
			if cur != "" {
				WriteClipboard(cur)
			}
		} else if isCut {
			if cur != "" {
				WriteClipboard(cur)
				tf.Value.Set("")
				if tf.OnChanged != nil {
					tf.OnChanged("")
				}
			}
		} else if isSelectAll {
			// Can clear or mark
		} else if e.Key == input.KeyBackspace || e.Rune == 8 || e.Rune == 127 {
			if len(cur) > 0 {
				newVal := cur[:len(cur)-1]
				tf.Value.Set(newVal)
				if tf.OnChanged != nil {
					tf.OnChanged(newVal)
				}
			}
		} else if e.Key == input.KeyEnter || e.Rune == '\r' || e.Rune == '\n' {
			if tf.OnSubmitted != nil {
				tf.OnSubmitted()
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
		canvas.DrawText(tf.Label, geom.Pt(0, 0), 12, font.WeightMedium, t.Palette.TextSecondary)
		yOffset += 18.0
	}

	boxH := 34.0
	fieldRect := geom.NewRect(0, yOffset, node.Bounds.Width, boxH)
	radius := geom.RadiusUniform(6)

	// Border color
	borderCol := t.Palette.Border
	borderWidth := 1.0
	if tf.Error != "" {
		borderCol = t.Palette.Error
	} else if node.IsFocused {
		borderCol = t.Palette.Primary
		borderWidth = 1.5
	} else if node.IsHovered {
		borderCol = t.Palette.BorderHover
	}

	canvas.FillRoundedRect(fieldRect, radius, t.Palette.Surface)
	canvas.StrokeRoundedRect(fieldRect, radius, borderCol, borderWidth)

	valStr := ""
	if tf.Value != nil {
		valStr = tf.Value.Get()
	}

	textY := yOffset + 9.0
	if valStr != "" {
		canvas.DrawText(valStr, geom.Pt(10, textY), 13, font.WeightRegular, t.Palette.TextPrimary)
		if node.IsFocused {
			curSz := text.MeasureText(valStr, 13, font.WeightRegular)
			canvas.DrawLine(geom.Pt(10+curSz.Width+1, yOffset+6), geom.Pt(10+curSz.Width+1, yOffset+28), t.Palette.Primary, 1.5)
		}
	} else {
		canvas.DrawText(tf.Placeholder, geom.Pt(10, textY), 13, font.WeightRegular, t.Palette.TextMuted)
		if node.IsFocused {
			canvas.DrawLine(geom.Pt(10, yOffset+6), geom.Pt(10, yOffset+28), t.Palette.Primary, 1.5)
		}
	}

	// Draw error text
	if tf.Error != "" {
		canvas.DrawText(tf.Error, geom.Pt(2, yOffset+38), 11, font.WeightRegular, t.Palette.Error)
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
	OnSubmitted func()
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

func (pf *PasswordFieldComponent) WithWidth(w float64) *PasswordFieldComponent {
	pf.Width = w
	return pf
}

func (pf *PasswordFieldComponent) OnSubmit(fn func()) *PasswordFieldComponent {
	pf.OnSubmitted = fn
	return pf
}

func (pf *PasswordFieldComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := 34.0
	if pf.Label != "" {
		h += 18.0
	}
	if pf.Error != "" {
		h += 16.0
	}

	node.OnPointerDown = func(e *event.PointerEvent) {}

	node.OnKeyDown = func(e *event.KeyEvent) {
		if pf.Value == nil {
			return
		}
		cur := pf.Value.Get()

		// Check for Paste
		isPaste := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyV || e.Rune == 'v' || e.Rune == 'V') || e.Rune == 22
		// Check for Copy
		isCopy := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyC || e.Rune == 'c' || e.Rune == 'C') || e.Rune == 3
		// Check for Cut
		isCut := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyX || e.Rune == 'x' || e.Rune == 'X') || e.Rune == 24

		if isPaste {
			clip := ReadClipboard()
			if clip != "" {
				pf.Value.Set(cur + clip)
			}
		} else if isCopy {
			if cur != "" {
				WriteClipboard(cur)
			}
		} else if isCut {
			if cur != "" {
				WriteClipboard(cur)
				pf.Value.Set("")
			}
		} else if e.Key == input.KeyBackspace || e.Rune == 8 || e.Rune == 127 {
			if len(cur) > 0 {
				newVal := cur[:len(cur)-1]
				pf.Value.Set(newVal)
			}
		} else if e.Key == input.KeyEnter || e.Rune == '\r' || e.Rune == '\n' {
			if pf.OnSubmitted != nil {
				pf.OnSubmitted()
			}
		} else if e.Rune >= 32 && e.Rune != 127 {
			newVal := cur + string(e.Rune)
			pf.Value.Set(newVal)
		}
	}

	return constraints.Constrain(geom.Sz(pf.Width, h))
}

func (pf *PasswordFieldComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	yOffset := 0.0
	if pf.Label != "" {
		canvas.DrawText(pf.Label, geom.Pt(0, 0), 12, font.WeightMedium, t.Palette.TextSecondary)
		yOffset += 18.0
	}

	fieldRect := geom.NewRect(0, yOffset, node.Bounds.Width, 34)
	borderCol := t.Palette.Border
	borderWidth := 1.0
	if node.IsFocused {
		borderCol = t.Palette.Primary
		borderWidth = 1.5
	} else if node.IsHovered {
		borderCol = t.Palette.BorderHover
	}

	canvas.FillRoundedRect(fieldRect, geom.RadiusUniform(6), t.Palette.Surface)
	canvas.StrokeRoundedRect(fieldRect, geom.RadiusUniform(6), borderCol, borderWidth)

	valStr := ""
	if pf.Value != nil {
		valStr = pf.Value.Get()
	}

	textY := yOffset + 9.0
	if valStr != "" {
		display := valStr
		if !pf.ShowText.Get() {
			display = ""
			for range valStr {
				display += "•"
			}
		}
		canvas.DrawText(display, geom.Pt(10, textY), 13, font.WeightRegular, t.Palette.TextPrimary)
		if node.IsFocused {
			curSz := text.MeasureText(display, 13, font.WeightRegular)
			canvas.DrawLine(geom.Pt(10+curSz.Width+1, yOffset+6), geom.Pt(10+curSz.Width+1, yOffset+28), t.Palette.Primary, 1.5)
		}
	} else {
		canvas.DrawText(pf.Placeholder, geom.Pt(10, textY), 13, font.WeightRegular, t.Palette.TextMuted)
		if node.IsFocused {
			canvas.DrawLine(geom.Pt(10, yOffset+6), geom.Pt(10, yOffset+28), t.Palette.Primary, 1.5)
		}
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

func (ta *TextAreaComponent) WithWidth(w float64) *TextAreaComponent {
	ta.Width = w
	return ta
}

func (ta *TextAreaComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := float64(ta.Rows)*22.0 + 24.0
	if ta.Label != "" {
		h += 18.0
	}

	node.OnPointerDown = func(e *event.PointerEvent) {}

	node.OnKeyDown = func(e *event.KeyEvent) {
		if ta.Value == nil {
			return
		}
		cur := ta.Value.Get()

		// Check for Paste
		isPaste := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyV || e.Rune == 'v' || e.Rune == 'V') || e.Rune == 22
		// Check for Copy
		isCopy := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyC || e.Rune == 'c' || e.Rune == 'C') || e.Rune == 3
		// Check for Cut
		isCut := (e.Mods.Has(input.ModCtrl) || e.Mods.Has(input.ModMeta)) && (e.Key == input.KeyX || e.Rune == 'x' || e.Rune == 'X') || e.Rune == 24

		if isPaste {
			clip := ReadClipboard()
			if clip != "" {
				ta.Value.Set(cur + clip)
			}
		} else if isCopy {
			if cur != "" {
				WriteClipboard(cur)
			}
		} else if isCut {
			if cur != "" {
				WriteClipboard(cur)
				ta.Value.Set("")
			}
		} else if e.Key == input.KeyBackspace || e.Rune == 8 || e.Rune == 127 {
			if len(cur) > 0 {
				newVal := cur[:len(cur)-1]
				ta.Value.Set(newVal)
			}
		} else if e.Key == input.KeyEnter || e.Rune == '\n' || e.Rune == '\r' {
			ta.Value.Set(cur + "\n")
		} else if e.Rune >= 32 && e.Rune != 127 {
			newVal := cur + string(e.Rune)
			ta.Value.Set(newVal)
		}
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
	borderCol := t.Palette.Border
	borderWidth := 1.0
	if node.IsFocused {
		borderCol = t.Palette.Primary
		borderWidth = 1.5
	}

	canvas.FillRoundedRect(fieldRect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(fieldRect, t.Radii.MD, borderCol, borderWidth)

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
		if node.IsFocused {
			canvas.DrawLine(geom.Pt(12, yOffset+12), geom.Pt(12, yOffset+28), t.Palette.Primary, 1.5)
		}
	}
}

// --- NumberInput ---

type NumberInputComponent struct {
	ui.BaseComponent
	Label    string
	Prefix   string
	Suffix   string
	Value    *state.Value[float64]
	Min      float64
	Max      float64
	Step     float64
	Width    float64
	OnChange func(float64)
}

// NumberInput creates a numeric stepper input.
func NumberInput(initial float64) *NumberInputComponent {
	return &NumberInputComponent{
		Value: state.Float(initial),
		Min:   -1000000,
		Max:   1000000,
		Step:  100.0,
		Width: 160,
	}
}

func (ni *NumberInputComponent) Bind(val *state.Value[float64]) *NumberInputComponent {
	ni.Value = val
	return ni
}

func (ni *NumberInputComponent) WithLabel(label string) *NumberInputComponent {
	ni.Label = label
	return ni
}

func (ni *NumberInputComponent) WithPrefix(prefix string) *NumberInputComponent {
	ni.Prefix = prefix
	return ni
}

func (ni *NumberInputComponent) WithSuffix(suffix string) *NumberInputComponent {
	ni.Suffix = suffix
	return ni
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

func (ni *NumberInputComponent) WithWidth(w float64) *NumberInputComponent {
	ni.Width = w
	return ni
}

func (ni *NumberInputComponent) OnChanged(fn func(float64)) *NumberInputComponent {
	ni.OnChange = fn
	return ni
}

func (ni *NumberInputComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := 34.0
	if ni.Label != "" {
		h += 18.0
	}

	w := ni.Width
	if constraints.HasBoundedWidth() && (ni.Width <= 0 || ni.Width > constraints.MaxWidth) {
		w = constraints.MaxWidth
	}

	node.OnPointerDown = func(e *event.PointerEvent) {
		if ni.Value == nil {
			return
		}
		cur := ni.Value.Get()
		yOffset := 0.0
		if ni.Label != "" {
			yOffset = 18.0
		}
		boxH := 34.0
		btnW := 24.0

		minusX := node.Bounds.Width - btnW*2
		plusX := node.Bounds.Width - btnW

		if e.Position.Y >= yOffset && e.Position.Y <= yOffset+boxH {
			if e.Position.X >= minusX && e.Position.X < plusX {
				step := ni.Step
				if step <= 0 {
					step = 1.0
				}
				newVal := cur - step
				if newVal < ni.Min {
					newVal = ni.Min
				}
				ni.Value.Set(newVal)
				if ni.OnChange != nil {
					ni.OnChange(newVal)
				}
			} else if e.Position.X >= plusX && e.Position.X <= node.Bounds.Width+8 {
				step := ni.Step
				if step <= 0 {
					step = 1.0
				}
				newVal := cur + step
				if newVal > ni.Max {
					newVal = ni.Max
				}
				ni.Value.Set(newVal)
				if ni.OnChange != nil {
					ni.OnChange(newVal)
				}
			}
		}
	}

	node.OnKeyDown = func(e *event.KeyEvent) {
		if ni.Value == nil {
			return
		}
		cur := ni.Value.Get()
		step := ni.Step
		if step <= 0 {
			step = 1.0
		}
		if e.Key == input.KeyArrowUp {
			newVal := cur + step
			if newVal > ni.Max {
				newVal = ni.Max
			}
			ni.Value.Set(newVal)
			if ni.OnChange != nil {
				ni.OnChange(newVal)
			}
		} else if e.Key == input.KeyArrowDown {
			newVal := cur - step
			if newVal < ni.Min {
				newVal = ni.Min
			}
			ni.Value.Set(newVal)
			if ni.OnChange != nil {
				ni.OnChange(newVal)
			}
		}
	}

	return constraints.Constrain(geom.Sz(w, h))
}

func (ni *NumberInputComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	yOffset := 0.0

	// Draw label if present
	if ni.Label != "" {
		canvas.DrawText(ni.Label, geom.Pt(0, 0), 12, font.WeightMedium, t.Palette.TextSecondary)
		yOffset += 18.0
	}

	boxH := 34.0
	fieldRect := geom.NewRect(0, yOffset, node.Bounds.Width, boxH)
	radius := geom.RadiusUniform(6)

	borderCol := t.Palette.Border
	if node.IsFocused {
		borderCol = t.Palette.BorderFocus
	} else if node.IsHovered {
		borderCol = t.Palette.BorderHover
	}

	canvas.FillRoundedRect(fieldRect, radius, t.Palette.Surface)
	canvas.StrokeRoundedRect(fieldRect, radius, borderCol, 1.0)

	val := 0.0
	if ni.Value != nil {
		val = ni.Value.Get()
	}

	displayStr := ""
	if ni.Prefix != "" {
		displayStr += ni.Prefix + " "
	}
	displayStr += fmt.Sprintf("%.2f", val)
	if ni.Suffix != "" {
		displayStr += " " + ni.Suffix
	}

	textY := yOffset + 9.0
	canvas.DrawText(displayStr, geom.Pt(10, textY), 13, font.WeightMedium, t.Palette.TextPrimary)

	// Step buttons
	btnW := 24.0
	divX := node.Bounds.Width - btnW*2
	canvas.DrawLine(geom.Pt(divX, yOffset+2), geom.Pt(divX, yOffset+boxH-2), t.Palette.Border, 1.0)
	canvas.DrawLine(geom.Pt(divX+btnW, yOffset+2), geom.Pt(divX+btnW, yOffset+boxH-2), t.Palette.Border, 1.0)

	// Minus button
	minusRect := geom.NewRect(divX+1, yOffset+1, btnW-2, boxH-2)
	canvas.FillRoundedRect(minusRect, geom.RadiusUniform(4), t.Palette.Secondary)
	canvas.DrawText("-", geom.Pt(minusRect.X+8, yOffset+8), 14, font.WeightBold, t.Palette.SecondaryText)

	// Plus button
	plusRect := geom.NewRect(divX+btnW+1, yOffset+1, btnW-2, boxH-2)
	canvas.FillRoundedRect(plusRect, geom.RadiusUniform(4), t.Palette.Secondary)
	canvas.DrawText("+", geom.Pt(plusRect.X+7, yOffset+8), 14, font.WeightBold, t.Palette.SecondaryText)
}
