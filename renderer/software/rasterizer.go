package software

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"

	corecolor "github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/renderer"
)

// Rasterizer is a high-performance software renderer that outputs to an RGBA pixel buffer.
type Rasterizer struct {
	target    renderer.RenderTarget
	buffer    *image.RGBA
	width     int
	height    int
	scale     float64
	clipStack []image.Rectangle
}

// NewRasterizer creates a new software rasterizer.
func NewRasterizer() *Rasterizer {
	return &Rasterizer{
		clipStack: make([]image.Rectangle, 0, 16),
	}
}

// Init initializes rasterizer.
func (r *Rasterizer) Init(target renderer.RenderTarget) error {
	r.target = target
	return nil
}

// Buffer returns current RGBA pixel buffer.
func (r *Rasterizer) Buffer() *image.RGBA {
	return r.buffer
}

// BeginFrame prepares pixel buffer for drawing.
func (r *Rasterizer) BeginFrame(size geom.Size, scale float64) {
	r.scale = scale
	if r.scale <= 0 {
		r.scale = 1.0
	}

	w := int(math.Ceil(size.Width * r.scale))
	h := int(math.Ceil(size.Height * r.scale))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	if r.buffer == nil || r.width != w || r.height != h {
		r.width = w
		r.height = h
		r.buffer = image.NewRGBA(image.Rect(0, 0, w, h))
	} else {
		// Clear buffer with transparent black
		for i := range r.buffer.Pix {
			r.buffer.Pix[i] = 0
		}
	}

	r.clipStack = r.clipStack[:0]
	r.clipStack = append(r.clipStack, r.buffer.Bounds())
}

// Render executes all render commands.
func (r *Rasterizer) Render(commands []render.Command) {
	for _, cmd := range commands {
		r.executeCommand(cmd)
	}
}

// EndFrame finishes frame rendering.
func (r *Rasterizer) EndFrame() {}

// Destroy cleans up resources.
func (r *Rasterizer) Destroy() {
	r.buffer = nil
}

// SavePNG exports current framebuffer to a PNG file on disk.
func (r *Rasterizer) SavePNG(path string) error {
	if r.buffer == nil {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, r.buffer)
}

func (r *Rasterizer) currentClip() image.Rectangle {
	if len(r.clipStack) == 0 {
		return r.buffer.Bounds()
	}
	return r.clipStack[len(r.clipStack)-1]
}

func (r *Rasterizer) executeCommand(cmd render.Command) {
	switch cmd.Type {
	case render.CmdPushClip:
		clip := r.rectToImageRect(cmd.Bounds).Intersect(r.currentClip())
		r.clipStack = append(r.clipStack, clip)

	case render.CmdPopClip:
		if len(r.clipStack) > 1 {
			r.clipStack = r.clipStack[:len(r.clipStack)-1]
		}

	case render.CmdFillRect:
		r.fillRect(cmd.Bounds, cmd.Color)

	case render.CmdStrokeRect:
		r.strokeRect(cmd.Bounds, cmd.StrokeColor, cmd.StrokeWidth)

	case render.CmdFillRoundedRect:
		r.fillRoundedRect(cmd.Bounds, cmd.Radius, cmd.Color)

	case render.CmdStrokeRoundedRect:
		r.strokeRoundedRect(cmd.Bounds, cmd.Radius, cmd.StrokeColor, cmd.StrokeWidth)

	case render.CmdFillCircle:
		r.fillCircle(cmd.P1, cmd.Radius.TopLeft, cmd.Color)

	case render.CmdStrokeCircle:
		r.strokeCircle(cmd.P1, cmd.Radius.TopLeft, cmd.StrokeColor, cmd.StrokeWidth)

	case render.CmdLine:
		r.drawLine(cmd.P1, cmd.P2, cmd.StrokeColor, cmd.StrokeWidth)

	case render.CmdText:
		r.drawText(cmd.Text, cmd.P1, cmd.FontSize, cmd.FontWeight, cmd.Color)

	case render.CmdImage:
		r.drawImage(cmd.Image, cmd.Bounds)

	case render.CmdShadow:
		r.drawShadow(cmd.Bounds, cmd.Radius, cmd.Shadow)
	}
}

func (r *Rasterizer) rectToImageRect(rect geom.Rect) image.Rectangle {
	x0 := int(math.Floor(rect.X * r.scale))
	y0 := int(math.Floor(rect.Y * r.scale))
	x1 := int(math.Ceil((rect.X + rect.Width) * r.scale))
	y1 := int(math.Ceil((rect.Y + rect.Height) * r.scale))
	return image.Rect(x0, y0, x1, y1)
}

