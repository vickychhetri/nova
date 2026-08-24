package color

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
)

// Color represents an RGBA color with normalized float64 components in the
// range 0.0 to 1.0.
//
// Keeping channels normalized provides precision for blending and color
// manipulation. Constructors clamp floating-point inputs where documented;
// methods that operate directly on an existing Color preserve its RGB values
// unless they explicitly modify them.
type Color struct {
	R float64 // Red channel [0..1].
	G float64 // Green channel [0..1].
	B float64 // Blue channel [0..1].
	A float64 // Alpha channel [0..1], where 0 is transparent and 1 opaque.
}

// RGBA creates a Color from 8-bit red, green, blue, and alpha components.
// Each channel is divided by 255 to produce the normalized representation.
func RGBA(r, g, b, a uint8) Color {
	return Color{
		R: float64(r) / 255.0,
		G: float64(g) / 255.0,
		B: float64(b) / 255.0,
		A: float64(a) / 255.0,
	}
}

// RGB creates an opaque Color from 8-bit red, green, and blue components.
func RGB(r, g, b uint8) Color {
	return RGBA(r, g, b, 255)
}

// FloatRGBA creates a Color from normalized floating-point components.
// Values outside [0, 1] are clamped independently for each channel.
func FloatRGBA(r, g, b, a float64) Color {
	return Color{
		R: clamp01(r),
		G: clamp01(g),
		B: clamp01(b),
		A: clamp01(a),
	}
}

// Hex creates a Color from a hexadecimal string.
//
// Supported forms are RGB (#RGB), RGBA (#RGBA), RRGGBB (#RRGGBB), and RRGGBBAA
// (#RRGGBBAA). A leading # is optional. Invalid lengths or digits silently
// leave the affected parsed channels at their current defaults.
func Hex(hex string) Color {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b, a uint64 = 0, 0, 0, 255

	switch len(hex) {
	case 3: // RGB (e.g. "F00")
		r, _ = strconv.ParseUint(string([]byte{hex[0], hex[0]}), 16, 8)
		g, _ = strconv.ParseUint(string([]byte{hex[1], hex[1]}), 16, 8)
		b, _ = strconv.ParseUint(string([]byte{hex[2], hex[2]}), 16, 8)
	case 4: // RGBA (e.g. "F00F")
		r, _ = strconv.ParseUint(string([]byte{hex[0], hex[0]}), 16, 8)
		g, _ = strconv.ParseUint(string([]byte{hex[1], hex[1]}), 16, 8)
		b, _ = strconv.ParseUint(string([]byte{hex[2], hex[2]}), 16, 8)
		a, _ = strconv.ParseUint(string([]byte{hex[3], hex[3]}), 16, 8)
	case 6: // RRGGBB (e.g. "FF0000")
		r, _ = strconv.ParseUint(hex[0:2], 16, 8)
		g, _ = strconv.ParseUint(hex[2:4], 16, 8)
		b, _ = strconv.ParseUint(hex[4:6], 16, 8)
	case 8: // RRGGBBAA (e.g. "FF000088")
		r, _ = strconv.ParseUint(hex[0:2], 16, 8)
		g, _ = strconv.ParseUint(hex[2:4], 16, 8)
		b, _ = strconv.ParseUint(hex[4:6], 16, 8)
		a, _ = strconv.ParseUint(hex[6:8], 16, 8)
	}

	return RGBA(uint8(r), uint8(g), uint8(b), uint8(a))
}

// HSL creates an opaque Color from hue in degrees, saturation in [0, 1], and
// lightness in [0, 1]. Hue wraps around the color wheel.
func HSL(h, s, l float64) Color {
	return HSLA(h, s, l, 1.0)
}

// HSLA creates a Color from HSL components and normalized alpha.
//
// Hue is normalized into [0, 360), while saturation, lightness, and alpha are
// clamped to [0, 1]. The conversion uses the standard chroma/intermediate/match
// formulation and returns normalized RGB channels.
func HSLA(h, s, l, a float64) Color {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	s = clamp01(s)
	l = clamp01(l)
	a = clamp01(a)

	c := (1.0 - math.Abs(2.0*l-1.0)) * s
	x := c * (1.0 - math.Abs(math.Mod(h/60.0, 2.0)-1.0))
	m := l - c/2.0

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return Color{
		R: r + m,
		G: g + m,
		B: b + m,
		A: a,
	}
}

// NRGBA converts the normalized color to Go's 8-bit non-premultiplied NRGBA.
// Channels are rounded to the nearest byte.
func (c Color) NRGBA() color.NRGBA {
	return color.NRGBA{
		R: uint8(math.Round(c.R * 255)),
		G: uint8(math.Round(c.G * 255)),
		B: uint8(math.Round(c.B * 255)),
		A: uint8(math.Round(c.A * 255)),
	}
}

