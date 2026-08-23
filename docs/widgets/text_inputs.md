# Text Inputs & Form Fields

This guide covers text inputs, password fields, multiline text areas, and numeric spinbox steppers in Nova.

---

## 1. TextField (`widgets.TextField`)

### Summary & Purpose
`TextField` provides single-line editable text input with two-way reactive state binding (`.Bind(val)`), top labels, error indicators, focus rings, and cursor rendering.

### Go Code Example
```go
nameState := state.String("Alexander Wright")

input := widgets.TextField("Enter employee name").
    WithLabel("Employee Full Name").
    WithWidth(300).
    Bind(nameState).
    OnChanged(func(newText string) {
        fmt.Printf("Updated name: %s\n", newText)
    })
```

### Fluent Builders
- `.WithLabel(label string)`: Adds a 12px medium font label above the input box.
- `.WithWidth(w float64)`: Sets explicit control width.
- `.WithError(err string)`: Renders red error outline and message beneath the input.
- `.Bind(val *state.Value[string])`: Binds directly to a reactive string signal.
- `.OnChanged(fn func(string))`: Event callback triggered on every keystroke.

### Under the Hood (How It Works Internally)
1. **Keyboard Event Handling (`node.OnKeyDown`)**:
   - When focused (`node.IsFocused == true`), incoming `event.KeyEvent` is routed to the field.
   - Handles `input.KeyBackspace` and ASCII codes (8, 127) by slicing `cur[:len(cur)-1]`.
   - Appends valid character runes (`e.Rune >= 32`).
   - Updates `tf.Value.Set(newVal)`, which invalidates dependent computed expressions.
2. **Paint Pass**:
   - Renders 6px rounded input box with `BorderFocus` (#2563EB) when focused.
   - Draws text and vertical cursor line at measured text width + 10px offset.

---

## 2. NumberInput (`widgets.NumberInput`)

### Summary & Purpose
`NumberInput` provides a numeric spinbox stepper control with reactive `float64` two-way binding, dollar/unit prefix and suffix, interactive `[-]` / `[+]` clickable stepper buttons, and keyboard arrow key stepping.

### Go Code Example
```go
salaryState := state.Float(8500.0)

numInput := widgets.NumberInput(8500.0).
    Bind(salaryState).
    WithLabel("Basic Salary ($)").
    WithPrefix("$").
    WithStep(500.0).
    WithMinMax(0.0, 100000.0).
    WithWidth(210.0).
    OnChanged(func(newVal float64) {
        fmt.Printf("Salary updated: $%.2f\n", newVal)
    })
```

### Stepper Interaction Under the Hood
1. **Local Coordinate Pointer Down (`node.OnPointerDown`)**:
   - `Window.DispatchPointerDown` translates window mouse coordinates into node-local coordinates via `HitTestLocal`.
   - Divides control into text display area and right stepper buttons (`divX = width - 48`).
   - Clicking between `[width - 48, width - 24)` decrements by `Step`.
   - Clicking between `[width - 24, width]` increments by `Step`.
2. **Keyboard Arrow Stepping (`node.OnKeyDown`)**:
   - Pressing <kbd>↑</kbd> increments by `Step`.
   - Pressing <kbd>↓</kbd> decrements by `Step`.

---

## 3. PasswordField (`widgets.PasswordField`) & TextArea (`widgets.TextArea`)

- `PasswordField`: Masked text input rendering dot bullets (`•`) with show/hide toggle support.
- `TextArea`: Multi-line text box with `.WithRows(n)` and <kbd>Enter</kbd> newline support.
