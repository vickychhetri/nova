package event

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/input"
)

// Event is the base interface implemented by all Nova user-interaction events.
//
// Events are mutable during dispatch: a handler can call StopPropagation when
// it has consumed the event. The dispatcher can then check IsHandled and avoid
// delivering the same event to later handlers or ancestors.
type Event interface {
	// IsHandled reports whether a handler has consumed this event.
	IsHandled() bool
	// StopPropagation marks the event as handled and stops further delivery.
	StopPropagation()
}

// baseEvent contains the propagation state shared by every concrete event.
// It is embedded in event structs so its methods are promoted to the public
// event type without exposing the internal handled field directly.
type baseEvent struct {
	// handled is true after StopPropagation has been called.
	handled bool
}

// IsHandled reports whether this event has been marked as consumed.
func (e *baseEvent) IsHandled() bool {
	return e.handled
}

// StopPropagation marks this event as handled. Calling it more than once is
// harmless; the state remains true.
func (e *baseEvent) StopPropagation() {
	e.handled = true
}

// PointerEvent represents a mouse, pen, or touch interaction.
//
// Position is expressed in the coordinate space used by the event dispatcher.
// Button identifies the affected pointer button, and Mods records keyboard
// modifiers held during the interaction.
type PointerEvent struct {
	baseEvent
	// Position is the pointer location associated with the event.
	Position geom.Point
	// Button identifies the mouse or pointer button involved.
	Button input.MouseButton
	// Mods contains modifier keys active when the event occurred.
	Mods input.Modifiers
}

// ScrollEvent represents a mouse-wheel or trackpad scroll gesture.
//
// DeltaX and DeltaY describe the scroll amount. Their units and sign follow
// the input backend that produced the event; consumers should avoid assuming
// that every device reports identical step sizes.
type ScrollEvent struct {
	baseEvent
	// Position is the pointer location where the scroll occurred.
	Position geom.Point
	// DeltaX is horizontal scroll movement.
	DeltaX float64
	// DeltaY is vertical scroll movement.
	DeltaY float64
	// Mods contains modifier keys active during the gesture.
	Mods input.Modifiers
}

// KeyEvent represents a keyboard key press or release.
//
// Key identifies the normalized Nova key. Rune contains the associated
// character when the backend can provide one, and may be zero for non-printing
// keys such as arrows or function keys. Mods records active modifiers.
type KeyEvent struct {
	baseEvent
	// Key is Nova's normalized key identifier.
	Key input.Key
	// Rune is the Unicode character associated with the key, when available.
	Rune rune
	// Mods contains modifier keys active when the key event occurred.
	Mods input.Modifiers
}

// FocusEvent represents focus gained or lost.
//
// Gained is true when focus was received and false when focus was lost.
type FocusEvent struct {
	baseEvent
	// Gained indicates whether the component or window received focus.
	Gained bool
}