// RGBAUint32 packs rounded channels into a 0xRRGGBBAA uint32 value.
func (c Color) RGBAUint32() uint32 {
	r := uint32(math.Round(c.R * 255))
	g := uint32(math.Round(c.G * 255))
	b := uint32(math.Round(c.B * 255))
	a := uint32(math.Round(c.A * 255))
	return (r << 24) | (g << 16) | (b << 8) | a
}

// WithAlpha returns a copy with alpha replaced by the clamped value.
// The RGB channels are unchanged.
func (c Color) WithAlpha(alpha float64) Color {
	return Color{R: c.R, G: c.G, B: c.B, A: clamp01(alpha)}
}

// MultiplyAlpha returns a copy whose alpha is multiplied by factor and then
// clamped. RGB channels are unchanged, which is useful for glyph coverage and
// translucent effects.
func (c Color) MultiplyAlpha(factor float64) Color {
	return Color{R: c.R, G: c.G, B: c.B, A: clamp01(c.A * factor)}
}

// Lighten moves each RGB channel toward white by amount and preserves alpha.
// Amount is expected in [0, 1]; values outside that range are clamped through
// the resulting channels.
func (c Color) Lighten(amount float64) Color {
	return Color{
		R: clamp01(c.R + (1.0-c.R)*amount),
		G: clamp01(c.G + (1.0-c.G)*amount),
		B: clamp01(c.B + (1.0-c.B)*amount),
		A: c.A,
	}
}

// Darken scales each RGB channel toward black by amount and preserves alpha.
func (c Color) Darken(amount float64) Color {
	return Color{
		R: clamp01(c.R * (1.0 - amount)),
		G: clamp01(c.G * (1.0 - amount)),
		B: clamp01(c.B * (1.0 - amount)),
		A: c.A,
	}
}

// Lerp linearly interpolates all four channels toward other by t. The factor
// is clamped to [0, 1], so t=0 returns c and t=1 returns other.
func (c Color) Lerp(other Color, t float64) Color {
	t = clamp01(t)
	return Color{
		R: c.R + (other.R-c.R)*t,
		G: c.G + (other.G-c.G)*t,
		B: c.B + (other.B-c.B)*t,
		A: c.A + (other.A-c.A)*t,
	}
}

// Luminance returns relative sRGB luminance according to WCAG 2.1.
// Alpha is ignored; luminance describes the RGB color itself.
func (c Color) Luminance() float64 {
	r := srgbToLinear(c.R)
	g := srgbToLinear(c.G)
	b := srgbToLinear(c.B)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// ContrastRatio calculates the WCAG contrast ratio against another RGB color.
// The theoretical result ranges from 1 to 21, and alpha is ignored.
func (c Color) ContrastRatio(other Color) float64 {
	l1 := c.Luminance()
	l2 := other.Luminance()
	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)
	return (lighter + 0.05) / (darker + 0.05)
}

// HexString returns an uppercase hexadecimal representation.
// Opaque colors use #RRGGBB; colors with alpha below approximately 1 use
// #RRGGBBAA. Channels are rounded to bytes before formatting.
func (c Color) HexString() string {
	r := int(math.Round(c.R * 255))
	g := int(math.Round(c.G * 255))
	b := int(math.Round(c.B * 255))
	if c.A < 0.999 {
		a := int(math.Round(c.A * 255))
		return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
	}
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func (c Color) String() string {
	// String follows the compact hex representation used by HexString.
	return c.HexString()
}

// clamp01 restricts a normalized channel or factor to the closed interval
// [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// srgbToLinear converts an sRGB channel into linear-light space for luminance
// calculations. The piecewise transfer function is defined by WCAG/sRGB.
func srgbToLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// Common color constants used throughout Nova's themes and examples.
var (
	Transparent = FloatRGBA(0, 0, 0, 0)
	Black       = RGB(0, 0, 0)
	White       = RGB(255, 255, 255)
	Red         = RGB(239, 68, 68)
	Green       = RGB(34, 197, 94)
	Blue        = RGB(59, 130, 246)
	Yellow      = RGB(234, 179, 8)
	Cyan        = RGB(6, 182, 212)
	Magenta     = RGB(217, 70, 239)
	Gray50      = Hex("#F9FAFB")
	Gray100     = Hex("#F3F4F6")
	Gray200     = Hex("#E5E7EB")
	Gray300     = Hex("#D1D5DB")
	Gray400     = Hex("#9CA3AF")
	Gray500     = Hex("#6B7280")
	Gray600     = Hex("#4B5563")
	Gray700     = Hex("#374151")
	Gray800     = Hex("#1F2937")
	Gray900     = Hex("#111827")
	Gray950     = Hex("#030712")
)
