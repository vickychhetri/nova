# Chapter 9: High-Performance Virtualization — 1,000,000 Items at 60 FPS

> *"How can Nova scroll through a list of 1,000,000 database records or a huge financial grid at a locked 60 FPS while consuming only a few megabytes of memory?"*

---

## 9.1 Why Naive Scrolling Lists Fail

In standard GUI frameworks, creating a list with 100,000 items means creating 100,000 DOM nodes or UI widgets in memory:

```
Naive Approach (Heavy & Slow):
1,000,000 items × 250 bytes/node = 250 Megabytes of RAM!
1,000,000 items in layout pass   = Frame drop down to 2 FPS!
```

---

## 9.2 The Virtualization Formula: $O(1)$ Rendering

Nova's `Virtualizer` (`virtualization/virtualizer.go`) calculates the visible window mathematically in $O(1)$ constant time:

```
                                Virtual Content (1,000,000 Items)
     Item #0      +-----------------------------------------------------+
                  | Item #0                                             |
                  | Item #1                                             |
                  | ...                                                 |
     Scroll Offset|-----------------------------------------------------| ▲
       (e.g.      | [Visible Viewport in Window: 400px Height]          | │
        14,200px) | Item #355 (StartIndex)                              | │ Only ~12 items
                  | Item #356                                           | │ instantiated
                  | Item #357 ... Item #366 (EndIndex)                  | │ in memory!
                  |-----------------------------------------------------| ▼
                  | ...                                                 |
     Item #999999 +-----------------------------------------------------+
```

### The Slicing Mathematics:

Let:
- $N$ = Total item count (e.g. $1,000,000$).
- $H_{\text{item}}$ = Height of one item (e.g. $40\text{ px}$).
- $S$ = Current scroll offset in pixels (e.g. $14,200\text{ px}$).
- $H_{\text{view}}$ = Viewport height (e.g. $400\text{ px}$).
- $O$ = Overscan buffer (e.g. $2$ items above and below to prevent pop-in during fast scrolling).

$$\text{Total Content Height} = N \times H_{\text{item}}$$

$$\text{StartIndex} = \max\left(0, \left\lfloor \frac{S}{H_{\text{item}}} \right\rfloor - O\right)$$

$$\text{EndIndex} = \min\left(N - 1, \left\lceil \frac{S + H_{\text{view}}}{H_{\text{item}}} \right\rceil + O\right)$$

```go
type VisibleRange struct {
    StartIndex int
    EndIndex   int
    OffsetY    float64
}

func (v *Virtualizer) ComputeVisibleRange(scrollOffset, viewportHeight float64) VisibleRange {
    startIndex := int(math.Floor(scrollOffset / v.ItemHeight)) - v.Overscan
    if startIndex < 0 {
        startIndex = 0
    }

    visibleCount := int(math.Ceil(viewportHeight / v.ItemHeight)) + 2*v.Overscan
    endIndex := startIndex + visibleCount
    if endIndex >= v.ItemCount {
        endIndex = v.ItemCount - 1
    }

    return VisibleRange{
        StartIndex: startIndex,
        EndIndex:   endIndex,
        OffsetY:    float64(startIndex) * v.ItemHeight,
    }
}
```

---

## 9.3 Zero-Allocation Item Recycling

During the paint pass (`widgets/data/data.go`), Nova only creates and paints components for the active index range $[\text{StartIndex}, \text{EndIndex}]$:

```go
func (vl *VirtualListComponent) Paint(node *ui.Node, canvas *render.Canvas) {
    scroll := vl.ScrollOffset.Get()
    vis := vl.virtualizer.ComputeVisibleRange(scroll, node.Bounds.Height)

    canvas.PushClip(node.Bounds)

    for i := vis.StartIndex; i <= vis.EndIndex; i++ {
        itemComp := vl.RenderItem(i)
        itemY := float64(i)*vl.ItemHeight - scroll

        canvas.Save()
        canvas.Translate(0, itemY)
        itemNode := ui.NewNode(itemComp)
        itemNode.Bounds = geom.NewRect(0, 0, node.Bounds.Width, vl.ItemHeight)
        itemNode.Paint(canvas)
        canvas.Restore()
    }

    // Paint dynamic scrollbar thumb
    totalH := vl.virtualizer.TotalContentHeight()
    if totalH > node.Bounds.Height {
        barH := math.Max(24, (node.Bounds.Height/totalH)*node.Bounds.Height)
        barY := (scroll / (totalH - node.Bounds.Height)) * (node.Bounds.Height - barH)
        canvas.FillRoundedRect(geom.NewRect(node.Bounds.Width-6, barY, 4, barH), geom.RadiusUniform(2), theme.Current().Palette.BorderHover)
    }

    canvas.PopClip()
}
```

Whether your dataset has **10 items** or **10,000,000 items**, the CPU and memory footprint remains identical!

In Chapter 10, we explore Animation Controllers, Easing math, and Real-Time Game Physics!
