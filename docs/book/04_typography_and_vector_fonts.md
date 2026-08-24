# Chapter 4: Typography & Vector Fonts — The Text Shaping Engine

> *"How does Nova transform vector Bézier curves from TrueType `.ttf` font files into sharp, legible, subpixel anti-aliased text across all screen resolutions?"*

---

## 4.1 What is Inside a TrueType Font (`.ttf`)?

A TrueType font is not an array of images. It is a mathematical database containing:
1. **`cmap` Table**: Maps Unicode code points (e.g. `'A'` $\rightarrow U+0041$) to internal Glyph Index numbers.
2. **`glyf` Table**: Vector coordinate points defining the contour curves of each character.
3. **`hmtx` Table**: Horizontal metrics—advance widths (how many pixels to move the cursor forward after drawing the character) and left-side bearings.
4. **`head` / `hhea` Tables**: Global metrics—Ascent, Descent, LineGap, and Units Per Em (typically $1000$ or $2048$).

```
                 TrueType Font Architecture
+-------------------------------------------------------------+
| Unicode ('B')  ==>  cmap Table  ==>  Glyph Index #35        |
+-------------------------------------------------------------+
                              │
                              ▼
+-------------------------------------------------------------+
| glyf Table: Vector Contour Points:                          |
| Point 0: (120, 0)   [On-curve]                              |
| Point 1: (120, 700) [On-curve]                              |
| Point 2: (450, 700) [Off-curve Bézier Control Point]        |
| Point 3: (600, 520) [On-curve]                              |
| ...                                                         |
+-------------------------------------------------------------+
```

---

## 4.2 Mathematical Bézier Curves in Font Outlines

TrueType glyph contours are made of straight line segments and **Quadratic Bézier Curves**:

$$B(t) = (1 - t)^2 P_0 + 2(1 - t)t P_1 + t^2 P_2, \quad 0 \le t \le 1$$

Where:
- $P_0$: Starting on-curve anchor point.
- $P_1$: Off-curve control point (determines curvature direction and pull).
- $P_2$: Ending on-curve anchor point.

```
       P1 (Control Point)
       / \
      /   \
     /  .  \   <--- Quadratic Bézier Curve B(t)
    / .     \
   P0        P2
```

Nova scales these normalized font coordinates from font units ($2048\text{ units/em}$) to physical screen pixels:

$$\text{Scale Factor} = \frac{\text{FontSize in Pixels}}{\text{Units Per Em}}$$

$$\text{Screen } X = \text{Font } X \times \text{Scale Factor}$$

---

## 4.3 Text Metrics & Font Geometry

When laying out lines of text, Nova computes typography bounding boxes using four fundamental vertical metrics (`font/font.go`):

```
       Top Boundary of Bounding Box
       ---------------------------------------------  ▲
                                                      │ Ascent
       "Typography" Baseline                          │
       =============================================  ▼
                                                      │ Descent
       ---------------------------------------------  ▼
                                                      │ Line Gap
       ---------------------------------------------  ▲
```

- **Ascent**: Distance from baseline to the tallest capital letter (e.g. $'T', 'h', 'k'$).
- **Descent**: Distance below the baseline that characters extend (e.g. $'p', 'y', 'g', 'j'$).
- **Line Gap**: Recommended whitespace buffer between consecutive text lines.
- **Line Height**: $\text{Line Height} = \text{Ascent} + |\text{Descent}| + \text{Line Gap}$.

---

## 4.4 The Glyph Mask Cache: Achieving 60+ FPS

Parsing TrueType vector tables and solving Bézier equations on every single frame would be far too slow for smooth 60 FPS animation.

Nova implements an in-memory **Glyph Atlas & Alpha Mask Cache** (`font/cache.go`):

```
+-----------------------------------------------------------------+
|                        Glyph Cache Key                          |
|        (Rune: 'A', Size: 14px, Weight: Bold, SubpixelOffset)    |
+-----------------------------------------------------------------+
                                │
                  Cached Map Lookup (O(1))
                                ▼
+-----------------------------------------------------------------+
|                    Cached Alpha Mask (8-bit)                    |
|                                                                 |
|   0   0  12  85 240 255 255 240  85  12   0   0                 |
|   0  15 140 255 255  60  60 255 255 140  15   0                 |
|   0  80 255 255  20   0   0  20 255 255  80   0                 |
|  25 240 255 255 255 255 255 255 255 255 240  25                 |
|  90 255 255  30   0   0   0   0  30 255 255  90                 |
+-----------------------------------------------------------------+
```

### The Fast Glyph Blit:
When rendering text, Nova simply iterates over characters, looks up the pre-computed 8-bit alpha mask in the cache, and writes pixels into the RGBA framebuffer tinted with whatever text color was requested:

$$\text{Alpha} = \frac{\text{MaskCoverage} \times \text{TextColor.A}}{255}$$

```go
func blitGlyphMask(buf *image.RGBA, startX, startY int, mask *GlyphMask, col color.Color) {
    for row := 0; row < mask.Height; row++ {
        for colIdx := 0; colIdx < mask.Width; colIdx++ {
            coverage := mask.Coverage[row*mask.Width + colIdx]
            if coverage == 0 {
                continue
            }
            tintedColor := col.WithAlpha(float64(coverage)/255.0 * float64(col.A)/255.0)
            blendPixel(buf, startX+colIdx, startY+row, tintedColor)
        }
    }
}
```

---

## 4.5 Word Wrapping & Ellipsis Truncation Algorithms

When displaying paragraphs or responsive labels within a fixed-width container (`text/wrap.go`):

```go
func WrapText(str string, maxWidth float64, size float64, weight font.Weight) []string {
    words := strings.Fields(str)
    var lines []string
    var currentLine strings.Builder
    currentW := 0.0

    spaceW := text.MeasureText(" ", size, weight).Width

    for _, word := range words {
        wordSz := text.MeasureText(word, size, weight)
        if currentLine.Len() == 0 {
            currentLine.WriteString(word)
            currentW = wordSz.Width
        } else if currentW + spaceW + wordSz.Width <= maxWidth {
            currentLine.WriteString(" " + word)
            currentW += spaceW + wordSz.Width
        } else {
            lines = append(lines, currentLine.String())
            currentLine.Reset()
            currentLine.WriteString(word)
            currentW = wordSz.Width
        }
    }
    if currentLine.Len() > 0 {
        lines = append(lines, currentLine.String())
    }
    return lines
}
```

If text exceeds single-line constraints with truncation enabled:

$$\text{Truncated} = \text{FitSubstr}(S, \text{Width} - \text{Width}(\text{"..."})) + \text{"..."}$$

In Chapter 5, we explore the Declarative Component Tree and how Nova reconciles UI nodes!
