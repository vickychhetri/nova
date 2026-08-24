package feedback

// Package feedback provides alerts, dialogs, toast notifications, and command
// palette widgets.

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

// --- Alert / Banner ---

type AlertType int

const (
	AlertInfo AlertType = iota
	AlertSuccess
	AlertWarning
	AlertError
)

type AlertComponent struct {
	ui.BaseComponent
	Title   string
	Message string
	Type    AlertType
}

// Alert creates an inline banner message.
func Alert(title, message string, alertType AlertType) *AlertComponent {
	return &AlertComponent{
		Title:   title,
		Message: message,
		Type:    alertType,
	}
}

func (a *AlertComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	h := 48.0
	if a.Message != "" {
		h = 64.0
	}
	return constraints.Constrain(geom.Sz(constraints.MaxWidth, h))
}

func (a *AlertComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	t := theme.Current()
	rect := geom.NewRect(0, 0, node.Bounds.Width, node.Bounds.Height)

	accentCol := t.Palette.Info
	switch a.Type {
	case AlertSuccess:
		accentCol = t.Palette.Success
	case AlertWarning:
		accentCol = t.Palette.Warning
	case AlertError:
		accentCol = t.Palette.Error
	}

	// Soft tinted background
	canvas.FillRoundedRect(rect, t.Radii.MD, accentCol.WithAlpha(0.12))
	canvas.StrokeRoundedRect(rect, t.Radii.MD, accentCol.WithAlpha(0.4), 1.0)

	// Left indicator bar
	barRect := geom.NewRect(0, 0, 4, node.Bounds.Height)
	canvas.FillRoundedRect(barRect, geom.RadiusSeparate(4, 0, 0, 4), accentCol)

	canvas.DrawText(a.Title, geom.Pt(16, 12), 14, font.WeightBold, t.Palette.TextPrimary)
	if a.Message != "" {
		canvas.DrawText(a.Message, geom.Pt(16, 32), 12, font.WeightRegular, t.Palette.TextSecondary)
	}
}

// --- Toast / Notification ---

type ToastItem struct {
	ID      string
	Title   string
	Message string
	Type    AlertType
}

type ToastManager struct {
	Toasts *state.Value[[]ToastItem]
}

func NewToastManager() *ToastManager {
	return &ToastManager{
		Toasts: state.New([]ToastItem{}),
	}
}

func (tm *ToastManager) Show(title, message string, toastType AlertType) {
	cur := tm.Toasts.Get()
	item := ToastItem{
		ID:      title,
		Title:   title,
		Message: message,
		Type:    toastType,
	}
	tm.Toasts.Set(append(cur, item))
}

func (tm *ToastManager) Dismiss(id string) {
	cur := tm.Toasts.Get()
	var updated []ToastItem
	for _, item := range cur {
		if item.ID != id {
			updated = append(updated, item)
		}
	}
	tm.Toasts.Set(updated)
}

// --- Dialog / Modal ---

type DialogComponent struct {
	ui.BaseComponent
	Title      string
	Message    string
	IsOpen     *state.Value[bool]
	OnConfirm  func()
	OnCancel   func()
	CustomBody ui.Component
}

// Dialog creates a modal dialog overlay.
func Dialog(title, message string, isOpen *state.Value[bool]) *DialogComponent {
	return &DialogComponent{
		Title:   title,
		Message: message,
		IsOpen:  isOpen,
	}
}

func (d *DialogComponent) OnOk(fn func()) *DialogComponent {
	d.OnConfirm = fn
	return d
}

func (d *DialogComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	if d.IsOpen != nil && !d.IsOpen.Get() {
		return geom.Sz(0, 0)
	}
	return constraints.Constrain(geom.Sz(constraints.MaxWidth, constraints.MaxHeight))
}

