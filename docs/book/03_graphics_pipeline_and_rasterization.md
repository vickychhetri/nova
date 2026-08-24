# Chapter 3: Graphics Pipeline & Rasterization — The 2D Software Engine

> *"How does Nova take abstract draw commands like `FillRoundedRect` or `DrawLine` and turn them into smooth, anti-aliased pixels inside a pure Go software rasterizer?"*

---

## 3.1 The Command Buffer Pattern: Immediate vs. Retained Painting

In Nova, UI components do **not** draw directly into the pixel array. If every component drew directly to the framebuffer during the UI tree traversal:
- Clipping parent bounds would be difficult.
- Floating overlays, modals, and dropdown menus would be overwritten by sibling elements.
- Transformations, scrolling offsets, and canvas invalidation would require expensive full-screen redraws.

Instead, Nova uses the **Command Buffer Pattern** (`render/command_buffer.go` and `render/canvas.go`):

```
+------------------+        Paint()        +------------------------+
|  UI Component    |  ==================>  |  render.Canvas         |
| (Button / Card)  |                       | (High-level API)       |
+------------------+                       +------------------------+
                                                       │ Emits
                                                       ▼
                                           +------------------------+
                                           |  render.CommandBuffer  |
                                           | (Sequential Draw Cmds) |
                                           +------------------------+
                                                       │
                                                       │ Render() pass
                                                       ▼
                                           +------------------------+
                                           |  software.Rasterizer   |
                                           | (Scanlines & Math)     |
                                           +------------------------+
                                                       │
                                                       ▼
                                           +------------------------+
                                           |  *image.RGBA Buffer    |
                                           +------------------------+
```

### The Draw Command Enum:
```go
type CommandType int

const (
    CmdFillRect CommandType = iota
    CmdStrokeRect
    CmdFillRoundedRect
    CmdStrokeRoundedRect
    CmdDrawLine
    CmdFillCircle
    CmdStrokeCircle
    CmdDrawText
    CmdPushClip
    CmdPopClip
    CmdDrawImage
)
```

---

## 3.2 Pure Go Rasterization Algorithms

Let's inspect how the software rasterizer (`renderer/software/rasterizer.go`) implements core geometric rendering in pure Go without external C libraries.

### 1. Solid & Rounded Rectangles
To render a rectangle with rounded corners of radius $r$:
For every coordinate $(x, y)$ inside bounding box $[X_0, Y_0, X_1, Y_1]$:

```
          (x0, y0) +----+------------+----+ (x1, y0)
                   | C1 |  Top Edge  | C2 |
                   +----+------------+----+
                   |    |            |    |
                   | L  |   Center   | R  |
                   |    |            |    |
                   +----+------------+----+
                   | C3 | Bottom Edge| C4 |
          (x0, y1) +----+------------+----+ (x1, y1)
```

1. If $(x, y)$ lies in the central body or edge rectangles $\implies$ **Fill immediately**.
2. If $(x, y)$ lies in one of the 4 corner boxes $C_1, C_2, C_3, C_4$:
   Compute Euclidean distance $d$ from the corner circle center $(c_x, c_y)$:

$$d = \sqrt{(x - c_x)^2 + (y - c_y)^2}$$

- If $d \le r - 0.5 \implies$ Fully inside (100% alpha).
- If $r - 0.5 < d \le r + 0.5 \implies$ Subpixel antialiasing edge ($0 < \text{alpha} < 1$).
- If $d > r + 0.5 \implies$ Outside the rounded corner (discard pixel).

---

### 2. Bresenham's Integer Line Algorithm with Thickness

To draw lines rapidly between $(X_1, Y_1)$ and $(X_2, Y_2)$ with zero floating-point division:

$$\Delta X = |X_2 - X_1|, \quad \Delta Y = |Y_2 - Y_1|$$

An integer error term $D = 2\Delta Y - \Delta X$ tracks whether to step horizontally or diagonally on each iteration:

```go
func drawLineBresenham(buf *image.RGBA, x0, y0, x1, y1 int, col color.Color) {
    dx := abs(x1 - x0)
    dy := -abs(y1 - y0)
    sx := 1
    if x0 >= x1 { sx = -1 }
    sy := 1
    if y0 >= y1 { sy = -1 }
    err := dx + dy

    for {
        blendPixel(buf, x0, y0, col)
        if x0 == x1 && y0 == y1 { break }
        e2 := 2 * err
        if e2 >= dy {
            err += dy
            x0 += sx
        }
        if e2 <= dx {
            err += dx
            y0 += sy
        }
    }
}
```

---

## 3.3 The Porter-Duff Alpha Blending Operator

When rendering translucent layers (such as a 40% opaque dark overlay, a neon glow circle, or antialiased font edges), the newly rendered pixel (Source, $S$) must blend smoothly on top of whatever was already drawn in the framebuffer (Destination, $D$).

### The Porter-Duff Over Formula:

$$C_{\text{out}} = \alpha_{\text{src}} \cdot C_{\text{src}} + (1 - \alpha_{\text{src}}) \cdot C_{\text{dst}}$$

$$\alpha_{\text{out}} = \alpha_{\text{src}} + \alpha_{\text{dst}} \cdot (1 - \alpha_{\text{src}})$$

### High-Performance Integer Blending in Nova:
Floating point operations per pixel at $1920 \times 1080$ ($2,073,600$ pixels) would consume too many CPU cycles. Nova performs alpha blending strictly using **integer arithmetic**:

```go
func blendPixel(buf *image.RGBA, x, y int, src color.Color) {
    if x < 0 || x >= buf.Rect.Dx() || y < 0 || y >= buf.Rect.Dy() {
        return
    }
    offset := y*buf.Stride + x*4
    sa := uint32(src.A)
    if sa == 0 {
        return
    }
    if sa == 255 {
        buf.Pix[offset+0] = src.R
        buf.Pix[offset+1] = src.G
        buf.Pix[offset+2] = src.B
        buf.Pix[offset+3] = 255
        return
    }

    invA := 255 - sa
    dr := uint32(buf.Pix[offset+0])
    dg := uint32(buf.Pix[offset+1])
    db := uint32(buf.Pix[offset+2])
    da := uint32(buf.Pix[offset+3])

    buf.Pix[offset+0] = uint8((uint32(src.R)*sa + dr*invA) / 255)
    buf.Pix[offset+1] = uint8((uint32(src.G)*sa + dg*invA) / 255)
    buf.Pix[offset+2] = uint8((uint32(src.B)*sa + db*invA) / 255)
    buf.Pix[offset+3] = uint8(sa + (da*invA)/255)
}
```

---

## 3.4 Clipping Regions & Scissor Testing

When a `ListView`, `VirtualList`, or `Card` contains child components that extend beyond its physical rectangular boundaries (e.g. while scrolling), Nova uses a **Scissor Clip Stack**:

```go
canvas.PushClip(bounds) // Limits all subsequent draw operations to bounds
// ... draw child components ...
canvas.PopClip()        // Restores previous clipping region
```

The rasterizer checks every pixel against the active clip rectangle before writing:

```go
if x < clip.Min.X || x >= clip.Max.X || y < clip.Min.Y || y >= clip.Max.Y {
    continue // Discard pixel outside scissor region!
}
```

In Chapter 4, we explore how Nova renders TrueType vector fonts and handles paragraph layout!
