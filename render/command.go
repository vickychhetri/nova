package render

import (
	"image"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
)

// CommandType identifies the operation described by a Command.
//
// A Command uses one shared structure for all rendering operations. The Type
// value tells the renderer which fields are meaningful for that command; fields
// unrelated to the selected type retain their zero values.
type CommandType int

const (
	// CmdNone represents an empty or unspecified command.
	CmdNone CommandType = iota
	// CmdFillRect fills an axis-aligned rectangle with Color.
	CmdFillRect
	// CmdStrokeRect draws a rectangle outline using StrokeColor and StrokeWidth.
	CmdStrokeRect
	// CmdFillRoundedRect fills a rounded rectangle using Bounds and Radius.
	CmdFillRoundedRect
	// CmdStrokeRoundedRect draws a rounded rectangle outline.
	CmdStrokeRoundedRect
	// CmdFillCircle fills a circle centered at P1 with Radius.
	CmdFillCircle
	// CmdStrokeCircle draws a circle outline centered at P1.
	CmdStrokeCircle
	// CmdLine draws a line between P1 and P2.
	CmdLine
	// CmdText draws Text at P1 using the font properties in the command.
	CmdText
	// CmdImage draws Image inside Bounds.
	CmdImage
	// CmdShadow describes a shadow for the shape in Bounds.
	CmdShadow
	// CmdPushClip begins a rectangular clipping scope using Bounds.
	CmdPushClip
	// CmdPopClip ends the most recently pushed clipping scope.
	CmdPopClip
	// CmdPushTransform begins a transform scope using Transform.
	CmdPushTransform
	// CmdPopTransform ends the most recently pushed transform scope.
	CmdPopTransform
)

// ShadowParams defines the appearance and placement of a drop or box shadow.
//
// OffsetX and OffsetY move the shadow relative to the source shape. Blur
// controls softness, and Spread expands or contracts the shadow geometry. The
// renderer is responsible for converting these parameters into pixels.
type ShadowParams struct {
	// Color is the shadow's compositing color, including its alpha value.
	Color color.Color
	// OffsetX and OffsetY are the shadow displacement from the source bounds.
	OffsetX float64
	OffsetY float64
	// Blur controls the amount of edge softening applied by the renderer.
	Blur float64
	// Spread changes the shadow shape before blur is applied.
	Spread float64
}

// Command represents one self-contained drawing instruction for a renderer
// backend.
//
// The structure is intentionally a tagged union implemented with ordinary Go
// fields: Type is the tag, and the fields relevant to that type form its
// payload. For example, CmdLine uses P1, P2, StrokeColor, and StrokeWidth,
// while CmdText uses P1, Text, FontSize, FontWeight, and Color.
//
// Commands are normally created by Canvas and consumed in sequence by a
// renderer. Command itself does not validate geometry, execute drawing, or
// enforce clipping and transform scopes.
type Command struct {
	// Type selects the renderer operation and determines which other fields
	// are meaningful.
	Type CommandType
	// Bounds stores rectangle geometry for rectangle, image, shadow, and clip
	// commands.
	Bounds geom.Rect
	// Color is used for fills and text.
	Color color.Color
	// StrokeColor is used by outline and line commands.
	StrokeColor color.Color
	// StrokeWidth is the requested outline or line thickness.
	StrokeWidth float64
	// Radius stores corner-radius data or the uniform radius of a circle.
	Radius geom.CornerRadius
	// P1 and P2 store operation-specific points, such as a circle center,
	// text origin, or line endpoints.
	P1, P2 geom.Point
	// Text is the string rendered by CmdText.
	Text string
	// FontSize is the requested text size for CmdText.
	FontSize float64
	// FontWeight uses the conventional numeric weights: 400 is regular and
	// 700 is bold.
	FontWeight int
	// Image is the source image used by CmdImage. The image is referenced, not
	// deep-copied, when a command is created.
	Image image.Image
	// Shadow contains the drop-shadow parameters for CmdShadow.
	Shadow ShadowParams
	// Transform contains the matrix used by transform-scope commands.
	Transform geom.Matrix2D
}
