# Containers & Cards

This guide covers structured grouping panels in Nova: **Card**, **GroupBox**, **Alert**, **Badge**, **ProgressBar**, and **Spinner**.

---

## 1. GroupBox (`widgets.GroupBox`)

### Summary & Purpose
`GroupBox` (Qt `QGroupBox` style) renders a framed section container with a 32px tinted header band, 1px perimeter border, and internal padding.

### Go Code Example
```go
empGroupBox := widgets.GroupBox("1. Employee Position Information",
    ui.Column(
        ui.Row(
            widgets.TextField("Enter Name").WithLabel("Full Name").WithWidth(300),
            widgets.TextField("EMP-100").WithLabel("Employee ID").WithWidth(300),
        ).GapSpacing(14),
    ).GapSpacing(10),
)
```

### Under the Hood
- **Deflated Layout**: Insets child constraints by `headerHeight (32) + padding (14)`.
- **Canvas Clipping & Frame**: Draws a surface-hover tinted header bar, bottom divider line, and delegates child painting without double-translating coordinates.

---

## 2. Card (`widgets.Card`)

### Summary & Purpose
`Card` renders an elevated surface tile with 8px radius, 1px border, title header, subtitle, and body children. When wrapped in `ui.Expanded(widgets.Card(...))`, cards distribute evenly across a row.

```go
grossCard := ui.Expanded(widgets.Card("GROSS MONTHLY EARNINGS",
    ui.Column(
        ui.Text("$13,700.00").Size(20).Weight(700),
        widgets.Badge("4 CTC Allowances").Info(),
    ).GapSpacing(4),
))
```

---

## 3. Alert (`widgets.Alert`)

### Summary & Purpose
`Alert` renders a notification banner with soft background tint, left colored accent border, title, and descriptive message.

```go
alert := widgets.Alert(
    "Payroll Audit Passed",
    "All statutory deductions comply with 2026 IRS matrices.",
    widgets.AlertSuccess, // AlertInfo, AlertWarning, AlertError
)
```

---

## 4. Badge & Chip (`widgets.Badge`)

### Summary & Purpose
`Badge` renders compact status chips with rounded pill borders and semantic color variants.

```go
badge1 := widgets.Badge("Active Direct Deposit").Success()
badge2 := widgets.Badge("Tax Review Required").Warning()
badge3 := widgets.Badge("ACME Global Corp").Info()
```
