# Reactive State Management

This guide covers fine-grained reactivity in Nova: **Value[T]**, **Compute**, **Effect**, and **Batch**.

---

## 1. Reactive Signals (`state.Value[T]`)

### Summary & Purpose
`state.Value[T]` is a thread-safe reactive state container. Calling `.Get()` inside a computed context records an automatic dependency; calling `.Set(newVal)` notifies subscribers and marks UI nodes dirty for redraw.

### Go Code Example
```go
// Create reactive state variables
basicSalary := state.Float(8500.0)
hraAllowance := state.Float(2400.0)

// Two-way binding with widgets
input := widgets.NumberInput(8500.0).Bind(basicSalary)

// Update state anywhere
basicSalary.Set(9000.0)
```

---

## 2. Derived Computed Signals (`state.Compute`)

### Summary & Purpose
`state.Compute(fn)` creates a reactive computed value that re-evaluates automatically whenever any of its dependencies change.

```go
grossEarnings := state.Compute(func() float64 {
    return basicSalary.Get() + hraAllowance.Get()
})

// Bind computed string directly to a Text widget
summaryText := ui.Text(state.Compute(func() string {
    return fmt.Sprintf("Total Gross: $%.2f", grossEarnings.Get())
}))
```

---

## 3. Side Effects & Batching (`state.Effect`, `state.Batch`)

- **`state.Effect(fn)`**: Runs whenever observed signals change (e.g. auto-saving to SQLite/JSON on state changes).
- **`state.Batch(fn)`**: Groups multiple `.Set()` calls into a single atomic notification, preventing redundant recalculations.
