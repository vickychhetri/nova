# Nova Developer Guide

Welcome to the **Nova Developer Guide**. This document provides an in-depth explanation of Nova's architecture, design patterns, extending the widget library, creating custom components, and developing high-performance cross-platform desktop applications in Go.

---

## Table of Contents

1. [Architectural Philosophy](#architectural-philosophy)
2. [Package Architecture](#package-architecture)
3. [Declarative UI & Component Lifecycle](#declarative-ui--component-lifecycle)
4. [Reactive State System](#reactive-state-system)
5. [Layout Engine Deep Dive](#layout-engine-deep-dive)
6. [2D Canvas & Render Commands](#2d-canvas--render-commands)
7. [Creating Custom Widgets](#creating-custom-widgets)
8. [Virtualization Engine](#virtualization-engine)
9. [Developer Tooling & CLI](#developer-tooling--cli)

---

## Architectural Philosophy

Nova is built on five core principles:
1. **Go-First**: Public APIs are designed idiomatically for Go developers. No HTML, CSS, JavaScript, or browser runtimes are required.
2. **Decoupled Layering**: UI declaration, reactive state, layout calculation, command buffer recording, and GPU rasterization are independent subsystems.
3. **Fine-Grained Reactivity**: Signal-based state graphs dirty-mark only the affected subtrees. A state change never re-evaluates or re-renders the whole application tree.
4. **Predictable Layout Engine**: Independent box-model and flex calculations without CSS quirks.
5. **Zero Friction Testing**: A high-performance pure-Go software rasterizer runs headless in CI/CD without GPU or X11/Wayland dependencies.

---

## Package Architecture

```
github.com/vickychhetri/nova
│
├── nova.go                          # Top-level application initialization & window options
├── core/
│   ├── geom/                        # Point, Size, Rect, Insets, Matrix2D, Radius
│   └── color/                       # Color, RGBA, Hex, HSL, alpha blending, contrast math
├── state/                           # Value[T], Computed[T], Effect, Batch
├── layout/                          # BoxConstraints, Flex, Stack, Grid, Box, Scroll
├── render/                          # CommandBuffer, Canvas, Display List, Draw Ops
├── renderer/
│   ├── renderer.go                  # Renderer & RenderTarget interfaces
│   └── software/                    # Pure-Go software rasterizer & PNG exporter
├── font/                            # Scalable typography & glyph rasterizer
├── text/                            # Multi-line measurement, word wrap, ellipsis
├── input/                           # Key codes, modifier flags, mouse buttons
├── event/                           # Event hierarchy, pointer, scroll, key events
├── theme/                           # Design tokens, Light & Dark themes
├── animation/                       # Timing curves (Spring, Bounce, Ease) & Tweens
├── ui/                              # Retained Node tree, Component interface & builders
├── widgets/                         # High-level components & forms
└── virtualization/                  # Viewport windowing for 1M+ items
```

---

## Declarative UI & Component Lifecycle

### `ui.Component` Interface

Every UI element implements the `Component` interface:

```go
type Component interface {
    Build(ctx BuildContext) Component
    Key() string
}
```

- **Compositional Components**: Implement `Build(ctx BuildContext) Component` by composing other primitives (e.g. `Card`, `Forms`, `Dialog`).
- **Renderable Leaf Components**: Implement `ui.RenderableComponent` to directly participate in layout measurement and canvas rendering:

```go
type RenderableComponent interface {
    Component
    Layout(node *Node, constraints layout.BoxConstraints) geom.Size
    Paint(node *Node, canvas *render.Canvas)
}
```

### Tree Reconciliation & Mounting

1. **Mounting**: When `window.Content(root)` is called, a root `ui.Node` is instantiated.
2. **Reconcile**: During tree reconciliation, child components are constructed or updated.
3. **Layout Pass**: `node.Layout(constraints)` cascades top-down, computing exact pixel dimensions and relative coordinates.
4. **Paint Pass**: `node.Paint(canvas)` records draw commands into an optimized `render.CommandBuffer`.
5. **Rasterization**: Commands are dispatched to the active backend renderer.

---

## Reactive State System

Nova uses a fine-grained signal graph.

### State Primitives

```go
import "github.com/vickychhetri/nova/state"

// Reactive state container
count := state.Int(0)
name := state.String("Alice")
active := state.Bool(true)
custom := state.New(MyStruct{ID: 1})

// Reading value (automatically tracks dependency inside computed/effect)
currentCount := count.Get()

// Mutating value (triggers subscribed listeners)
count.Set(5)
count.Update(func(c int) int { return c + 1 })
```

### Computed Signals

`state.Compute` creates a memoized reactive computation that automatically re-evaluates only when its referenced dependencies change:

```go
firstName := state.String("John")
lastName := state.String("Doe")

fullName := state.Compute(func() string {
    return firstName.Get() + " " + lastName.Get()
})
```

### Batching Updates

To prevent redundant UI frame triggers during multiple mutations:

```go
state.Batch(func() {
    firstName.Set("Jane")
    lastName.Set("Smith")
    count.Set(100)
})
```

---

## Layout Engine Deep Dive

Nova's layout engine operates deterministically with two passes:
1. **Constraints Down**: Parents pass `BoxConstraints` (min/max width and height) to children.
2. **Sizes Up**: Children return their measured `geom.Size`.

### Flex Container Math (`ui.Column`, `ui.Row`)

```go
ui.Row(
    ui.Text("Fixed Item"),                  // flex factor = 0
    ui.Expanded(ui.Text("Stretches")),      // flex factor = 1
    ui.Spacer(),                            // expands empty space
    widgets.Button("Action"),
).GapSpacing(12).AlignCross(layout.CrossCenter)
```

---

## Creating Custom Widgets

You can build custom widgets in two ways:

### 1. Composite Widget (High-Level)

```go
type UserProfileCard struct {
    ui.BaseComponent
    Username string
    Role     string
    Avatar   string
}

func (u *UserProfileCard) Build(ctx ui.BuildContext) ui.Component {
    return widgets.Card(u.Username,
        ui.Row(
            widgets.Avatar(u.Avatar),
            ui.Column(
                ui.Text(u.Username).Size(16).Weight(700),
                widgets.Badge(u.Role).Info(),
            ).GapSpacing(4),
        ).GapSpacing(12),
    )
}
```

### 2. Custom 2D Painted Widget (Low-Level)

```go
type CircularGauge struct {
    ui.BaseComponent
    Value float64 // 0.0 to 1.0
}

func (g *CircularGauge) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
    return constraints.Constrain(geom.Sz(100, 100))
}

func (g *CircularGauge) Paint(node *ui.Node, canvas *render.Canvas) {
    center := geom.Pt(50, 50)
    canvas.StrokeCircle(center, 40, theme.Current().Palette.Secondary, 8.0)
    canvas.StrokeCircle(center, 40, theme.Current().Palette.Primary, 8.0)
}
```

---

## Developer Tooling & CLI

```bash
# Scaffold new project
nova create myapp

# Run development mode with live reload
nova dev

# Build release binary with dead-code stripping
nova build

# Run unit tests and layout verification
nova test

# Run environment health check
nova doctor
```
