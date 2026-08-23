# Selection Controls

This guide covers toggle and choice input widgets in Nova: **Checkbox**, **Radio**, **Switch**, **Slider**, **Select**, **DatePicker**, and **FilePicker**.

---

## 1. Checkbox (`widgets.Checkbox`)

### Summary & Purpose
`Checkbox` renders a square selection box with checkmark graphic and adjacent label, supporting two-way boolean reactive binding.

### Go Code Example
```go
agreeTerms := state.Bool(false)

cb := widgets.Checkbox("I agree to terms and conditions").
    Bind(agreeTerms).
    OnChanged(func(checked bool) {
        fmt.Printf("Terms agreed: %v\n", checked)
    })
```

---

## 2. Switch Toggle (`widgets.Switch`)

### Summary & Purpose
`Switch` renders a modern iOS/Fluent style sliding toggle pill for binary settings.

```go
notifications := state.Bool(true)

sw := widgets.Switch("Enable Email Alerts").
    Bind(notifications)
```

---

## 3. Slider (`widgets.Slider`)

### Summary & Purpose
`Slider` provides a continuous range drag control with minimum, maximum, and reactive numeric binding.

```go
volume := state.Float(75.0)

slider := widgets.Slider(0, 100).
    Bind(volume).
    WithLabel("Master Volume").
    WithStep(5.0)
```

---

## 4. Select Dropdown (`widgets.Select`)

### Summary & Purpose
`Select` provides a dropdown menu for selecting from a list of options.

```go
role := state.String("developer")

dropdown := widgets.Select(
    widgets.SelectOption{Label: "Software Engineer", Value: "developer"},
    widgets.SelectOption{Label: "Product Manager", Value: "pm"},
    widgets.SelectOption{Label: "System Architect", Value: "architect"},
).Bind(role)
```
