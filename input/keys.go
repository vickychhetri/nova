package input

// Key represents a normalized physical or virtual keyboard key.
//
// Platform backends translate native key codes into these values so the rest
// of Nova does not need to depend on X11, Wayland, or another window system.
type Key int

const (
	// KeyUnknown represents a key that could not be normalized.
	KeyUnknown Key = iota
	// KeyEscape represents the Escape key.
	KeyEscape
	// KeyEnter represents the Return or Enter key.
	KeyEnter
	// KeyTab represents the Tab key.
	KeyTab
	// KeyBackspace represents the Backspace key.
	KeyBackspace
	// KeySpace represents the Space key.
	KeySpace
	// KeyDelete represents the Delete key.
	KeyDelete
	// KeyArrowUp represents the Up Arrow key.
	KeyArrowUp
	// KeyArrowDown represents the Down Arrow key.
	KeyArrowDown
	// KeyArrowLeft represents the Left Arrow key.
	KeyArrowLeft
	// KeyArrowRight represents the Right Arrow key.
	KeyArrowRight
	// KeyHome represents the Home key.
	KeyHome
	// KeyEnd represents the End key.
	KeyEnd
	// KeyPageUp represents the Page Up key.
	KeyPageUp
	// KeyPageDown represents the Page Down key.
	KeyPageDown
	// KeyA through KeyZ represent alphabetic keys.
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	// Key0 through Key9 represent numeric digit keys.
	Key0
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
	// KeyF1 through KeyF12 represent function keys.
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// Modifiers represents a set of active keyboard modifier flags.
//
// Values are powers of two so several modifiers can be combined with bitwise
// OR, for example ModCtrl|ModShift. Use Has to test whether a flag is active.
type Modifiers int

const (
	// ModShift indicates that Shift is active.
	ModShift Modifiers = 1 << iota
	// ModCtrl indicates that Control is active.
	ModCtrl
	// ModAlt indicates that Alt is active.
	ModAlt
	// ModMeta indicates that Meta, Command, or Super is active.
	ModMeta
)

// Has reports whether any bit in flag is active in m.
//
// Passing a combined value therefore succeeds when at least one requested
// modifier is active. Passing zero returns false because no active bit can
// match the zero flag.
func (m Modifiers) Has(flag Modifiers) bool {
	return (m & flag) != 0
}

// MouseButton represents a normalized mouse or pointer button.
//
// Native button numbers are converted into these values by platform backends.
// ButtonNone represents the absence of a button or an unsupported button.
type MouseButton int

const (
	// ButtonNone indicates that no pointer button is involved.
	ButtonNone MouseButton = iota
	// ButtonLeft represents the primary or left pointer button.
	ButtonLeft
	// ButtonRight represents the secondary or right pointer button.
	ButtonRight
	// ButtonMiddle represents the middle pointer button.
	ButtonMiddle
)
