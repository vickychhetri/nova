package theme

import (
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
)

// Palette contains color tokens for UI components.
type Palette struct {
	Primary         color.Color
	PrimaryHover    color.Color
	PrimaryActive   color.Color
	PrimaryText     color.Color

	Secondary       color.Color
	SecondaryHover  color.Color
	SecondaryActive color.Color
	SecondaryText   color.Color

	Background      color.Color
	Surface         color.Color
	SurfaceHover    color.Color
	Card            color.Color
	Border          color.Color
	BorderHover     color.Color
	BorderFocus     color.Color

	TextPrimary     color.Color
	TextSecondary   color.Color
	TextMuted       color.Color
	TextDisabled    color.Color

	Success         color.Color
	Warning         color.Color
	Error           color.Color
	Info            color.Color
}

// Typography contains font sizes and line heights.
type Typography struct {
	FontFamily  string
	H1Size      float64
	H2Size      float64
	H3Size      float64
	H4Size      float64
	BodySize    float64
	SmallSize   float64
	CodeSize    float64
}

// Spacing tokens.
type Spacing struct {
	XXS float64 // 2
	XS  float64 // 4
	SM  float64 // 8
	MD  float64 // 12
	LG  float64 // 16
	XL  float64 // 24
	XXL float64 // 32
}

// Radii tokens for rounded corners.
type Radii struct {
	SM   geom.CornerRadius // 4
	MD   geom.CornerRadius // 8
	LG   geom.CornerRadius // 12
	XL   geom.CornerRadius // 16
	Full geom.CornerRadius // 9999
}

// Theme encapsulates a complete design system.
type Theme struct {
	IsDark     bool
	Palette    Palette
	Typography Typography
	Spacing    Spacing
	Radii      Radii
}

// Default spacing scale
var DefaultSpacing = Spacing{
	XXS: 2,
	XS:  4,
	SM:  8,
	MD:  12,
	LG:  16,
	XL:  24,
	XXL: 32,
}

// Default radii scale
var DefaultRadii = Radii{
	SM:   geom.RadiusUniform(4),
	MD:   geom.RadiusUniform(8),
	LG:   geom.RadiusUniform(12),
	XL:   geom.RadiusUniform(16),
	Full: geom.RadiusUniform(9999),
}

// Default typography scale
var DefaultTypography = Typography{
	FontFamily: "Inter, system-ui, sans-serif",
	H1Size:     32,
	H2Size:     24,
	H3Size:     20,
	H4Size:     18,
	BodySize:   14,
	SmallSize:  12,
	CodeSize:   13,
}

// Dark returns default dark theme.
func Dark() *Theme {
	return &Theme{
		IsDark: true,
		Palette: Palette{
			Primary:         color.Hex("#3B82F6"),
			PrimaryHover:    color.Hex("#2563EB"),
			PrimaryActive:   color.Hex("#1D4ED8"),
			PrimaryText:     color.White,

			Secondary:       color.Hex("#374151"),
			SecondaryHover:  color.Hex("#4B5563"),
			SecondaryActive: color.Hex("#6B7280"),
			SecondaryText:   color.Hex("#F9FAFB"),

			Background:      color.Hex("#0F172A"),
			Surface:         color.Hex("#1E293B"),
			SurfaceHover:    color.Hex("#334155"),
			Card:            color.Hex("#1E293B"),
			Border:          color.Hex("#334155"),
			BorderHover:     color.Hex("#475569"),
			BorderFocus:     color.Hex("#60A5FA"),

			TextPrimary:     color.Hex("#F8FAFC"),
			TextSecondary:   color.Hex("#CBD5E1"),
			TextMuted:       color.Hex("#64748B"),
			TextDisabled:    color.Hex("#475569"),

			Success:         color.Hex("#22C55E"),
			Warning:         color.Hex("#EAB308"),
			Error:           color.Hex("#EF4444"),
			Info:            color.Hex("#06B6D4"),
		},
		Typography: DefaultTypography,
		Spacing:    DefaultSpacing,
		Radii:      DefaultRadii,
	}
}

// Light returns default light theme.
func Light() *Theme {
	return &Theme{
		IsDark: false,
		Palette: Palette{
			Primary:         color.Hex("#2563EB"),
			PrimaryHover:    color.Hex("#1D4ED8"),
			PrimaryActive:   color.Hex("#1E40AF"),
			PrimaryText:     color.White,

			Secondary:       color.Hex("#F1F5F9"),
			SecondaryHover:  color.Hex("#E2E8F0"),
			SecondaryActive: color.Hex("#CBD5E1"),
			SecondaryText:   color.Hex("#0F172A"),

			Background:      color.Hex("#F8FAFC"),
			Surface:         color.Hex("#FFFFFF"),
			SurfaceHover:    color.Hex("#F1F5F9"),
			Card:            color.Hex("#FFFFFF"),
			Border:          color.Hex("#E2E8F0"),
			BorderHover:     color.Hex("#CBD5E1"),
			BorderFocus:     color.Hex("#3B82F6"),

			TextPrimary:     color.Hex("#0F172A"),
			TextSecondary:   color.Hex("#475569"),
			TextMuted:       color.Hex("#94A3B8"),
			TextDisabled:    color.Hex("#CBD5E1"),

			Success:         color.Hex("#16A34A"),
			Warning:         color.Hex("#CA8A04"),
			Error:           color.Hex("#DC2626"),
			Info:            color.Hex("#0891B2"),
		},
		Typography: DefaultTypography,
		Spacing:    DefaultSpacing,
		Radii:      DefaultRadii,
	}
}

var currentTheme = Dark()

// Current returns the active global theme.
func Current() *Theme {
	return currentTheme
}

// SetCurrent sets the active global theme.
func SetCurrent(t *Theme) {
	if t != nil {
		currentTheme = t
	}
}
