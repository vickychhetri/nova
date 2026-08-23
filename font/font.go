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

// FontWeight represents standard font weight constants.
const (
	WeightLight    = 300
	WeightRegular  = 400
	WeightMedium   = 500
	WeightSemiBold = 600
	WeightBold     = 700
)

// GlyphMetrics describes dimensions and offsets of a rasterized character.
type GlyphMetrics struct {
	Rune     rune
	Width    float64
	Height   float64
	AdvanceX float64
	BearingX float64
	BearingY float64
	Bitmap   *image.Alpha
}

// Global parsed font families
var (
	parsedRegular *sfnt.Font
	parsedMedium  *sfnt.Font
	parsedBold    *sfnt.Font
	parsedMono    *sfnt.Font
	fontsInitOnce sync.Once
)

func initFonts() {
	fontsInitOnce.Do(func() {
		parsedRegular, _ = opentype.Parse(goregular.TTF)
		parsedMedium, _ = opentype.Parse(gomedium.TTF)
		parsedBold, _ = opentype.Parse(gobold.TTF)
		parsedMono, _ = opentype.Parse(gomono.TTF)
	})
}

// Face cache to avoid re-creating OpenType faces
type faceKey struct {
	weight int
	size   int
}

var (
	faceCacheMu sync.RWMutex
	faceCache   = make(map[faceKey]font.Face)
)

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
		// Fallback to regular
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

// Glyph Cache
type glyphKey struct {
	r      rune
	size   int
	weight int
}

var (
	glyphCacheMu sync.RWMutex
	glyphCache   = make(map[glyphKey]GlyphMetrics)
)

// GetGlyph returns glyph metrics and smooth vector rasterized alpha bitmap for rune r.
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

// MeasureCharWidth calculates exact proportional character advance using OpenType face.
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
	face := getFace(fontSize, weight)
	metrics := face.Metrics()
	ascent := float64(metrics.Ascent) / 64.0
	descent := float64(metrics.Descent) / 64.0

	bounds, adv, ok := face.GlyphBounds(r)
	if !ok {
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

	dstRGBA := image.NewRGBA(image.Rect(0, 0, totalW, totalH))
	drawer := &font.Drawer{
		Dst:  dstRGBA,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(int(ascent))},
	}
	drawer.DrawString(string(r))

	alphaImg := image.NewAlpha(image.Rect(0, 0, totalW, totalH))
	for y := 0; y < totalH; y++ {
		for x := 0; x < totalW; x++ {
			c := dstRGBA.RGBAAt(x, y)
			alphaImg.SetAlpha(x, y, color.Alpha{A: c.A})
		}
	}

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
