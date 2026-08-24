# Chapter 5: Declarative UI & The Node Graph — Blueprint vs. Render Tree

> *"What is the difference between a declarative UI component and a stateful render node, and how does Nova construct and traverse the component hierarchy?"*

---

## 5.1 The Declarative UI Philosophy: $UI = f(\text{State})$

In traditional imperative GUI programming (e.g. Win32 GDI, raw GTK, or Java Swing), you create buttons and manually mutate their properties over time:

```c
// Imperative (Complex, error-prone, hard to synchronize state):
HWND btn = CreateWindow("BUTTON", "Submit", ...);
SetWindowText(btn, "Loading...");
EnableWindow(btn, FALSE);
```

In Nova, UI is **declarative**. Your UI function is a pure mathematical projection of your application state:

$$\text{User Interface} = f(\text{Current State})$$

```go
// Declarative (Nova):
func BuildOrderPanel(isSubmitting bool) ui.Component {
    btnText := "Submit Order"
    if isSubmitting {
        btnText = "Submitting..."
    }
    return ui.Row(
        widgets.Button(btnText).Primary().OnClick(func() { ... }),
        widgets.Badge("Ready").Success(),
    ).GapSpacing(12)
}
```

Whenever `isSubmitting` changes, Nova rebuilds the blueprint, solves the layout, and renders the updated state automatically!

---

## 5.2 `ui.Component` vs. `ui.Node`

Nova makes a strict architectural separation between **Blueprints (`ui.Component`)** and **Render Tree Nodes (`ui.Node`)**:

```
+-------------------------------------------------------------+
|                BLUEPRINT LAYER: ui.Component                |
|  • Ephemeral, lightweight, garbage-collected value structs |
|  • Examples: TextComponent, ButtonComponent, RowComponent  |
+-------------------------------------------------------------+
                              │
                    Mount() / Build Pass
                              ▼
+-------------------------------------------------------------+
|                 RENDER GRAPH LAYER: ui.Node                 |
|  • Persistent, stateful graph node                          |
|  • Stores computed layout: Bounds (X, Y, Width, Height)     |
|  • Stores interaction flags: IsHovered, IsFocused          |
|  • Stores parent & child graph pointers: Parent, Children[] |
|  • Stores event callback closures: OnClick, OnPointerDown   |
+-------------------------------------------------------------+
```

### The `ui.Component` Interface (`ui/component.go`):
```go
type Component interface {
    Mount(ctx BuildContext)
    Layout(node *Node, constraints layout.BoxConstraints) geom.Size
    Paint(node *Node, canvas *render.Canvas)
}
```

### The `ui.Node` Struct (`ui/node.go`):
```go
type Node struct {
    Component      Component
    Parent         *Node
    Children       []*Node
    Bounds         geom.Rect
    IsFocused      bool
    IsHovered      bool

    // Event hooks
    OnClick        func()
    OnPointerDown  func(*event.PointerEvent)
    OnPointerUp    func(*event.PointerEvent)
    OnPointerMove  func(*event.PointerEvent)
    OnPointerEnter func()
    OnPointerLeave func()
    OnKeyDown      func(*event.KeyEvent)
    OnScroll       func(*event.ScrollEvent)
}
```

---

## 5.3 Tree Traversal & The Paint Pass

When rendering a frame (`window/window.go`), Nova performs a depth-first traversal of the `ui.Node` tree:

```
                      Root Node (Window)
                              │
                    ┌─────────┴─────────┐
                    ▼                   ▼
              Header Bar            Main Content
                    │                   │
             ┌──────┴──────┐       ┌────┴────┐
             ▼             ▼       ▼         ▼
          Logo Text      Badges  Sidebar    Canvas
```

During the paint pass (`RenderFrame`):
1. **Clear Canvas**: The background is filled with the active theme background color (`#0B0F19`).
2. **Base UI Painting**: `w.rootNode.Paint(canvas)` is invoked recursively down the tree. Each node translates the canvas origin to its local `Bounds.Origin`, paints its borders/backgrounds/text, and paints its children.
3. **Overlay Layer Painting**: `w.rootNode.PaintOverlays(canvas)` paints dropdown menus, popups, and modal dialogs on top of all base UI content.

---

## 5.4 Floating Overlays & Portal Menus

In a standard UI tree, elements are painted in parent-child hierarchy order. However, if a Dropdown Menu or Select Box inside a deeply nested container opened inside its parent, it would be clipped by the parent's boundary or rendered beneath sibling cards!

Nova solves this with **Overlay Portals** (`ui/node.go`):

```go
// Register a floating overlay at the root level:
node.RegisterOverlay(menuNode)
```

1. **`PaintOverlays(canvas)`**: Traverses all registered overlay portals and renders them **after** the entire base UI tree is complete.
2. **`DispatchOverlayPointerDown(p)`**: When a mouse click occurs, Nova tests active overlays **first**. If an open dropdown intercepts the click, the event is consumed immediately, preventing clicks from leaking to underlying background buttons!

In Chapter 6, we explore the Constraint-based Layout Engine and Flexbox mathematics!
