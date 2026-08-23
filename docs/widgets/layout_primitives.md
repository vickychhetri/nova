# Layout Primitives

This guide covers layout building blocks in Nova: **Row**, **Column**, **Flex**, **Stack**, **Grid**, **Padding**, **Container**, and **Expanded**.

---

## 1. Row & Column (`ui.Row`, `ui.Column`)

### Summary & Purpose
`Row` and `Column` organize children along horizontal and vertical axes with gap spacing, cross-axis alignment, and flex factor distribution.

### Go Code Example
```go
layout := ui.Column(
    topHeader,
    ui.Row(
        sidebar,                     // fixed 230px
        ui.Expanded(rightContent),   // expands to fill remaining space
    ),
    bottomFooter,
).GapSpacing(10)
```

---

## 2. Expanded (`ui.Expanded`)

### Summary & Purpose
`ui.Expanded(c)` assigns a flex factor of `1.0` to a component within a `Row` or `Column`. If multiple siblings are wrapped in `ui.Expanded(...)`, available space is split evenly among them.

```go
// Three cards spanning exactly 33.3% width each
cardsRow := ui.Row(
    ui.Expanded(widgets.Card("Earnings", earningsContent)),
    ui.Expanded(widgets.Card("Deductions", deductionsContent)),
    ui.Expanded(widgets.Card("Net Pay", netPayContent)),
).GapSpacing(14)
```

---

## 3. Padding & Container (`ui.Padding`, `ui.Container`)

- `ui.Padding(geom.All(14), child)`: Insets a child component by uniform or custom TRBL margins.
- `ui.Container(child)`: Sets explicit width, height, background color, and corner radius.