func (r *Rasterizer) blendPixel(x, y int, src corecolor.Color) {
	if src.A <= 0.001 {
		return
	}
	clip := r.currentClip()
	if x < clip.Min.X || x >= clip.Max.X || y < clip.Min.Y || y >= clip.Max.Y {
		return
	}

	idx := (y*r.buffer.Stride) + (x * 4)
	if idx < 0 || idx+3 >= len(r.buffer.Pix) {
		return
	}

	dstR := float64(r.buffer.Pix[idx+0]) / 255.0
	dstG := float64(r.buffer.Pix[idx+1]) / 255.0
	dstB := float64(r.buffer.Pix[idx+2]) / 255.0
	dstA := float64(r.buffer.Pix[idx+3]) / 255.0

	// Alpha compositing (source-over)
	outA := src.A + dstA*(1.0-src.A)
	if outA > 0 {
		outR := (src.R*src.A + dstR*dstA*(1.0-src.A)) / outA
		outG := (src.G*src.A + dstG*dstA*(1.0-src.A)) / outA
		outB := (src.B*src.A + dstB*dstA*(1.0-src.A)) / outA

		r.buffer.Pix[idx+0] = uint8(math.Round(outR * 255))
		r.buffer.Pix[idx+1] = uint8(math.Round(outG * 255))
		r.buffer.Pix[idx+2] = uint8(math.Round(outB * 255))
		r.buffer.Pix[idx+3] = uint8(math.Round(outA * 255))
	}
}

func (r *Rasterizer) fillRect(rect geom.Rect, col corecolor.Color) {
	if col.A <= 0 {
		return
	}
	bounds := r.rectToImageRect(rect).Intersect(r.currentClip())
	if bounds.Empty() {
		return
	}

	if col.A >= 0.999 {
		// Fast opaque fill
		nrgba := col.NRGBA()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			rowStart := (y * r.buffer.Stride) + (bounds.Min.X * 4)
			rowEnd := (y * r.buffer.Stride) + (bounds.Max.X * 4)
			for idx := rowStart; idx < rowEnd; idx += 4 {
				r.buffer.Pix[idx+0] = nrgba.R
				r.buffer.Pix[idx+1] = nrgba.G
				r.buffer.Pix[idx+2] = nrgba.B
				r.buffer.Pix[idx+3] = 255
			}
		}
	} else {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r.blendPixel(x, y, col)
			}
		}
	}
}

func (r *Rasterizer) strokeRect(rect geom.Rect, col corecolor.Color, strokeWidth float64) {
	if col.A <= 0 || strokeWidth <= 0 {
		return
	}
	sw := strokeWidth
	// Top
	r.fillRect(geom.NewRect(rect.X, rect.Y, rect.Width, sw), col)
	// Bottom
	r.fillRect(geom.NewRect(rect.X, rect.Y+rect.Height-sw, rect.Width, sw), col)
	// Left
	r.fillRect(geom.NewRect(rect.X, rect.Y+sw, sw, math.Max(0, rect.Height-2*sw)), col)
	// Right
	r.fillRect(geom.NewRect(rect.X+rect.Width-sw, rect.Y+sw, sw, math.Max(0, rect.Height-2*sw)), col)
}

func (r *Rasterizer) fillRoundedRect(rect geom.Rect, radius geom.CornerRadius, col corecolor.Color) {
	if col.A <= 0 {
		return
	}
	if radius.TopLeft == 0 && radius.TopRight == 0 && radius.BottomRight == 0 && radius.BottomLeft == 0 {
		r.fillRect(rect, col)
		return
	}

	bounds := r.rectToImageRect(rect).Intersect(r.currentClip())
	if bounds.Empty() {
		return
	}

	rx0 := rect.X * r.scale
	ry0 := rect.Y * r.scale
	rx1 := (rect.X + rect.Width) * r.scale
	ry1 := (rect.Y + rect.Height) * r.scale

	rtl := radius.TopLeft * r.scale
	rtr := radius.TopRight * r.scale
	rbr := radius.BottomRight * r.scale
	rbl := radius.BottomLeft * r.scale

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		fy := float64(y) + 0.5
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			fx := float64(x) + 0.5
			inside := true

			// Top-Left corner
			if fx < rx0+rtl && fy < ry0+rtl {
				dx := (rx0 + rtl) - fx
				dy := (ry0 + rtl) - fy
				if dx*dx+dy*dy > rtl*rtl {
					inside = false
				}
			}
			// Top-Right corner
			if fx > rx1-rtr && fy < ry0+rtr {
				dx := fx - (rx1 - rtr)
				dy := (ry0 + rtr) - fy
				if dx*dx+dy*dy > rtr*rtr {
					inside = false
				}
			}
			// Bottom-Right corner
			if fx > rx1-rbr && fy > ry1-rbr {
				dx := fx - (rx1 - rbr)
				dy := fy - (ry1 - rbr)
				if dx*dx+dy*dy > rbr*rbr {
					inside = false
				}
			}
			// Bottom-Left corner
			if fx < rx0+rbl && fy > ry1-rbl {
				dx := (rx0 + rbl) - fx
				dy := fy - (ry1 - rbl)
				if dx*dx+dy*dy > rbl*rbl {
					inside = false
				}
			}

			if inside {
				r.blendPixel(x, y, col)
			}
		}
	}
}

