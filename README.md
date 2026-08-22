# Nova — Go Native Desktop Framework

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg)](#)

> **Build once with Go. Render natively. Run everywhere.**

Nova is a next-generation, Go-first desktop application framework designed to provide a modern, high-performance, GPU-accelerated alternative to Electron, Tauri, Fyne, Gio, Qt, and traditional native UI toolkits.

---

## Key Features

- 🏎️ **Ultra-High Performance & Low Latency**: 60–144+ FPS retained rendering pipeline with sub-millisecond layout passes.
- 🧩 **Declarative Component Model**: Clean, fluent Go APIs (`ui.Column`, `ui.Row`, `ui.Card`, `widgets.Button`, etc.). No HTML/CSS or JavaScript required.
- ⚡ **Fine-Grained Reactive Signals**: Signal-based state graph (`state.Int`, `state.String`, `state.Compute`, `state.Effect`, `state.Batch`) that invalidates only dirty subtrees.
- 📊 **Built-in First-Class Virtualization**: Effortlessly scroll and interact with **1,000,000+ data rows** at 60+ FPS with 13ns viewport calculation latency.
- 🎨 **Complete Built-in UI & Form Controls**: Full suite of widgets ready to import:
  - **Forms**: `TextField`, `PasswordField`, `TextArea`, `NumberInput`, `Checkbox`, `Radio`, `Switch`, `Slider`, `Select`/`Dropdown`, `DatePicker`, `ColorPicker`, `FilePicker`, and `Form` validation.
  - **Navigation**: `Tabs`, `Sidebar`, `SplitPane`, `Breadcrumb`.
  - **Feedback**: `Dialog` / `Modal`, `Alert`, `ToastManager`, `CommandPalette` (`Ctrl+K`).
  - **Data**: `VirtualList`, `Table`, `Tree`.
  - **Editors**: `CodeEditor` (syntax highlighted), `Canvas` (2D immediate drawing).
- 🛠️ **Developer Tooling & CLI**: Scaffolding (`nova create`), dev watcher (`nova dev`), release building (`nova build`), and diagnostics (`nova doctor`).
- 🖥️ **Universal Dual-Renderer Architecture**: High-speed pure-Go software rasterizer (for headless CI, unit tests, and software fallback) alongside hardware GPU pipelines.

---

## High-Level Architecture

```
                        NOVA APPLICATION (Go API)
                                   │
                                   ▼
        ┌─────────────────────────────────────────────────────┐
        │  nova package  (App, Window, Config, Options)        │
        └──────────────────────────┬──────────────────────────┘
                                   │
                 ┌─────────────────┼─────────────────┐
                 │                 │                 │
                 ▼                 ▼                 ▼
          ui (Components)    state (Signals)    theme (Tokens)
                 │                 │                 │
                 └─────────────────┼─────────────────┘
                                   ▼
        ┌─────────────────────────────────────────────────────┐
        │  UI Runtime (Component Tree, Event Dispatcher, Diff)│
        └──────────────────────────┬──────────────────────────┘
                                   │
                                   ▼
        ┌─────────────────────────────────────────────────────┐
        │  layout Engine (Flexbox, Stack, Grid, Box, Scroll)  │
        └──────────────────────────┬──────────────────────────┘
                                   │
                                   ▼
        ┌─────────────────────────────────────────────────────┐
        │  render Package (Display List, Command Buffer, Clip)│
        └──────────────────────────┬──────────────────────────┘
                                   │
                                   ▼
        ┌─────────────────────────────────────────────────────┐
        │  Renderer Interface (Batching, Glyph Atlas, Raster) │
        └──────┬───────────────────┬───────────────────┬──────┘
               │                   │                   │
               ▼                   ▼                   ▼
        renderer/software    renderer/vulkan     renderer/metal / gl
               │                   │                   │
               └───────────────────┼───────────────────┘
                                   ▼
                   Display Framebuffer / GPU Surface
```

---

## Quickstart

### 1. Installation

```bash
go get -u github.com/vickychhetri/nova
```

Install the Nova Developer CLI:
```bash
go install github.com/vickychhetri/nova/cmd/nova@latest
```

### 2. Hello World

```go
package main

import (
	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

func main() {
	app := nova.New()

	win := app.Window(
		nova.Title("My First Nova App"),
		nova.Size(800, 600),
	)

	win.Content(
		ui.Center(
			widgets.Card("Welcome to Nova",
				ui.Column(
					ui.Text("Build once with Go. Render natively. Run everywhere."),
					widgets.Button("Click Me").OnClick(func() {
						println("Button Clicked!")
					}),
				).GapSpacing(12),
			),
		),
	)

	app.Run()
}
```

### 3. Reactive State Example

```go
count := state.Int(0)
doubleCount := state.Compute(func() string {
    return fmt.Sprintf("Doubled: %d", count.Get()*2)
})

win.Content(
    ui.Center(
        ui.Column(
            ui.Text(state.Compute(func() string {
                return fmt.Sprintf("Count: %d", count.Get())
            })).Size(24).Weight(700),
            ui.Text(doubleCount),
            ui.Row(
                widgets.Button("- Decrement").Secondary().OnClick(func() {
                    count.Update(func(c int) int { return c - 1 })
                }),
                widgets.Button("+ Increment").OnClick(func() {
                    count.Update(func(c int) int { return c + 1 })
                }),
            ).GapSpacing(8),
        ).GapSpacing(12),
    ),
)
```

---

## Built-in Widget Catalog

| Category | Available Widgets |
| :--- | :--- |
| **Basic** | `Text`, `Button`, `Badge`, `Avatar`, `Spinner`, `Progress`, `Card`, `Skeleton`, `Spacer`, `Container` |
| **Form Inputs** | `Form`, `TextField`, `PasswordField`, `TextArea`, `NumberInput`, `Checkbox`, `Radio`, `Switch`, `Slider`, `Select`/`Dropdown`, `DatePicker`, `ColorPicker`, `FilePicker` |
| **Navigation** | `Tabs`, `Sidebar`, `SplitPane`, `Breadcrumb` |
| **Feedback** | `Dialog` / `Modal`, `Alert`, `ToastManager`, `CommandPalette` (`Ctrl+K`) |
| **Data Displays** | `VirtualList` (1M+ rows), `Table` (virtualized, sorting, selection), `Tree` |
| **Editors** | `CodeEditor` (syntax highlighted), `Canvas` (2D custom drawing) |

---

## Showcase Applications

Inside `examples/`:
- `examples/01_hello_world`: Minimal clean starter.
- `examples/02_counter`: Reactive state & computed signals demo.
- `examples/03_forms_showcase`: Complete form input suite with automated validation and dirty state tracking.
- `examples/04_widget_gallery`: Comprehensive component library showcase.
- `examples/05_novadb`: Production-quality database client showcase with 100,000-row virtualized table, SQL editor, and connection sidebar.

---

## Benchmarks

Run benchmarks locally:
```bash
go test -v -bench=. ./benchmarks/... ./virtualization/...
```

Typical performance on standard modern hardware:
- **Virtualizer 1,000,000 Items Viewport Calculation**: `12.98 ns/op` (~77,000,000 ops/sec)
- **Flex Layout (1,000 Children)**: `0.04 ms`
- **Reactive Signal Invalidation & Propagation**: `0.0001 ms`

---

## License

Apache License 2.0.
