# Nova Architecture & Performance Guide

This document provides a comprehensive technical breakdown of the **Nova UI Framework** architecture, event processing pipeline, rendering engine, and performance best practices.

---

## 1. High-Level Architecture Overview

Nova is structured as a decoupled, multi-layer GUI framework built in pure Go with native OS display bindings:

```
┌──────────────────────────────────────────────────────────────────┐
│                   Application Layer (Go Code)                    │
│     Examples / User Apps (Windows, Layouts, Forms, Data Views)   │
└──────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌──────────────────────────────────────────────────────────────────┐
│              Declarative Component & Widget Layer                │
│    widgets/ (Buttons, TextFields, Cards, GroupBoxes, Tables)     │
│    ui/ (Flex, Row, Column, Stack, Padding, Container, Builders)  │
└──────────────────────────────────────────────────────────────────┘
        │                                          ▲
        ▼                                          │ (triggers)
┌───────────────────────────┐          ┌───────────────────────────┐
│     Retained UI Tree      │          │   Reactive State Graph    │
│    ui.Node / Hit Testing  │◄─────────│   state.Value / Computed  │
│    Focus & Hover Dispatch │          │   Signals & Dependencies  │
└───────────────────────────┘          └───────────────────────────┘
        │
        ▼
┌──────────────────────────────────────────────────────────────────┐
│                      Layout & Geometry Engine                    │
│     layout/ (BoxConstraints, ComputeFlex, Alignment, Grid)       │
│     core/geom/ (Point, Rect, Size, Insets, CornerRadius)         │
└──────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌──────────────────────────────────────────────────────────────────┐
│                    2D Canvas & Display List                      │
│     render/ (Canvas, CommandBuffer, Display List Commands)       │
│     core/color/ (Color math, Palettes, Theme tokens)             │
└──────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌──────────────────────────────────────────────────────────────────┐
│                      Rasterization Engine                        │
│     renderer/software/ (Pure-Go Software Rasterizer)             │
│     font/ & text/ (OpenType SFNT Glyph Rasterizer & Text Wrap)   │
└──────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌──────────────────────────────────────────────────────────────────┐
│                    Platform & OS Window Backend                  │
│     platform/linux/ (X11 Native Window, 32-bit Framebuffer Blit) │
│     event/ & input/ (Event Routing, KeySyms, Focus & Pointers)   │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. Core Subsystems

### A. The Declarative UI & Retained Node Tree (`ui/`)
- **Components (`ui.Component`)**: Lightweight immutable descriptors defining how a widget should lay out and paint.
- **Nodes (`ui.Node`)**: The retained tree nodes that persist across frames. Nodes store:
  - `Bounds`: Computed layout rectangle in parent coordinates.
  - `IsHovered` / `IsFocused`: Active interaction states.
  - Event handlers (`OnClick`, `OnPointerDown`, `OnPointerUp`, `OnPointerMove`, `OnKeyDown`).
  - Hierarchical parent-child links for recursive hit-testing.

### B. Reactive State System (`state/`)
- **`state.Value[T]`**: Reactive state signal. Calling `.Get()` inside a computed context registers an automatic dependency; calling `.Set(val)` notifies listeners and triggers invalidation.
- **`state.Compute(fn)`**: Memoized derived state that automatically re-evaluates only when upstream dependencies change.

### C. Layout Engine (`layout/`)
- Pure, predictable box-model calculation using **BoxConstraints** (min/max width and height).
- Flex calculations (`ComputeFlex`) evaluate fixed-size elements first, then distribute remaining space among expanded/flex factors (`ui.Expanded`), supporting horizontal and vertical axes, cross-axis alignment, and gaps.

### D. Rendering & Rasterization (`render/` & `renderer/`)
1. **Canvas Command Buffer**: During the paint pass, widgets record high-level drawing operations (`CmdFillRoundedRect`, `CmdStrokeRoundedRect`, `CmdText`, `CmdLine`, `CmdPushClip`, `CmdPopClip`).
2. **Software Rasterizer (`software.Rasterizer`)**: Traverses the command list and rasters antialiased geometry and font glyphs into an `*image.RGBA` framebuffer.
3. **Hardware Blit**: Pushes the frame buffer directly to the OS surface (X11 `XPutImage` via persistent 32-bit BGRA buffer).

---

## 3. Event Loop & Performance Pipeline

```
 OS Event (Mouse/Key)
         │
         ▼
 X11 PollEvents() (Drains pending event queue)
         │
         ├── PointerDown ──► Update Focus (defocus previous) & Dispatch OnClick/OnPointerDown
         ├── PointerMove ──► Update Hover State (marks dirty ONLY if hover changed)
         └── KeyPress    ──► Maps KeySym to input.Key (Backspace, Enter, Arrows) & Rune
         │
         ▼
 Window.NeedsRedraw() == true ?
         │
    [ YES ] ──► 1. Layout Pass (ComputeFlex & Bounds)
                2. Paint Pass (Record Draw Commands)
                3. Rasterize Frame (Software RGBA)
                4. Blit to Native Surface (Optimized 32-bit XPutImage)
```

### Why Nova is Now Ultra-Fast:
1. **Event Draining Before Render**: All queued pointer motion events are drained in memory before rendering. Redraws only execute if hover or interaction state actually changed.
2. **Persistent Framebuffer**: Reuses a persistent 32-bit pixel buffer in C/X11 rather than reallocating megabytes of memory per frame.
3. **Focus Isolation**: Only the active focused node receives key input, and focus cleanly toggles when clicking different controls.
4. **Reliable Key Mapping**: KeySyms (e.g. `XK_BackSpace`, `XK_Return`, `XK_Tab`) are mapped to `input.Key` constants and runes, guaranteeing backspace and keyboard input always work.

---

## 4. Best Practices to Work Fast in Nova

### 1. Rapid Development Workflow
- **Run Example Apps Directly**:
  ```bash
  go run ./examples/07_payslip_generator/main.go
  ```
- **Instant Headless Frame Verification**:
  You can render and verify any UI layout without opening a window by saving a screenshot at the end of the script:
  ```go
  win.SaveScreenshot("preview.png")
  ```
- **Automated Testing**:
  Nova includes comprehensive unit tests across all layers. Run them anytime:
  ```bash
  go test ./...
  ```

### 2. Layout Tips
- **Use `ui.Row` and `ui.Column` with `ui.Expanded`**: Wrap side-by-side or stacked containers in `ui.Expanded(...)` to distribute space cleanly.
- **Set Explicit Widths for Fixed Columns**: E.g. `widgets.Sidebar` with 230px fixed width alongside `ui.Expanded(content)`.
- **Use Two-Way Reactive Binding**: Bind inputs directly to `state.Value[T]` using `.Bind(val)` so UI controls update application state automatically.
