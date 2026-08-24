# Chapter 8: Reactive State & Reactivity — Fine-Grained Signals

> *"How does Nova implement fine-grained reactive state signaling in pure Go, ensuring that state changes instantly propagate to UI components without memory leaks or excessive rerenders?"*

---

## 8.1 The Anatomy of `state.Value[T]`

Nova provides a modern generic reactive signal system (`state/state.go`):

```go
type Value[T any] struct {
    mu        sync.RWMutex
    val       T
    listeners []func(T)
}
```

```
+-------------------------------------------------------------+
|                      state.Value[T]                         |
|  • Stores generic value T (string, int, bool, struct)       |
|  • Thread-safe Read/Write with sync.RWMutex                 |
+-------------------------------------------------------------+
            │                                    │
       .Get() (Read)                       .Set(newVal) (Write)
            │                                    │
            ▼                                    ▼
   Returns value to UI                  1. Compares with old value
                                        2. Updates internal field
                                        3. Broadcasts to all listeners!
```

### Type Helpers:
```go
name := state.String("Vicky")     // *state.Value[string]
counter := state.Int(0)           // *state.Value[int]
volume := state.Float(75.0)       // *state.Value[float64]
isOnline := state.Bool(true)      // *state.Value[bool]
```

---

## 8.2 Two-Way Component Data Binding

Form controls in Nova bind directly to reactive signals:

```go
emailState := state.String("user@domain.com")

// The TextField reads from emailState and updates emailState on typing:
textField := forms.TextField("Email").Bind(emailState)
```

When the user types on the keyboard:
1. `TextFieldComponent.OnKeyDown` updates `emailState.Set(newText)`.
2. Any listeners or UI elements watching `emailState` are notified automatically.

---

## 8.3 Computed Signals & Reactive Derivation

A **Computed Signal** is a reactive value that automatically recalculates whenever its upstream dependencies change (`state/computed.go`):

$$\text{Total} = \text{Price} \times \text{Quantity} \times (1 - \text{Discount})$$

```go
price := state.Float(100.0)
qty := state.Int(3)

totalPrice := state.Compute(func() float64 {
    return price.Get() * float64(qty.Get())
})
```

When either `price.Set(...)` or `qty.Set(...)` is called, `totalPrice` recomputes automatically!

---

## 8.4 Lightweight Canvas Invalidation vs. Full Tree Rebuilding

A key design rule in Nova is knowing when to trigger a **Lightweight Canvas Invalidation** vs. a **Full Tree Rebuild**:

```
+--------------------------------------------------------------------------------+
|                   SCENARIO 1: Real-Time Game / Animation (60 FPS)              |
|                                                                                |
|  • Action: Ball moves in Breakout, Snake advances 1 tile, Chart updates point  |
|  • Method: win.Invalidate()                                                   |
|  • Cost: ~0.1 ms (Zero allocation, no widget remounting, repaints canvas)      |
+--------------------------------------------------------------------------------+
|                   SCENARIO 2: Structural UI Changes (Tabs, Modals)             |
|                                                                                |
|  • Action: User switches from "Basic Tier" to "Leaderboard" tab               |
|  • Method: win.Content(buildMainView())                                        |
|  • Cost: ~1.2 ms (Reconciles new component tree, mounts nodes, solves layout)  |
+--------------------------------------------------------------------------------+
```

In Chapter 9, we explore how Nova virtualizes lists of 1,000,000+ items at 60 FPS!
