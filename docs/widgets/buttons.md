# Buttons & Action Controls

This guide covers interactive button components in Nova: **Button**, **IconButton**, **ButtonGroup**, and styling modifiers.

---

## 1. Button (`widgets.Button` & `ui.Button`)

### Summary & Purpose
`Button` renders interactive desktop push buttons with 1px border, smooth 6px rounded corners, hover highlight effects, and click callbacks.

### Go Code Example
```go
// Primary Action Button
btn1 := widgets.Button("Submit Application").
    OnClick(func() {
        fmt.Println("Application submitted!")
    })

// Secondary / Outlined Button
btn2 := widgets.Button("Cancel").
    Secondary().
    OnClick(func() {
        fmt.Println("Cancelled")
    })

// Danger / Destructive Button
btn3 := widgets.Button("Delete Record").
    Danger().
    OnClick(func() {
        fmt.Println("Record deleted!")
    })
```

### Fluent Builders
- `.Secondary()`: Applies white surface fill with 1px border.
- `.Danger()`: Applies error red color scheme for destructive actions.
- `.WithWidth(w float64)`: Sets fixed button width.
- `.WithHeight(h float64)`: Sets custom button height (default: 34px).
- `.WithIcon(icon string)`: Sets leading icon label.
- `.Disabled(bool)`: Disables pointer interaction and greys out appearance.
- `.OnClick(fn func())`: Registers click event callback.

### Under the Hood (How It Works Internally)
1. **Layout Pass (`ButtonComponent.Layout`)**:
   - Measures label text using `text.MeasureText(btn.Label, 13, font.WeightMedium)`.
   - Computes minimum width = `textWidth + paddingX*2` (default 34px height).
   - Constrains to parent `BoxConstraints`.
2. **Node Interactive State**:
   - `node.OnClick = btn.OnClick`.
   - `node.IsHovered`: Automatically set during pointer motion inside `Window.DispatchPointerMove`.
3. **Paint Pass (`ButtonComponent.Paint`)**:
   - Checks `node.IsHovered` to select fill color (`t.Palette.PrimaryHover` vs `t.Palette.Primary`).
   - Draws rounded rect with `canvas.FillRoundedRect` and `canvas.StrokeRoundedRect`.
   - Centers and draws text using `canvas.DrawText`.
