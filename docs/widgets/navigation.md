# Navigation Components

This guide covers top-level desktop application navigation widgets in Nova: **MenuBar**, **Dropdown Menus**, **Sidebar**, **Toolbar**, **StatusBar**, **Tabs**, and **SplitPane**.

---

## 1. MenuBar & Dropdown Menus (`widgets.MenuBar`)

### Summary & Purpose
`MenuBar` renders a native top desktop menu bar (Qt `QMenuBar` style). Clicking top menu headers opens a floating dropdown popup with menu items, keyboard shortcuts, separator dividers, and event click callbacks.

### Go Code Example
```go
package main

import (
	"fmt"
	"github.com/vickychhetri/nova/app"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

func main() {
	a := app.New()
	win := a.Window(ui.WithTitle("MenuBar Demo"), ui.WithSize(800, 500))

	statusMsg := state.String("Ready")

	menuBar := widgets.MenuBar(
		widgets.NewMenu("File",
			widgets.ShortcutItem("New Document", "Ctrl+N", func() {
				statusMsg.Set("Event: Created New Document")
			}),
			widgets.ShortcutItem("Open Project...", "Ctrl+O", func() {
				statusMsg.Set("Event: Open File Dialog Triggered")
			}),
			widgets.DividerItem(),
			widgets.ShortcutItem("Save", "Ctrl+S", func() {
				statusMsg.Set("Event: Saved Document")
			}),
			widgets.DividerItem(),
			widgets.ShortcutItem("Exit", "Ctrl+Q", func() {
				a.Quit()
			}),
		),
		widgets.NewMenu("Edit",
			widgets.ShortcutItem("Undo", "Ctrl+Z", func() { statusMsg.Set("Event: Undo") }),
			widgets.ShortcutItem("Redo", "Ctrl+Y", func() { statusMsg.Set("Event: Redo") }),
		),
		widgets.NewMenu("Help",
			widgets.ActionItem("About Application", func() {
				statusMsg.Set("Event: Opened About Dialog")
			}),
		),
	)

	win.Content(ui.Column(
		menuBar,
		ui.Padding(geom.All(20), ui.Text(statusMsg)),
	))

	_ = a.Run()
}
```

### Fluent Builders & Items
- `widgets.NewMenu(title string, items ...MenuItem)`: Creates a top menu group.
- `widgets.ActionItem(title string, onClick func())`: Clickable menu item.
- `widgets.ShortcutItem(title, shortcut string, onClick func())`: Item with shortcut hint (e.g. `Ctrl+S`).
- `widgets.DividerItem()`: Horizontal separator line.
- `widgets.SimpleMenu(title string, onClick func())`: Top menu bar button that fires directly without a dropdown.

### Under the Hood (How It Works Internally)
1. **Layout & Bounding Expansion**:
   - When closed, `MenuBar` occupies `Height: 28px`.
   - When a menu is clicked (`ActiveMenu >= 0`), `Layout` dynamically expands its node height to encompass the dropdown popup (`28 + len(items)*26 + 12`), ensuring child clicks are captured by `HitTestLocal`.
2. **Event Routing**:
   - `node.OnPointerDown` checks if click is in top bar (`Y <= 28`) to toggle menus, or in the open dropdown (`Y > 28`) to trigger the targeted `item.OnClick()`.
3. **Paint Pass**:
   - Records `CmdFillRect` for top bar and active header pills.
   - When open, records `CmdFillRoundedRect` and `CmdStrokeRoundedRect` with floating border and shortcut alignment.

---

## 2. Sidebar Navigation (`widgets.Sidebar`)

### Summary & Purpose
`Sidebar` provides a 230px fixed-width left rail navigation panel with selected accent indicators, badge counters, and section headers.

### Go Code Example
```go
sidebar := widgets.Sidebar("Admin Console",
    widgets.SidebarItem{Title: "Dashboard", Selected: true, OnClick: func() { ... }},
    widgets.SidebarItem{Title: "Users", Badge: "1,420", OnClick: func() { ... }},
    widgets.SidebarItem{Title: "Analytics", OnClick: func() { ... }},
    widgets.SidebarItem{Title: "Settings", OnClick: func() { ... }},
)
```

---

## 3. Action Toolbar (`widgets.Toolbar`)

### Summary & Purpose
`Toolbar` renders a 46px horizontal action container with 1px bottom border, hosting buttons, search fields, and status badges.

```go
toolbar := widgets.Toolbar(
    widgets.Button("Save").OnClick(func() { ... }),
    widgets.Button("Reload").Secondary().OnClick(func() { ... }),
    widgets.Badge("Production").Success(),
)
```

---

## 4. StatusBar (`widgets.StatusBar`)

### Summary & Purpose
`StatusBar` (Qt `QStatusBar` style) renders a 26px bottom status panel with live status indicator dot and right-aligned information segments.

```go
statusBar := widgets.StatusBar("System Ready — All systems nominal.",
    widgets.StatusSegment{Text: "Theme: Light Enterprise", Width: 160},
    widgets.StatusSegment{Text: "Latency: 12ms", Width: 100},
)
```
