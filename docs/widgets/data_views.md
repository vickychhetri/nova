# Data Views & Virtualization

This guide covers data presentation components in Nova: **Table**, **List**, **Tree**, and **VirtualList** (for rendering 1,000,000+ items at 120 FPS).

---

## 1. Table (`widgets.Table`)

### Summary & Purpose
`Table` renders a structured spreadsheet/grid view with sortable headers, column widths, alternating zebra row striping, and cell formatters.

### Go Code Example
```go
type Employee struct {
    ID   string
    Name string
    Role string
    Pay  string
}

employees := []Employee{
    {"EMP-918", "Dr. Sarah Jenkins", "Chief Data Scientist", "$16,500.00"},
    {"EMP-804", "Alexander Wright", "Principal Architect", "$13,700.00"},
}

table := widgets.Table(
    []widgets.TableColumn{
        {Title: "Employee ID", Width: 120},
        {Title: "Full Name", Width: 200},
        {Title: "Designation", Width: 220},
        {Title: "Gross Monthly", Width: 140},
    },
    len(employees),
    func(row int, col int) string {
        e := employees[row]
        switch col {
        case 0: return e.ID
        case 1: return e.Name
        case 2: return e.Role
        case 3: return e.Pay
        default: return ""
        }
    },
)
```

---

## 2. VirtualList & 1M+ Row Virtualization (`virtualization.VirtualList`)

### Summary & Purpose
`VirtualList` uses viewport windowing to calculate which rows are currently visible on screen. Even with 1,000,000 items in memory, Nova only instantiates and paints ~25 visible rows, maintaining 0ms scroll latency and minimal memory footprint.

```go
vlist := virtualization.NewVirtualList(1000000, 32.0, func(index int) ui.Component {
    return ui.Row(
        ui.Text(fmt.Sprintf("Item #%d", index)),
    )
})
```
