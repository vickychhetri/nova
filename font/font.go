package font

import (
	"image"
	"image/color"
	"math"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// FontWeight values represent standard numeric font weights used when choosing
// an embedded font face. The constants follow common CSS/OpenType conventions.
const (
	WeightLight    = 300
	WeightRegular  = 400
	WeightMedium   = 500
	WeightSemiBold = 600
	WeightBold     = 700
)

// GlyphMetrics describes the layout metrics and alpha bitmap of one rasterized
// character.
//
// Width and Height describe the returned bitmap area. AdvanceX is the distance
// to move the text pen for the next glyph, while BearingX and BearingY describe
// the glyph's placement relative to its origin and baseline.
type GlyphMetrics struct {
	// Rune is the character represented by this metric record.
	Rune rune
	// Width and Height are the bitmap dimensions in logical pixels.
	Width  float64
	Height float64
	// AdvanceX is the horizontal layout advance supplied by OpenType.
	AdvanceX float64
	// BearingX and BearingY are the glyph's baseline-relative offsets.
	BearingX float64
	BearingY float64
	// Bitmap contains the glyph coverage as an alpha-only image.
	Bitmap *image.Alpha
}

// Embedded font families are parsed lazily by initFonts and reused by all face
// and glyph requests. The sfnt.Font values are immutable after parsing.
var (
	parsedRegular *sfnt.Font
	parsedMedium  *sfnt.Font
	parsedBold    *sfnt.Font
	parsedMono    *sfnt.Font
	fontsInitOnce sync.Once
)

func initFonts() {
	fontsInitOnce.Do(func() {
		// Parse each embedded TTF exactly once. Parse errors are ignored here
		// because the bundled assets are compile-time resources; getFace still
		// provides a fallback path if a face cannot be created.
		parsedRegular, _ = opentype.Parse(goregular.TTF)
		parsedMedium, _ = opentype.Parse(gomedium.TTF)
		parsedBold, _ = opentype.Parse(gobold.TTF)
		parsedMono, _ = opentype.Parse(gomono.TTF)
	})
}

// faceKey identifies a cached face by the normalized weight and integer point
// size used to create it.
type faceKey struct {
	weight int
	size   int
}

var (
	// faceCacheMu protects faceCache because faces may be requested by multiple
	// rendering or layout goroutines.
	faceCacheMu sync.RWMutex
	// faceCache avoids repeatedly constructing OpenType face instances.
	faceCache = make(map[faceKey]font.Face)
)

// getFace returns a cached OpenType face for the requested size and weight.
// Font sizes are rounded to integer values and clamped to a minimum of 8 so
// cache keys remain bounded and tiny faces remain usable.
func getFace(fontSize float64, weight int) font.Face {
	initFonts()

	sizeInt := int(math.Round(fontSize))
	if sizeInt < 8 {
		sizeInt = 8
	}

	key := faceKey{weight: weight, size: sizeInt}

	faceCacheMu.RLock()
	if f, ok := faceCache[key]; ok {
		faceCacheMu.RUnlock()
		return f
	}
	faceCacheMu.RUnlock()

	// Weight selection currently distinguishes regular, medium, and bold
	// families. Intermediate values use the closest supported lower tier.
	sf := parsedRegular
	if weight >= WeightBold && parsedBold != nil {
		sf = parsedBold
	} else if weight >= WeightMedium && parsedMedium != nil {
		sf = parsedMedium
	}

	face, err := opentype.NewFace(sf, &opentype.FaceOptions{
		Size:    float64(sizeInt),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		// Fall back to the regular family if the selected face cannot be built.
		face, _ = opentype.NewFace(parsedRegular, &opentype.FaceOptions{
			Size:    float64(sizeInt),
			DPI:     72,
			Hinting: font.HintingFull,
		})
	}

	faceCacheMu.Lock()
	faceCache[key] = face
	faceCacheMu.Unlock()

	return face
}

// glyphKey identifies a cached rasterized glyph by rune, normalized size, and
// requested weight.
type glyphKey struct {
	r      rune
	size   int
	weight int
}

var (
	// glyphCacheMu protects glyphCache during concurrent glyph lookup and
	// insertion.
	glyphCacheMu sync.RWMutex
	// glyphCache stores complete metric and bitmap results for reuse.
	glyphCache = make(map[glyphKey]GlyphMetrics)
)

// GetGlyph returns metrics and an alpha bitmap for rune r at fontSize and
// weight. Results are cached after the first rasterization for a given rune,
// normalized size, and weight combination.
func GetGlyph(r rune, fontSize float64, weight int) GlyphMetrics {
	sizeInt := int(math.Round(fontSize))
	if sizeInt < 8 {
		sizeInt = 8
	}

	key := glyphKey{r: r, size: sizeInt, weight: weight}

	glyphCacheMu.RLock()
	if gm, ok := glyphCache[key]; ok {
		glyphCacheMu.RUnlock()
		return gm
	}
	glyphCacheMu.RUnlock()

	gm := rasterizeVectorGlyph(r, float64(sizeInt), weight)

	glyphCacheMu.Lock()
	glyphCache[key] = gm
	glyphCacheMu.Unlock()

	return gm
}

// MeasureCharWidth returns the proportional horizontal advance for r using the
// selected OpenType face. The result is expressed in logical pixels rather than
// the font package's internal 26.6 fixed-point units.
func MeasureCharWidth(r rune, fontSize float64, weight int) float64 {
	face := getFace(fontSize, weight)
	if face == nil {
		return fontSize * 0.55
	}
	adv, ok := face.GlyphAdvance(r)
	if !ok {
		return fontSize * 0.55
	}
	return float64(adv) / 64.0
}

func rasterizeVectorGlyph(r rune, fontSize float64, weight int) GlyphMetrics {
	// OpenType metrics use 26.6 fixed-point values. Convert them to floating
	// point logical pixels before calculating bitmap dimensions and bearings.
	face := getFace(fontSize, weight)
	metrics := face.Metrics()
	ascent := float64(metrics.Ascent) / 64.0
	descent := float64(metrics.Descent) / 64.0

	bounds, adv, ok := face.GlyphBounds(r)
	if !ok {
		// Unknown glyphs still receive a predictable advance so text layout can
		// continue without a missing-glyph error.
		adv = fixed.I(int(fontSize * 0.55))
	}

	advanceX := float64(adv) / 64.0
	minX := float64(bounds.Min.X) / 64.0
	minY := float64(bounds.Min.Y) / 64.0
	maxX := float64(bounds.Max.X) / 64.0
	maxY := float64(bounds.Max.Y) / 64.0

	glyphW := int(math.Ceil(maxX - minX))
	glyphH := int(math.Ceil(maxY - minY))

	totalH := int(math.Ceil(ascent + descent))
	if totalH < 1 {
		totalH = int(math.Ceil(fontSize * 1.2))
	}
	totalW := int(math.Ceil(math.Max(advanceX, maxX)))
	if totalW < 1 {
		totalW = 1
	}

	// Draw into a temporary white RGBA image, then retain only alpha. This
	// produces a reusable coverage bitmap that the rasterizer can tint later.
	dstRGBA := image.NewRGBA(image.Rect(0, 0, totalW, totalH))
	drawer := &font.Drawer{
		Dst:  dstRGBA,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: face,
		// Position the glyph at the font ascent so the bitmap follows the
		// standard baseline convention used by x/image/font.
		Dot: fixed.Point26_6{X: fixed.I(0), Y: fixed.I(int(ascent))},
	}
	drawer.DrawString(string(r))

	alphaImg := image.NewAlpha(image.Rect(0, 0, totalW, totalH))
	for y := 0; y < totalH; y++ {
		for x := 0; x < totalW; x++ {
			c := dstRGBA.RGBAAt(x, y)
			alphaImg.SetAlpha(x, y, color.Alpha{A: c.A})
		}
	}

	// These values are retained from the glyph-bound calculation for future
	// bearing-aware bitmap placement; Width and Height currently use the full
	// font line box to keep cached glyph images consistently sized.
	_ = glyphW
	_ = glyphH
	_ = minY

	return GlyphMetrics{
		Rune:     r,
		Width:    float64(totalW),
		Height:   float64(totalH),
		AdvanceX: advanceX,
		BearingX: minX,
		BearingY: ascent,
		Bitmap:   alphaImg,
	}
}
