# Nova Functional Specification

This document defines the functional behavior, algorithms, data structures, and operational contracts of all Nova framework subsystems.

---

## 1. Subsystem Specifications

### 1.1 Geometry & Coordinates (`core/geom`)

- **Coordinate System**: Top-left origin `(0, 0)` with X increasing rightward and Y increasing downward.
- **Bounding Boxes (`geom.Rect`)**: Defined by `(X, Y, Width, Height)`. Supports `ContainsPoint`, `ContainsRect`, `Intersects`, `Intersection`, `Union`, `Inset`, and `Offset`.
- **Corner Radii (`geom.CornerRadius`)**: Supports independent `TopLeft`, `TopRight`, `BottomRight`, and `BottomLeft` rounding with continuous subpixel curves.
- **2D Transformation (`geom.Matrix2D`)**: 3x3 affine transformation matrix for translation, scaling, and rotation.

---

### 1.2 Color & Blending (`core/color`)

- **Color Model**: High-precision RGBA float64 channels normalized in `[0.0, 1.0]`.
- **Color Spaces**:
  - `RGBA(r, g, b, a uint8)`
  - `Hex(hexStr string)` supporting `#RGB`, `#RGBA`, `#RRGGBB`, `#RRGGBBAA`.
  - `HSL(h, s, l float64)` & `HSLA(h, s, l, a float64)` with standard trigonometric conversions.
- **Alpha Compositing**: Implements Porter-Duff source-over blending:
  $$\alpha_{out} = \alpha_{src} + \alpha_{dst}(1 - \alpha_{src})$$
  $$C_{out} = \frac{C_{src}\alpha_{src} + C_{dst}\alpha_{dst}(1 - \alpha_{src})}{\alpha_{out}}$$
- **WCAG 2.1 Contrast Calculation**: Relative luminance formula calculating contrast ratios from `1.0` to `21.0`.

---

### 1.3 Reactive Signals Engine (`state`)

- **Dependency Tracking**:
  - Context stack tracks active subscriptions automatically during computation evaluations (`state.Compute`) and effect executions (`state.Effect`).
  - Listeners execute with atomic thread safety (`sync.RWMutex`).
- **Batching (`state.Batch`)**: Suppresses listener notifications across multiple mutations until the outer batch scope concludes.

---

### 1.4 Layout Engine (`layout`)

- **Constraint Model**: `BoxConstraints(MinWidth, MaxWidth, MinHeight, MaxHeight)`.
- **Flexbox Algorithm (`layout.ComputeFlex`)**:
  1. Pass 1: Measure non-flexible children (`Flex == 0`).
  2. Pass 2: Calculate remaining main-axis space and allocate proportionally to flexible children (`Flex > 0`).
  3. Pass 3: Align children along main axis (`Start`, `Center`, `End`, `SpaceBetween`, `SpaceAround`, `SpaceEvenly`) and cross axis (`Start`, `Center`, `End`, `Stretch`).
- **Stack Algorithm (`layout.ComputeStack`)**: Sized by non-positioned children; positioned children evaluate explicit `Top`, `Right`, `Bottom`, `Left`, `Width`, `Height` anchors.
- **Grid Algorithm (`layout.ComputeGrid`)**: Multi-column cell positioning supporting explicit heights or aspect ratios.

---

### 1.5 Render Engine & Rasterizer (`render`, `renderer/software`)

- **Display List Buffering**: Retained `render.CommandBuffer` preallocates draw commands per frame.
- **Clipping Stack**: Supports nested hierarchical rectangular and rounded clipping regions.
- **Anti-Aliased Rasterization**: Subpixel edge testing for rounded rectangles, smooth lines, circular arcs, drop shadow blur approximation, and alpha-blitted typography.

---

### 1.6 Text & Typography Engine (`text`, `font`)

- **Font Metrics**: Proportional character width measurement, ascender/descender bounds, and baseline alignment.
- **Word Wrapping**: Greedy line-breaking algorithm preserving word boundaries.
- **Ellipsis Truncation**: Truncates overflowing text and appends `...` to fit within bounded width.

---

### 1.7 Virtualization Engine (`virtualization`)

- **Computational Complexity**: $\mathcal{O}(1)$ visible range lookup for fixed item heights.
- **Overscan Buffering**: Pre-renders configurable buffer items above and below the viewport to ensure tear-free, 120 FPS continuous scrolling.
- **Viewport Slicing Contract**:
  $$\text{FirstVisible} = \lfloor \text{ScrollOffset} / \text{ItemHeight} \rfloor$$
  $$\text{StartIndex} = \max(0, \text{FirstVisible} - \text{Overscan})$$
  $$\text{EndIndex} = \min(\text{TotalCount} - 1, \text{FirstVisible} + \text{VisibleCount} + \text{Overscan})$$

---

### 1.8 Form Validation Engine (`widgets/forms`)

- **Validation Rules**:
  - `Required()`: Verifies non-empty and non-whitespace string values.
  - `MinLength(n)`: Verifies string length $\ge n$.
  - `MaxLength(n)`: Verifies string length $\le n$.
  - `Email()`: RFC 5322 compliant email regex matching.
  - Custom rule functions: `func(value any) string`.
- **Form State Lifecycle**:
  - `RegisterField(name, rules...)`
  - `Set(name, value)` $\rightarrow$ marks dirty & triggers field validation
  - `Submit()` $\rightarrow$ executes validation over all fields and calls `OnSubmit` on success.