func (r *Rasterizer) strokeRoundedRect(rect geom.Rect, radius geom.CornerRadius, col corecolor.Color, strokeWidth float64) {
	if col.A <= 0 || strokeWidth <= 0 {
		return
	}

	bounds := r.rectToImageRect(rect).Intersect(r.currentClip())
	if bounds.Empty() {
		return
	}

	rx0 := rect.X * r.scale
	ry0 := rect.Y * r.scale
	rx1 := (rect.X + rect.Width) * r.scale
	ry1 := (rect.Y + rect.Height) * r.scale

	sw := strokeWidth * r.scale

	rtl := radius.TopLeft * r.scale
	rtr := radius.TopRight * r.scale
	rbr := radius.BottomRight * r.scale
	rbl := radius.BottomLeft * r.scale

	irx0 := rx0 + sw
	iry0 := ry0 + sw
	irx1 := rx1 - sw
	iry1 := ry1 - sw

	irtl := math.Max(0, rtl-sw)
	irtr := math.Max(0, rtr-sw)
	irbr := math.Max(0, rbr-sw)
	irbl := math.Max(0, rbl-sw)

	isInsideOuter := func(fx, fy float64) bool {
		if fx < rx0 || fx > rx1 || fy < ry0 || fy > ry1 {
			return false
		}
		if fx < rx0+rtl && fy < ry0+rtl {
			dx := (rx0 + rtl) - fx
			dy := (ry0 + rtl) - fy
			if dx*dx+dy*dy > rtl*rtl {
				return false
			}
		}
		if fx > rx1-rtr && fy < ry0+rtr {
			dx := fx - (rx1 - rtr)
			dy := (ry0 + rtr) - fy
			if dx*dx+dy*dy > rtr*rtr {
				return false
			}
		}
		if fx > rx1-rbr && fy > ry1-rbr {
			dx := fx - (rx1 - rbr)
			dy := fy - (ry1 - rbr)
			if dx*dx+dy*dy > rbr*rbr {
				return false
			}
		}
		if fx < rx0+rbl && fy > ry1-rbl {
			dx := (rx0 + rbl) - fx
			dy := fy - (ry1 - rbl)
			if dx*dx+dy*dy > rbl*rbl {
				return false
			}
		}
		return true
	}

	isInsideInner := func(fx, fy float64) bool {
		if fx < irx0 || fx > irx1 || fy < iry0 || fy > iry1 {
			return false
		}
		if fx < irx0+irtl && fy < iry0+irtl {
			dx := (irx0 + irtl) - fx
			dy := (iry0 + irtl) - fy
			if dx*dx+dy*dy > irtl*irtl {
				return false
			}
		}
		if fx > irx1-irtr && fy < iry0+irtr {
			dx := fx - (irx1 - irtr)
			dy := (iry0 + irtr) - fy
			if dx*dx+dy*dy > irtr*irtr {
				return false
			}
		}
		if fx > irx1-irbr && fy > iry1-irbr {
			dx := fx - (irx1 - irbr)
			dy := fy - (iry1 - irbr)
			if dx*dx+dy*dy > irbr*irbr {
				return false
			}
		}
		if fx < irx0+irbl && fy > iry1-irbl {
			dx := (irx0 + irbl) - fx
			dy := fy - (iry1 - irbl)
			if dx*dx+dy*dy > irbl*irbl {
				return false
			}
		}
		return true
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		fy := float64(y) + 0.5
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			fx := float64(x) + 0.5
			if isInsideOuter(fx, fy) && !isInsideInner(fx, fy) {
				r.blendPixel(x, y, col)
			}
		}
	}
}

func (r *Rasterizer) fillCircle(center geom.Point, radius float64, col corecolor.Color) {
	if col.A <= 0 || radius <= 0 {
		return
	}
	cx := center.X * r.scale
	cy := center.Y * r.scale
	cr := radius * r.scale

	bounds := image.Rect(
		int(math.Floor(cx-cr)),
		int(math.Floor(cy-cr)),
		int(math.Ceil(cx+cr)),
		int(math.Ceil(cy+cr)),
	).Intersect(r.currentClip())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		fy := float64(y) + 0.5
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			fx := float64(x) + 0.5
			dx := fx - cx
			dy := fy - cy
			if dx*dx+dy*dy <= cr*cr {
				r.blendPixel(x, y, col)
			}
		}
	}
}

