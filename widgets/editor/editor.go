package editor

import (
	"fmt"
	"strings"

	"github.com/vickychhetri/nova/core/color"
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

// CodeEditorComponent is a high-performance code editor widget with line numbers and syntax highlighting.
type CodeEditorComponent struct {
	ui.BaseComponent
	Text      *state.Value[string]
	Language  string
	FontSize  float64
	OnChanged func(string)
}

// CodeEditor creates a code editor widget.
func CodeEditor(initialText, language string) *CodeEditorComponent {
	return &CodeEditorComponent{
		Text:     state.String(initialText),
		Language: language,
		FontSize: 13,
	}
}

func (ce *CodeEditorComponent) Bind(val *state.Value[string]) *CodeEditorComponent {
	ce.Text = val
	return ce
}

func (ce *CodeEditorComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := constraints.MaxHeight
	if h >= layout.Infinity {
		h = 300
	}
	w := constraints.MaxWidth
	if w >= layout.Infinity {
		w = 500
	}

	node.OnKeyDown = func(e *event.KeyEvent) {
		if ce.Text == nil {
			return
		}
		cur := ce.Text.Get()
		if e.Key == input.KeyBackspace {
			if len(cur) > 0 {
				newVal := cur[:len(cur)-1]
				ce.Text.Set(newVal)
				if ce.OnChanged != nil {
					ce.OnChanged(newVal)
				}
			}
		} else if e.Key == input.KeyEnter {
			newVal := cur + "\n"
			ce.Text.Set(newVal)
			if ce.OnChanged != nil {
				ce.OnChanged(newVal)
			}
		} else if e.Rune >= 32 && e.Rune != 127 {
			newVal := cur + string(e.Rune)
			ce.Text.Set(newVal)
			if ce.OnChanged != nil {
				ce.OnChanged(newVal)
			}
		}
	}

	return constraints.Constrain(geom.Sz(w, h))
}

func (ce *CodeEditorComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	w := node.Bounds.Width
	h := node.Bounds.Height

	// Editor background
	canvas.FillRoundedRect(geom.NewRect(0, 0, w, h), t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(geom.NewRect(0, 0, w, h), t.Radii.MD, t.Palette.Border, 1.0)

	gutterW := 48.0
	// Line number gutter background
	gutterRect := geom.NewRect(0, 0, gutterW, h)
	canvas.FillRoundedRect(gutterRect, geom.RadiusSeparate(8, 0, 0, 8), t.Palette.Secondary)
	canvas.DrawLine(geom.Pt(gutterW, 0), geom.Pt(gutterW, h), t.Palette.Border, 1.0)

	rawText := ""
	if ce.Text != nil {
		rawText = ce.Text.Get()
	}

	lines := strings.Split(rawText, "\n")
	lineH := 20.0
	curY := 8.0

	// SQL / Keyword highlighting tokens
	keywords := map[string]color.Color{
		"SELECT":  color.Hex("#F43F5E"),
		"FROM":    color.Hex("#F43F5E"),
		"WHERE":   color.Hex("#F43F5E"),
		"INSERT":  color.Hex("#F43F5E"),
		"UPDATE":  color.Hex("#F43F5E"),
		"DELETE":  color.Hex("#F43F5E"),
		"JOIN":    color.Hex("#F43F5E"),
		"AND":     color.Hex("#38BDF8"),
		"OR":      color.Hex("#38BDF8"),
		"LIMIT":   color.Hex("#F43F5E"),
		"ORDER":   color.Hex("#F43F5E"),
		"BY":      color.Hex("#F43F5E"),
		"func":    color.Hex("#F43F5E"),
		"package": color.Hex("#F43F5E"),
		"import":  color.Hex("#F43F5E"),
		"return":  color.Hex("#F43F5E"),
		"type":    color.Hex("#F43F5E"),
		"struct":  color.Hex("#F43F5E"),
	}

	for i, line := range lines {
		// Draw line number
		lineNumStr := fmt.Sprintf("%d", i+1)
		numSz := text.MeasureText(lineNumStr, 11, font.WeightRegular)
		canvas.DrawText(lineNumStr, geom.Pt(gutterW-12-numSz.Width, curY+3), 11, font.WeightRegular, t.Palette.TextMuted)

		// Draw line text with token highlighting
		words := strings.Split(line, " ")
		textX := gutterW + 12.0

		for wi, word := range words {
			col := t.Palette.TextPrimary
			cleaned := strings.Trim(strings.ToUpper(word), ";,()")
			if kwCol, ok := keywords[cleaned]; ok {
				col = kwCol
			} else if strings.HasPrefix(word, "\"") || strings.HasPrefix(word, "'") {
				col = color.Hex("#A3E635") // string literal green
			}

			canvas.DrawText(word, geom.Pt(textX, curY+3), ce.FontSize, font.WeightRegular, col)
			textX += text.MeasureText(word+" ", ce.FontSize, font.WeightRegular).Width

			if wi < len(words)-1 {
				// space already included in measurement
			}
		}

		curY += lineH
	}
}

// --- Interactive Canvas Widget ---

type CustomDrawFunc func(c *render.Canvas, bounds geom.Rect)

type CanvasWidgetComponent struct {
	ui.BaseComponent
	Width  float64
	Height float64
	OnDraw CustomDrawFunc
}

// Canvas creates a low-level 2D custom drawing component.
func Canvas(width, height float64, onDraw CustomDrawFunc) *CanvasWidgetComponent {
	return &CanvasWidgetComponent{
		Width:  width,
		Height: height,
		OnDraw: onDraw,
	}
}

func (cw *CanvasWidgetComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	return constraints.Constrain(geom.Sz(cw.Width, cw.Height))
}

func (cw *CanvasWidgetComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	if cw.OnDraw != nil {
		cw.OnDraw(canvas, node.Bounds)
	}
}
