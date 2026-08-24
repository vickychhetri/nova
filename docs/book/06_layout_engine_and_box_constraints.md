# Chapter 6: Layout Engine & Box Constraints — Constraint Solving & Flexbox

> *"How does Nova determine the exact position $(X, Y)$ and dimensions $(\text{Width}, \text{Height})$ of every button, text label, and container across responsive window sizes?"*

---

## 6.1 The Rule of 2D Layout: "Constraints Down, Sizes Up, Parent Sets Position"

Nova implements the box constraint layout architecture (similar to Flutter and CSS Subgrid):

```
                       Parent Component
                              │
                    1. Passes BoxConstraints (Min/Max W & H)
                              ▼
                       Child Component
                              │
                    2. Computes & Returns exact Size (W, H)
                              ▼
                       Parent Component
                              │
                    3. Assigns exact Bounds.Origin (X, Y)
```

1. **Constraints go Down**: A parent passes bounding constraints to its children (e.g. *"You can be anywhere between 100px and 500px wide, and exactly 40px tall"*).
2. **Sizes go Up**: The child measures its internal contents (text length, icons, child nodes) and tells the parent its exact resolved dimensions $(W, H)$.
3. **Parent Sets Position**: The parent places the child at a specific coordinate $(X, Y)$ in the render tree.

---

## 6.2 The `layout.BoxConstraints` Model

A constraint is defined by 4 bounding numbers (`layout/constraints.go`):

$$0 \le \text{MinWidth} \le \text{MaxWidth} \le \infty$$

$$0 \le \text{MinHeight} \le \text{MaxHeight} \le \infty$$

```go
type BoxConstraints struct {
    MinWidth  float64
    MaxWidth  float64
    MinHeight float64
    MaxHeight float64
}
```

### Constraint Types:
- **Tight Constraints**: $\text{MinWidth} == \text{MaxWidth}$ and $\text{MinHeight} == \text{MaxHeight}$. The child is forced to take an exact fixed size (e.g. the Root Window boundary `Tight(win.Size())`).
- **Loose Constraints**: $\text{MinWidth} == 0$ and $\text{MinHeight} == 0$. The child can be as small as it wants, up to $\text{MaxWidth}$ and $\text{MaxHeight}$.
- **Unbounded / Infinite Constraints**: $\text{MaxWidth} == \infty$ (e.g. inside a horizontally scrollable list).

---

## 6.3 The Flexbox Algorithm (`Row` and `Column`)

The Flexbox layout engine (`layout/flex.go`) calculates the geometry of `ui.Row` and `ui.Column` through a 5-step pass:

```
[Fixed Button] ─── Gap ───> [Spacer Element] ─── Gap ───> [Fixed Badge]
   (120px)          (12px)       (Auto-Expand)       (12px)      (80px)
|<──────────────────────────── Total Width: 600px ───────────────────────────>|
```

### The Math:

Let $W_{\text{total}}$ be the total available width from parent constraints.
Let $N$ be the number of children, and $G$ be the gap spacing between children.
Let $F$ be the set of fixed/inflexible children, and $S$ be the count of `Spacer` flexible children.

$$\text{Total Gap Width} = (N - 1) \times G$$

$$\text{Fixed Width Sum} = \sum_{i \in F} W_i$$

$$\text{Remaining Free Space} = W_{\text{total}} - \text{Fixed Width Sum} - \text{Total Gap Width}$$

$$\text{Per-Spacer Width} = \frac{\text{Remaining Free Space}}{S}$$

```go
func LayoutFlexRow(children []*ui.Node, constraints layout.BoxConstraints, gap float64) geom.Size {
    var fixedWidthSum float64
    var spacerCount int

    // Step 1: Measure non-flexible fixed children
    for _, child := range children {
        if isSpacer(child) {
            spacerCount++
        } else {
            sz := child.Layout(layout.Loose(constraints.MaxSize()))
            child.Bounds.Size = sz
            fixedWidthSum += sz.Width
        }
    }

    // Step 2: Calculate space for flexible spacers
    totalGaps := float64(len(children)-1) * gap
    freeSpace := math.Max(0, constraints.MaxWidth - fixedWidthSum - totalGaps)
    spacerW := 0.0
    if spacerCount > 0 {
        spacerW = freeSpace / float64(spacerCount)
    }

    // Step 3: Position each child horizontally
    curX := 0.0
    maxH := 0.0
    for _, child := range children {
        w := child.Bounds.Width
        if isSpacer(child) {
            w = spacerW
        }
        child.Bounds.Origin = geom.Pt(curX, 0)
        child.Bounds.Width = w
        if child.Bounds.Height > maxH {
            maxH = child.Bounds.Height
        }
        curX += w + gap
    }

    return geom.Sz(curX - gap, maxH)
}
```

---

## 6.4 Cross-Axis Alignment

Along the perpendicular axis (vertical height in a `Row`, or horizontal width in a `Column`), Nova solves alignment according to `CrossAxisAlignment`:

- **`Start`**: Align child to top / left ($Y = 0$).
- **`Center`**: Center child vertically: $Y = \frac{\text{MaxRowHeight} - \text{ChildHeight}}{2}$.
- **`End`**: Align child to bottom ($Y = \text{MaxRowHeight} - \text{ChildHeight}$).
- **`Stretch`**: Force child height to match the full row height ($\text{ChildHeight} = \text{MaxRowHeight}$).

In Chapter 7, we explore the Event Loop, Hit Testing, and thread-safe Input Routing!