func (r *Rasterizer) strokeCircle(center geom.Point, radius float64, col corecolor.Color, strokeWidth float64) {
	if col.A <= 0 || radius <= 0 || strokeWidth <= 0 {
		return
	}
	cx := center.X * r.scale
	cy := center.Y * r.scale
	outerR := (radius + strokeWidth/2.0) * r.scale
	innerR := (radius - strokeWidth/2.0) * r.scale

	bounds := image.Rect(
		int(math.Floor(cx-outerR)),
		int(math.Floor(cy-outerR)),
		int(math.Ceil(cx+outerR)),
		int(math.Ceil(cy+outerR)),
	).Intersect(r.currentClip())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		fy := float64(y) + 0.5
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			fx := float64(x) + 0.5
			dx := fx - cx
			dy := fy - cy
			distSq := dx*dx + dy*dy
			if distSq <= outerR*outerR && distSq >= innerR*innerR {
				r.blendPixel(x, y, col)
			}
		}
	}
}

func (r *Rasterizer) drawLine(p1, p2 geom.Point, col corecolor.Color, strokeWidth float64) {
	if col.A <= 0 || strokeWidth <= 0 {
		return
	}
	x0 := int(math.Round(p1.X * r.scale))
	y0 := int(math.Round(p1.Y * r.scale))
	x1 := int(math.Round(p2.X * r.scale))
	y1 := int(math.Round(p2.Y * r.scale))
	sw := int(math.Max(1, math.Round(strokeWidth*r.scale)))

	dx := int(math.Abs(float64(x1 - x0)))
	dy := int(math.Abs(float64(y1 - y0)))
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	for {
		for ox := -sw / 2; ox <= sw/2; ox++ {
			for oy := -sw / 2; oy <= sw/2; oy++ {
				r.blendPixel(x0+ox, y0+oy, col)
			}
		}

		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func (r *Rasterizer) drawText(text string, origin geom.Point, fontSize float64, weight int, col corecolor.Color) {
	if text == "" || col.A <= 0 {
		return
	}

	curX := origin.X * r.scale
	destY := int(math.Round(origin.Y * r.scale))
	scaledFontSize := fontSize * r.scale

	for _, rn := range text {
		gm := font.GetGlyph(rn, scaledFontSize, weight)
		if gm.Bitmap != nil {
			bmpW := gm.Bitmap.Bounds().Dx()
			bmpH := gm.Bitmap.Bounds().Dy()
			destX := int(math.Round(curX))

			for gy := 0; gy < bmpH; gy++ {
				for gx := 0; gx < bmpW; gx++ {
					alphaVal := gm.Bitmap.AlphaAt(gx, gy).A
					if alphaVal > 0 {
						aFactor := float64(alphaVal) / 255.0
						if aFactor > 0.04 {
							aFactor = math.Pow(aFactor, 0.72)
						}
						pixCol := col.MultiplyAlpha(aFactor)
						r.blendPixel(destX+gx, destY+gy, pixCol)
					}
				}
			}
		}
		curX += gm.AdvanceX * r.scale
	}
}

func (r *Rasterizer) drawImage(img image.Image, dest geom.Rect) {
	if img == nil {
		return
	}
	imgBounds := r.rectToImageRect(dest).Intersect(r.currentClip())
	if imgBounds.Empty() {
		return
	}
	draw.Draw(r.buffer, imgBounds, img, image.Point{}, draw.Over)
}

func (r *Rasterizer) drawShadow(bounds geom.Rect, radius geom.CornerRadius, shadow render.ShadowParams) {
	if shadow.Color.A <= 0 {
		return
	}
	shadowBounds := bounds.Offset(shadow.OffsetX, shadow.OffsetY).Inset(geom.All(-shadow.Spread))
	shadowRadius := geom.CornerRadius{
		TopLeft:     radius.TopLeft + shadow.Spread,
		TopRight:    radius.TopRight + shadow.Spread,
		BottomRight: radius.BottomRight + shadow.Spread,
		BottomLeft:  radius.BottomLeft + shadow.Spread,
	}

	// Approximate soft shadow blur
	steps := int(math.Max(1, shadow.Blur/2.0))
	for i := steps; i >= 1; i-- {
		factor := float64(i) / float64(steps)
		alpha := shadow.Color.A * (1.0 / float64(steps))
		blurBounds := shadowBounds.Inset(geom.All(-float64(i) * 1.5))
		r.fillRoundedRect(blurBounds, shadowRadius, shadow.Color.WithAlpha(alpha*factor))
	}
}
