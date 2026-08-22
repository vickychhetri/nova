package event

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/input"
)

// Event is the base interface for all Nova user interaction events.
type Event interface {
	IsHandled() bool
	StopPropagation()
}

type baseEvent struct {
	handled bool
}

func (e *baseEvent) IsHandled() bool {
	return e.handled
}

func (e *baseEvent) StopPropagation() {
	e.handled = true
}

// PointerEvent represents a mouse, pen, or touch interaction.
type PointerEvent struct {
	baseEvent
	Position geom.Point
	Button   input.MouseButton
	Mods     input.Modifiers
}

// ScrollEvent represents a mouse wheel or trackpad scroll gesture.
type ScrollEvent struct {
	baseEvent
	Position geom.Point
	DeltaX   float64
	DeltaY   float64
	Mods     input.Modifiers
}

// KeyEvent represents keyboard key presses and releases.
type KeyEvent struct {
	baseEvent
	Key  input.Key
	Rune rune
	Mods input.Modifiers
}

// FocusEvent represents focus gained or lost.
type FocusEvent struct {
	baseEvent
	Gained bool
}
