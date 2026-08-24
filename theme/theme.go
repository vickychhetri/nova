package theme

import (
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
)

// Palette contains semantic color tokens for UI components.
//
// Tokens describe roles rather than individual widgets. Components should use
// the role that matches their state so a complete theme can change appearance
// without requiring component-specific color logic.
type Palette struct {
	// Primary colors describe the main action and its interactive states.
	Primary       color.Color
	PrimaryHover  color.Color
	PrimaryActive color.Color
	PrimaryText   color.Color

	// Secondary colors describe supporting controls and their interactive states.
	Secondary       color.Color
	SecondaryHover  color.Color
	SecondaryActive color.Color
	SecondaryText   color.Color

	// Background and surface tokens describe application layers and elevation.
	Background   color.Color
	Surface      color.Color
	SurfaceHover color.Color
	Card         color.Color
	Border       color.Color
	BorderHover  color.Color
	BorderFocus  color.Color

	// Text tokens express hierarchy and availability for foreground content.
	TextPrimary   color.Color
	TextSecondary color.Color
	TextMuted     color.Color
	TextDisabled  color.Color

	// Status colors communicate semantic outcomes and information.
	Success color.Color
	Warning color.Color
	Error   color.Color
	Info    color.Color
}

// Typography contains the font family and nominal text sizes used by the
// design system. Values are expressed in logical pixels; line-height values are
// not currently stored in this structure.
type Typography struct {
	FontFamily string
	H1Size     float64
	H2Size     float64
	H3Size     float64
	H4Size     float64
	BodySize   float64
	SmallSize  float64
	CodeSize   float64
}

// Spacing contains reusable distance tokens for padding, margins, and gaps.
// The names progress from the smallest to the largest default distance.
type Spacing struct {
	XXS float64 // 2
	XS  float64 // 4
	SM  float64 // 8
	MD  float64 // 12
	LG  float64 // 16
	XL  float64 // 24
	XXL float64 // 32
}

// Radii contains reusable corner-radius tokens for surfaces and controls.
// Each value uses geom.CornerRadius so per-corner radii remain possible.
type Radii struct {
	SM   geom.CornerRadius // 4
	MD   geom.CornerRadius // 8
	LG   geom.CornerRadius // 12
	XL   geom.CornerRadius // 16
	Full geom.CornerRadius // 9999
}

// Theme encapsulates a complete design system for one visual mode.
//
// IsDark identifies the intended mode, while Palette, Typography, Spacing, and
// Radii provide the tokens consumed by components. The flag does not itself
// derive or validate any token values.
type Theme struct {
	IsDark     bool
	Palette    Palette
	Typography Typography
	Spacing    Spacing
	Radii      Radii
}

// DefaultSpacing is the standard Nova spacing scale in logical pixels.
var DefaultSpacing = Spacing{
	XXS: 2,
	XS:  4,
	SM:  8,
	MD:  12,
	LG:  16,
	XL:  24,
	XXL: 32,
}

// DefaultRadii is the standard Nova corner-radius scale in logical pixels.
var DefaultRadii = Radii{
	SM:   geom.RadiusUniform(4),
	MD:   geom.RadiusUniform(8),
	LG:   geom.RadiusUniform(12),
	XL:   geom.RadiusUniform(16),
	Full: geom.RadiusUniform(9999),
}

// DefaultTypography is the standard Nova typography scale in logical pixels.
// FontFamily is a CSS-like preference string for consumers that support
// fallback families.
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

// Dark returns a new default dark theme.
//
// The returned Theme owns its value fields, while the default typography,
// spacing, and radii values are copied into the new structure.
func Dark() *Theme {
	return &Theme{
		IsDark: true,
		Palette: Palette{
			Primary:       color.Hex("#3B82F6"),
			PrimaryHover:  color.Hex("#2563EB"),
			PrimaryActive: color.Hex("#1D4ED8"),
			PrimaryText:   color.White,

			Secondary:       color.Hex("#374151"),
			SecondaryHover:  color.Hex("#4B5563"),
			SecondaryActive: color.Hex("#6B7280"),
			SecondaryText:   color.Hex("#F9FAFB"),

			Background:   color.Hex("#0F172A"),
			Surface:      color.Hex("#1E293B"),
			SurfaceHover: color.Hex("#334155"),
			Card:         color.Hex("#1E293B"),
			Border:       color.Hex("#334155"),
			BorderHover:  color.Hex("#475569"),
			BorderFocus:  color.Hex("#60A5FA"),

			TextPrimary:   color.Hex("#F8FAFC"),
			TextSecondary: color.Hex("#CBD5E1"),
			TextMuted:     color.Hex("#64748B"),
			TextDisabled:  color.Hex("#475569"),

			Success: color.Hex("#22C55E"),
			Warning: color.Hex("#EAB308"),
			Error:   color.Hex("#EF4444"),
			Info:    color.Hex("#06B6D4"),
		},
		Typography: DefaultTypography,
		Spacing:    DefaultSpacing,
		Radii:      DefaultRadii,
	}
}

// Light returns a new default light theme.
//
// Light and Dark share the same structural token scales but provide different
// semantic palette values for surfaces, text, controls, and status colors.
func Light() *Theme {
	return &Theme{
		IsDark: false,
		Palette: Palette{
			Primary:       color.Hex("#2563EB"),
			PrimaryHover:  color.Hex("#1D4ED8"),
			PrimaryActive: color.Hex("#1E40AF"),
			PrimaryText:   color.White,

			Secondary:       color.Hex("#FFFFFF"),
			SecondaryHover:  color.Hex("#F1F5F9"),
			SecondaryActive: color.Hex("#E2E8F0"),
			SecondaryText:   color.Hex("#1E293B"),

			Background:   color.Hex("#F8FAFC"),
			Surface:      color.Hex("#FFFFFF"),
			SurfaceHover: color.Hex("#F1F5F9"),
			Card:         color.Hex("#FFFFFF"),
			Border:       color.Hex("#E2E8F0"),
			BorderHover:  color.Hex("#94A3B8"),
			BorderFocus:  color.Hex("#2563EB"),

			TextPrimary:   color.Hex("#0F172A"),
			TextSecondary: color.Hex("#475569"),
			TextMuted:     color.Hex("#64748B"),
			TextDisabled:  color.Hex("#94A3B8"),

			Success: color.Hex("#059669"),
			Warning: color.Hex("#D97706"),
			Error:   color.Hex("#DC2626"),
			Info:    color.Hex("#0284C7"),
		},
		Typography: DefaultTypography,
		Spacing:    DefaultSpacing,
		Radii:      DefaultRadii,
	}
}

// currentTheme is the process-wide theme used by code that calls Current.
// Access is intentionally centralized through Current and SetCurrent.
var currentTheme = Dark()

// Current returns the process-wide active theme.
//
// The returned pointer is the active object itself, not a defensive copy.
// Callers that mutate it affect every consumer of the global theme.
func Current() *Theme {
	return currentTheme
}

// SetCurrent replaces the process-wide active theme when t is non-nil.
// Passing nil is ignored so Current never becomes nil through this setter.
// The global pointer is not synchronized; applications should set the theme
// during initialization or provide external synchronization for concurrent
// reads and writes.
func SetCurrent(t *Theme) {
	if t != nil {
		currentTheme = t
	}
}
