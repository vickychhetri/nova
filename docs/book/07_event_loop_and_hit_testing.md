# Chapter 7: Event Loop, Routing & Hit Testing — Input Dispatch & Concurrency

> *"When a user clicks a button or presses a key on their keyboard, how does Nova traverse the visual tree, locate the exact hit target, and invoke the correct Go closure without race conditions or deadlocks?"*

---

## 7.1 The Life of a Mouse Click: From Hardware to Action

```
[Hardware Mouse Click] ──> [OS Kernel Driver] ──> [X11 / Display Server]
                                                            │
                                                   XNextEvent() FFI
                                                            ▼
                                              [Window.DispatchPointerDown(P)]
                                                            │
                                                   Recursive Hit Testing
                                                            ▼
                                              [Target Node: ButtonComponent]
                                                            │
                                                   Release Mutex Lock
                                                            ▼
                                              [Execute User OnClick() Closure]
```

When you click at coordinates $(450, 320)$:
1. The OS window manager notifies Nova with a native `ButtonPress` event.
2. Nova routes coordinates $(X, Y)$ to `Window.DispatchPointerDown`.
3. Nova performs **Recursive Spatial Hit Testing** to locate the innermost component at that exact point.
4. Nova updates focus states and invokes the button's `OnClick()` handler.

---

## 7.2 Recursive Spatial Hit Testing (`HitTestLocal`)

Every `ui.Node` in the render tree has computed geometric bounds:

$$\text{Bounds} = [X, Y, \text{Width}, \text{Height}]$$

To find which node was clicked (`ui/node.go`):

```go
func (n *Node) HitTestLocal(p geom.Point) (*Node, geom.Point) {
    // 1. Point-in-Rectangle check
    if !n.Bounds.Contains(p) {
        return nil, p
    }

    // 2. Transform point into local child coordinate space
    localP := geom.Pt(p.X - n.Bounds.X, p.Y - n.Bounds.Y)

    // 3. Test children in REVERSE order (topmost z-index child tested first!)
    for i := len(n.Children) - 1; i >= 0; i-- {
        child := n.Children[i]
        if hit, childLocal := child.HitTestLocal(localP); hit != nil {
            return hit, childLocal
        }
    }

    // 4. If no children matched, this node itself was clicked
    return n, localP
}
```

---

## 7.3 Hover States & Focus Management

Nova automatically tracks pointer movement without generating redundant rendering frames:

```
[Pointer Moves] ──> HitTestLocal(P) ──> New Hovered Node?
                                               │
                      ┌────────────────────────┴────────────────────────┐
                      ▼                                                 ▼
                 YES (Changed)                                      NO (Same)
                      │                                                 │
      1. Previous Node.IsHovered = false                          No Redraw Needed!
         Invoke Previous.OnPointerLeave()                         (Zero CPU cycles)
      2. New Node.IsHovered = true
         Invoke New.OnPointerEnter()
      3. Set needsRedraw = true
```

When clicking an input field:
- The previously focused node receives `IsFocused = false`.
- The new node receives `IsFocused = true`.
- Key events are automatically routed to the focused node.

---

## 7.4 Non-Blocking Mutex Boundaries (Eliminating Deadlocks)

A critical architectural challenge in multithreaded GUI systems is **Mutex Re-entrancy**.

In Go, `sync.RWMutex` is **non-reentrant** (a goroutine cannot acquire a lock it already holds without deadlocking).

### The Bug (Deadlock):
```go
func (w *Window) DispatchPointerDown(p geom.Point) {
    w.mu.Lock() // <-- Lock Window
    hit, _ := w.rootNode.HitTestLocal(p)
    if hit.OnClick != nil {
        hit.OnClick() // User callback calls win.Content() or rebuildView()!
                      // win.Content() tries to acquire w.mu.Lock() on the SAME thread!
                      // ===> INSTANT DEADLOCK / WINDOW FREEZES! <===
    }
    w.mu.Unlock()
}
```

### The Solution: Clean Mutex Boundary
Extract all callback closures to local variables, **unlock the mutex**, and then invoke the closures:

```go
func (w *Window) DispatchPointerDown(p geom.Point, btn int) {
    var clickHandler func()
    var downHandler func(*event.PointerEvent)
    var localP geom.Point

    w.mu.Lock()
    if w.rootNode == nil {
        w.mu.Unlock()
        return
    }

    hit, lp := w.rootNode.HitTestLocal(p)
    localP = lp
    if hit != nil {
        clickHandler = hit.OnClick
        downHandler = hit.OnPointerDown
    }
    w.mu.Unlock() // <--- UNLOCK BEFORE CALLING USER CODE!

    // User closures execute safely without lock contention
    if clickHandler != nil {
        clickHandler()
    }
    if downHandler != nil {
        downHandler(&event.PointerEvent{Position: localP})
    }
}
```

In Chapter 8, we explore Nova's Reactive State System and fine-grained signals!