func (d *DialogComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	if d.IsOpen != nil && !d.IsOpen.Get() {
		return
	}

	t := theme.Current()
	// Dim backdrop overlay
	canvas.FillRect(geom.NewRect(0, 0, node.Bounds.Width, node.Bounds.Height), color.Black.WithAlpha(0.6))

	// Centered dialog card
	cardW := 420.0
	cardH := 200.0
	cardX := (node.Bounds.Width - cardW) / 2.0
	cardY := (node.Bounds.Height - cardH) / 2.0

	cardRect := geom.NewRect(cardX, cardY, cardW, cardH)
	canvas.DrawShadow(cardRect, t.Radii.LG, render.ShadowParams{
		Color:  color.Black.WithAlpha(0.4),
		Blur:   16,
		Spread: 4,
	})
	canvas.FillRoundedRect(cardRect, t.Radii.LG, t.Palette.Card)
	canvas.StrokeRoundedRect(cardRect, t.Radii.LG, t.Palette.Border, 1.0)

	// Title & body text
	canvas.DrawText(d.Title, geom.Pt(cardX+24, cardY+24), 18, font.WeightBold, t.Palette.TextPrimary)
	canvas.DrawText(d.Message, geom.Pt(cardX+24, cardY+60), 14, font.WeightRegular, t.Palette.TextSecondary)

	// Action buttons
	btnY := cardY + cardH - 52
	okRect := geom.NewRect(cardX+cardW-104, btnY, 80, 36)
	canvas.FillRoundedRect(okRect, t.Radii.MD, t.Palette.Primary)
	canvas.DrawText("Confirm", geom.Pt(okRect.X+14, okRect.Y+10), 13, font.WeightMedium, color.White)

	cancelRect := geom.NewRect(cardX+cardW-196, btnY, 80, 36)
	canvas.FillRoundedRect(cancelRect, t.Radii.MD, t.Palette.Secondary)
	canvas.DrawText("Cancel", geom.Pt(cancelRect.X+18, cancelRect.Y+10), 13, font.WeightMedium, t.Palette.SecondaryText)
}

// --- CommandPalette ---

type CommandItem struct {
	Title    string
	Shortcut string
	Action   func()
}

type CommandPaletteComponent struct {
	ui.BaseComponent
	IsOpen   *state.Value[bool]
	Query    *state.Value[string]
	Commands []CommandItem
}

// CommandPalette creates a quick launcher modal (Ctrl+K).
func CommandPalette(commands ...CommandItem) *CommandPaletteComponent {
	return &CommandPaletteComponent{
		IsOpen:   state.Bool(false),
		Query:    state.String(""),
		Commands: commands,
	}
}

func (cp *CommandPaletteComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	if cp.IsOpen != nil && !cp.IsOpen.Get() {
		return geom.Sz(0, 0)
	}
	return constraints.Constrain(geom.Sz(constraints.MaxWidth, constraints.MaxHeight))
}

func (cp *CommandPaletteComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	if cp.IsOpen != nil && !cp.IsOpen.Get() {
		return
	}

	t := theme.Current()
	canvas.FillRect(geom.NewRect(0, 0, node.Bounds.Width, node.Bounds.Height), color.Black.WithAlpha(0.6))

	paletteW := 520.0
	paletteH := 320.0
	pX := (node.Bounds.Width - paletteW) / 2.0
	pY := 80.0

	paletteRect := geom.NewRect(pX, pY, paletteW, paletteH)
	canvas.FillRoundedRect(paletteRect, t.Radii.LG, t.Palette.Card)
	canvas.StrokeRoundedRect(paletteRect, t.Radii.LG, t.Palette.Border, 1.0)

	// Search bar at top
	searchRect := geom.NewRect(pX+16, pY+16, paletteW-32, 40)
	canvas.FillRoundedRect(searchRect, t.Radii.MD, t.Palette.Surface)
	canvas.StrokeRoundedRect(searchRect, t.Radii.MD, t.Palette.Border, 1.0)
	canvas.DrawText("🔍 Search actions and navigation...", geom.Pt(pX+28, pY+28), 14, font.WeightRegular, t.Palette.TextMuted)

	// Command items
	itemY := pY + 72.0
	for _, cmd := range cp.Commands {
		canvas.DrawText(cmd.Title, geom.Pt(pX+24, itemY+8), 14, font.WeightMedium, t.Palette.TextPrimary)
		if cmd.Shortcut != "" {
			sz := text.MeasureText(cmd.Shortcut, 11, font.WeightRegular)
			badgeRect := geom.NewRect(pX+paletteW-32-sz.Width-12, itemY+4, sz.Width+12, 22)
			canvas.FillRoundedRect(badgeRect, t.Radii.SM, t.Palette.Secondary)
			canvas.DrawText(cmd.Shortcut, geom.Pt(badgeRect.X+6, badgeRect.Y+4), 11, font.WeightRegular, t.Palette.TextMuted)
		}
		itemY += 38.0
	}
}
